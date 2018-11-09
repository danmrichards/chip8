package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
)

type opcodeHandler struct {
	opcode  string
	handler opcodeHandlerFunc
}

// opcodeHandlerFunc returns the handled value and an error if it occurred.
// It is possible for the handler to derive a shifted opcode during processing
// from the original opcode set on the chip8 itself.
type opcodeHandlerFunc func() (uint16, error)

// registerHandlers registers up the map of opcode handlers
func (c *chip8) registerHandlers() {
	// TODO: names for opcodes.
	c.handlers = map[uint16]opcodeHandler{
		0x0000: {
			opcode:  "handle0x0000",
			handler: c.handle0x0000,
		},
		0x1000: {
			opcode:  "1NNN",
			handler: c.jump,
		},
		0x2000: {
			opcode:  "2NNN",
			handler: c.callSub,
		},
		0x3000: {
			opcode:  "3XNN",
			handler: c.skipVxNN,
		},
		0x4000: {
			opcode:  "4XNN",
			handler: c.skipVxNotNN,
		},
		0x6000: {
			opcode:  "6XNN",
			handler: c.setVx,
		},
		0x7000: {
			opcode:  "7XNN",
			handler: c.incVx,
		},
		0x8000: {
			opcode:  "handle0x8000",
			handler: c.handle0x8000,
		},
		0x9000: {
			opcode:  "9XY0",
			handler: c.skipVxNotVy,
		},
		0xA000: {
			opcode:  "ANNN",
			handler: c.setAddress,
		},
		0xC000: {
			opcode:  "CXNN",
			handler: c.setVxRand,
		},
		0xD000: {
			opcode:  "DXYN",
			handler: c.drawDisp,
		},
		0xE000: {
			opcode:  "handle0xE000",
			handler: c.handle0xE000,
		},
		0xF000: {
			opcode:  "handle0xF000",
			handler: c.handle0xF000,
		},
	}
}

// handle attempts to handle the current opcode.
//
// If there is no registered handler for opc then an error is returned. The
// handler can also return an error.
func (c *chip8) handle() error {
	// Note that we are not using c.opc to determine the handler here because
	// in the majority of cases the first 4 bits of the opcode are what
	// indicates the action to take.
	h, ok := c.handlers[c.opc&0xF000]
	if !ok {
		return fmt.Errorf("unsupported opcode: 0x%X", c.opc)
	}

	// Handle the opcode.
	v, err := h.handler()
	if err != nil {
		return fmt.Errorf("error handling opcode: %s value: 0x%X: %s", h.opcode, v, err)
	}

	// TODO: debug mode support.
	log.Printf("opcode: %s value: 0x%X\n", h.opcode, v)

	return nil
}

// handle0x0000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (c *chip8) handle0x0000() (uint16, error) {
	switch {
	case c.opc&0x00FF == 0x00E0:
		return c.clrDisp()

	case c.opc&0x00FF == 0x00EE:
		return c.subRet()

	case c.opc&0xF000 == 0x0000:
		return c.callSys()

	default:
		return c.opc, errors.New("unknown opcode [0x0000]")
	}
}

// clrDisp clears the display.
func (c *chip8) clrDisp() (uint16, error) {
	return c.opc & 0x00FF, errors.New("TODO: clrDisp")
}

// subRet returns from a subroutine.
func (c *chip8) subRet() (uint16, error) {
	// Return to the program counter stored in the stack (adding 2 for the
	// next instruction as usual).
	c.pc = c.stack[c.sp] + 2
	c.sp--

	return c.opc & 0x00FF, nil
}

// callSys calls RCA 1802 program at address NNN. Not necessary for most ROMs.
func (c *chip8) callSys() (uint16, error) {
	return c.opc & 0xF000, errors.New("TODO: callSys")
}

// jump jumps to address NNN.
func (c *chip8) jump() (uint16, error) {
	// Jump to NNN.
	c.pc = c.opc & 0x0FFF

	return c.opc, nil
}

// callSub calls subroutine at NNN.
func (c *chip8) callSub() (uint16, error) {
	// Store the current program counter temporarily while we jump to
	// the subroutine. Incrementing the stack pointer to prevent overwrite.
	c.sp++
	c.stack[c.sp] = c.pc

	// Jump to NNN.
	c.pc = c.opc & 0x0FFF

	return c.opc, nil
}

// skipVxNN skips the next instruction if VX equals NN. Usually the next instruction
// is a jump to skip a code block.
func (c *chip8) skipVxNN() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(c.opc & 0x00FF) // Get the last 2 chars.

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if c.V[x] == nn {
		c.pc += 4
	} else {
		c.pc += 2
	}

	return c.opc, nil
}

// skipVxNN skips the next instruction if VX does not equal NN. Usually the next
// instruction is a jump to skip a code block.
func (c *chip8) skipVxNotNN() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(c.opc & 0x00FF) // Get the last 2 chars.

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if c.V[x] != nn {
		c.pc += 4
	} else {
		c.pc += 2
	}

	return c.opc, nil
}

// setVx sets CPU register Vx (where x is A-E) to NN.
func (c *chip8) setVx() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(c.opc & 0x00FF) // Get the last 2 chars.

	c.V[x] = nn
	c.pc += 2

	return c.opc, nil
}

// incVx adds NN to CPU register Vx.
func (c *chip8) incVx() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(c.opc & 0x00FF) // Get the last 2 chars.

	c.V[x] += nn
	c.pc += 2

	return c.opc, nil
}

// handle0x8000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (c *chip8) handle0x8000() (uint16, error) {
	switch c.opc & 0x000F {
	case 0x0000:
		return c.setVxVy()

	case 0x0002:
		return c.setVxAndVy()

	case 0x0004:
		return c.incVxVy()

	case 0x0005:
		return c.decVxVy()

	default:
		return c.opc & 0xFFFF, errors.New("TODO: handle0x8000")
	}
}

// setVxVy sets VX to the value of VY.
func (c *chip8) setVxVy() (uint16, error) {
	c.V[(c.opc&0x0F00)>>8] = c.V[(c.opc&0x00F0)>>4]
	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// setVxAndVy sets VX to VX & VY (bitwise AND operation).
func (c *chip8) setVxAndVy() (uint16, error) {
	c.V[(c.opc&0x0F00)>>8] &= c.V[(c.opc&0x00F0)>>4]
	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// incVxVy adds VY to VX. VF is set to 1 when there's a carry, and to 0 when
// there isn't.
func (c *chip8) incVxVy() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8
	y := (c.opc & 0x00F0) >> 4

	if c.V[y] > (0xFF - c.V[x]) {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
	c.V[x] += c.V[y]

	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// decVxVy VY is subtracted from VX. VF is set to 0 when there's a borrow, and 1
// when there isn't.
func (c *chip8) decVxVy() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8
	y := (c.opc & 0x00F0) >> 4

	if c.V[y] > c.V[x] {
		c.V[0xF] = 0
	} else {
		c.V[0xF] = 1
	}
	c.V[x] -= c.V[y]

	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// skipVxNotVy skips the next instruction if VX doesn't equal VY. Usually the
// next instruction is a jump to skip a code block.
func (c *chip8) skipVxNotVy() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8
	y := (c.opc & 0x00F0) >> 4

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if c.V[x] != c.V[y] {
		c.pc += 4
	} else {
		c.pc += 2
	}

	return c.opc, nil
}

// setAddress sets the index register to the address NNN.
func (c *chip8) setAddress() (uint16, error) {
	c.I = c.opc & 0x0FFF
	c.pc += 2 // Increase by 2 because each instruction is 2 bytes long.

	return c.opc, nil
}

// setVxRand sets VX to the result of a bitwise and operation on a random
// number (typically: 0 to 255) and NN.
func (c *chip8) setVxRand() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(c.opc & 0x00FF) // Get the last 2 chars.

	// TODO: better rand.
	c.V[x] = byte(rand.Intn(256)) & nn
	c.pc += 2

	return c.opc, nil
}

// drawDisp draws a sprite at coordinate (VX, VY) that has a width of 8 pixels
// and a height of N pixels. Each row of 8 pixels is read as bit-coded starting
// from memory location I; I doesn't change after the execution of this
// instruction. VF is set to 1 if any screen pixels are flipped from set to
// unset when the sprite is drawn, and to 0 if that doesn't happen.
func (c *chip8) drawDisp() (uint16, error) {
	x := uint16(c.V[(c.opc&0x0F00)>>8])
	y := uint16(c.V[(c.opc&0x00F0)>>4])
	height := c.opc & 0x000F
	var pixel uint16

	c.V[0xF] = 0
	for cY := uint16(0); cY < height; cY++ {
		pixel = uint16(c.memory[c.I+cY])
		for cX := uint16(0); cX < 8; cX++ {
			index := x + cX + ((y + cY) * 64)
			if index > uint16(len(c.disp)) {
				continue
			}
			if (pixel & (0x80 >> cX)) != 0 {
				if c.disp[index] == 1 {
					c.V[0xF] = 1
				}
				c.disp[index] ^= 1
			}
		}
	}

	c.draw = true
	c.pc += 2

	return c.opc, nil
}

// handle0xF000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (c *chip8) handle0xE000() (uint16, error) {
	switch c.opc & 0x00FF {
	case 0x00A1:
		return c.skipVxKeyNotPressed()
	default:
		return c.opc & 0xFFFF, errors.New("TODO: handle0xF000")
	}
}

// skipVxKeyNotPressed the next instruction if the key stored in VX isn't
// pressed (usually the next instruction is a jump to skip a code block).
func (c *chip8) skipVxKeyNotPressed() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if c.key[c.V[x]] == 0 {
		c.pc += 4
	} else {
		c.pc += 2
	}

	return c.opc & 0xFFFF, nil
}

// handle0xF000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (c *chip8) handle0xF000() (uint16, error) {
	switch c.opc & 0x00FF {
	case 0x0007:
		return c.getDelayTimer()

	case 0x0015:
		return c.setDelayTimer()

	case 0x0018:
		return c.setSoundTimer()

	case 0x0029:
		return c.loadFont()

	case 0x0033:
		return c.setBCD()

	case 0x0065:
		return c.regLoad()

	default:
		return c.opc & 0xFFFF, errors.New("TODO: handle0xF000")
	}
}

// getDelayTimer the delay timer to VX.
func (c *chip8) getDelayTimer() (uint16, error) {
	c.V[(c.opc&0x0F00)>>8] = c.delayTimer
	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// setDelayTimer sets the delay timer to VX.
func (c *chip8) setDelayTimer() (uint16, error) {
	c.delayTimer = c.V[(c.opc&0x0F00)>>8]
	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// setDelayTimer sets the sound timer to VX.
func (c *chip8) setSoundTimer() (uint16, error) {
	c.soundTimer = c.V[(c.opc&0x0F00)>>8]
	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// loadFont sets I to the location of the sprite for the character in VX.
// Characters 0-F (in hexadecimal) are represented by a 4x5 font.
func (c *chip8) loadFont() (uint16, error) {
	c.I = uint16(c.V[(c.opc&0x0F00)>>8]) * 5
	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// setBCD stores the binary code decimal representation of VX with the most
// significant of three digits at the address in I, the middle digit at I plus
// 1, and the least significant digit at I plus 2. (In other words, take the
// decimal representation of VX, place the hundreds digit in memory at location
// in I, the tens digit at location I+1, and the ones digit at location I+2).1
func (c *chip8) setBCD() (uint16, error) {
	x := (c.opc & 0x0F00) >> 8

	c.memory[c.I] = c.V[x] / 100          // Hundreds.
	c.memory[c.I+1] = (c.V[x] / 10) % 10  // Tens.
	c.memory[c.I+2] = (c.V[x] % 100) % 10 // Ones.
	c.pc += 2

	return c.opc & 0xFFFF, nil
}

// fillV0Vx stores V0 to VX (including VX) in memory starting at address I. The
// offset from I is increased by 1 for each value written, but I itself is left
// unmodified.
func (c *chip8) regLoad() (uint16, error) {
	for i := uint16(0); i <= (c.opc&0x0F00)>>8; i++ {
		c.V[i] = c.memory[c.I+i]
	}
	c.pc += 2

	return c.opc & 0xFFFF, nil
}
