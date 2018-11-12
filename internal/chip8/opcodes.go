package chip8

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
func (v *VM) registerHandlers() {
	v.handlers = map[uint16]opcodeHandler{
		0x0000: {
			opcode:  "handle0x0000",
			handler: v.handle0x0000,
		},
		0x1000: {
			opcode:  "1NNN",
			handler: v.jump,
		},
		0x2000: {
			opcode:  "2NNN",
			handler: v.callSub,
		},
		0x3000: {
			opcode:  "3XNN",
			handler: v.skipVxNN,
		},
		0x4000: {
			opcode:  "4XNN",
			handler: v.skipVxNotNN,
		},
		0x6000: {
			opcode:  "6XNN",
			handler: v.setVx,
		},
		0x7000: {
			opcode:  "7XNN",
			handler: v.incVx,
		},
		0x8000: {
			opcode:  "handle0x8000",
			handler: v.handle0x8000,
		},
		0x9000: {
			opcode:  "9XY0",
			handler: v.skipVxNotVy,
		},
		0xA000: {
			opcode:  "ANNN",
			handler: v.setAddress,
		},
		0xC000: {
			opcode:  "CXNN",
			handler: v.setVxRand,
		},
		0xD000: {
			opcode:  "DXYN",
			handler: v.draw,
		},
		0xE000: {
			opcode:  "handle0xE000",
			handler: v.handle0xE000,
		},
		0xF000: {
			opcode:  "handle0xF000",
			handler: v.handle0xF000,
		},
	}
}

// handle attempts to handle the current opcode.
//
// If there is no registered handler for opc then an error is returned. The
// handler can also return an error.
func (v *VM) handle() error {
	// Note that we are not using v.opc to determine the handler here because
	// in the majority of cases the first 4 bits of the opcode are what
	// indicates the action to take.
	h, ok := v.handlers[v.opc&0xF000]
	if !ok {
		return fmt.Errorf("unsupported opcode: 0x%X", v.opc)
	}

	// Handle the opcode.
	val, err := h.handler()
	if err != nil {
		return fmt.Errorf("error handling opcode: %s value: 0x%X: %s", h.opcode, val, err)
	}

	if v.Debug {
		log.Printf("opcode: %s value: 0x%X\n", h.opcode, val)
	}

	return nil
}

// handle0x0000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (v *VM) handle0x0000() (uint16, error) {
	switch {
	case v.opc&0x00FF == 0x00E0:
		return v.clrDisp()

	case v.opc&0x00FF == 0x00EE:
		return v.subRet()

	case v.opc&0xF000 == 0x0000:
		return v.callSys()

	default:
		return v.opc, errors.New("unknown opcode [0x0000]")
	}
}

// clrDisp clears the display.
func (v *VM) clrDisp() (uint16, error) {
	return v.opc & 0x00FF, errors.New("TODO: clrDisp")
}

// subRet returns from a subroutine.
func (v *VM) subRet() (uint16, error) {
	// Return to the program counter stored in the stack (adding 2 for the
	// next instruction as usual).
	v.pc = v.stack[v.sp] + 2
	v.sp--

	return v.opc & 0x00FF, nil
}

// callSys calls RCA 1802 program at address NNN. Not necessary for most ROMs.
func (v *VM) callSys() (uint16, error) {
	return v.opc & 0xF000, errors.New("TODO: callSys")
}

// jump jumps to address NNN.
func (v *VM) jump() (uint16, error) {
	// Jump to NNN.
	v.pc = v.opc & 0x0FFF

	return v.opc, nil
}

// callSub calls subroutine at NNN.
func (v *VM) callSub() (uint16, error) {
	// Store the current program counter temporarily while we jump to
	// the subroutine. Incrementing the stack pointer to prevent overwrite.
	v.sp++
	v.stack[v.sp] = v.pc

	// Jump to NNN.
	v.pc = v.opc & 0x0FFF

	return v.opc, nil
}

// skipVxNN skips the next instruction if VX equals NN. Usually the next instruction
// is a jump to skip a code block.
func (v *VM) skipVxNN() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(v.opc & 0x00FF) // Get the last 2 chars.

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if v.v[x] == nn {
		v.pc += 4
	} else {
		v.pc += 2
	}

	return v.opc, nil
}

// skipVxNN skips the next instruction if VX does not equal NN. Usually the next
// instruction is a jump to skip a code block.
func (v *VM) skipVxNotNN() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(v.opc & 0x00FF) // Get the last 2 chars.

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if v.v[x] != nn {
		v.pc += 4
	} else {
		v.pc += 2
	}

	return v.opc, nil
}

// setVx sets CPU register Vx (where x is A-E) to NN.
func (v *VM) setVx() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(v.opc & 0x00FF) // Get the last 2 chars.

	v.v[x] = nn
	v.pc += 2

	return v.opc, nil
}

// incVx adds NN to CPU register Vx.
func (v *VM) incVx() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(v.opc & 0x00FF) // Get the last 2 chars.

	v.v[x] += nn
	v.pc += 2

	return v.opc, nil
}

// handle0x8000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (v *VM) handle0x8000() (uint16, error) {
	switch v.opc & 0x000F {
	case 0x0000:
		return v.setVxVy()

	case 0x0002:
		return v.setVxAndVy()

	case 0x0004:
		return v.incVxVy()

	case 0x0005:
		return v.decVxVy()

	default:
		return v.opc & 0xFFFF, errors.New("TODO: handle0x8000")
	}
}

// setVxVy sets VX to the value of VY.
func (v *VM) setVxVy() (uint16, error) {
	v.v[(v.opc&0x0F00)>>8] = v.v[(v.opc&0x00F0)>>4]
	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// setVxAndVy sets VX to VX & VY (bitwise AND operation).
func (v *VM) setVxAndVy() (uint16, error) {
	v.v[(v.opc&0x0F00)>>8] &= v.v[(v.opc&0x00F0)>>4]
	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// incVxVy adds VY to VX. VF is set to 1 when there's a carry, and to 0 when
// there isn't.
func (v *VM) incVxVy() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8
	y := (v.opc & 0x00F0) >> 4

	if v.v[y] > (0xFF - v.v[x]) {
		v.v[0xF] = 1
	} else {
		v.v[0xF] = 0
	}
	v.v[x] += v.v[y]

	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// decVxVy VY is subtracted from VX. VF is set to 0 when there's a borrow, and 1
// when there isn't.
func (v *VM) decVxVy() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8
	y := (v.opc & 0x00F0) >> 4

	if v.v[y] > v.v[x] {
		v.v[0xF] = 0
	} else {
		v.v[0xF] = 1
	}
	v.v[x] -= v.v[y]

	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// skipVxNotVy skips the next instruction if VX doesn't equal VY. Usually the
// next instruction is a jump to skip a code block.
func (v *VM) skipVxNotVy() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8
	y := (v.opc & 0x00F0) >> 4

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if v.v[x] != v.v[y] {
		v.pc += 4
	} else {
		v.pc += 2
	}

	return v.opc, nil
}

// setAddress sets the index register to the address NNN.
func (v *VM) setAddress() (uint16, error) {
	v.i = v.opc & 0x0FFF
	v.pc += 2 // Increase by 2 because each instruction is 2 bytes long.

	return v.opc, nil
}

// setVxRand sets VX to the result of a bitwise and operation on a random
// number (typically: 0 to 255) and NN.
func (v *VM) setVxRand() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(v.opc & 0x00FF) // Get the last 2 chars.

	v.v[x] = byte(rand.Intn(256)) & nn
	v.pc += 2

	return v.opc, nil
}

// draw draws a sprite at coordinate (VX, VY) that has a width of 8 pixels
// and a height of N pixels. Each row of 8 pixels is read as bit-coded starting
// from memory location i; i doesn't change after the execution of this
// instruction. VF is set to 1 if any screen pixels are flipped from set to
// unset when the sprite is drawn, and to 0 if that doesn't happen.
func (v *VM) draw() (uint16, error) {
	var (
		pixel  uint16
		x      = uint16(v.v[(v.opc&0x0F00)>>8])
		y      = uint16(v.v[(v.opc&0x00F0)>>4])
		height = v.opc & 0x000F
	)
	v.v[0xF] = 0

	for cY := uint16(0); cY < height; cY++ {
		pixel = uint16(v.mem[v.i+cY])
		for cX := uint16(0); cX < 8; cX++ {
			index := x + cX + ((y + cY) * 64)
			if pixel&(0x80>>cX) == 0 {
				continue
			}

			// If the pixel was already 'lit', set the VF register to 1.
			// This indicates a collision.
			if v.disp[index] == 1 {
				v.v[0xF] = 1
			}

			// Bitwise XOR to 'flip' the pixel.
			v.disp[index] ^= 1
		}
	}

	v.drawChan <- struct{}{}
	v.pc += 2

	return v.opc, nil
}

// handle0xF000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (v *VM) handle0xE000() (uint16, error) {
	switch v.opc & 0x00FF {
	case 0x00A1:
		return v.skipVxKeyNotPressed()
	default:
		return v.opc & 0xFFFF, errors.New("TODO: handle0xF000")
	}
}

// skipVxKeyNotPressed the next instruction if the key stored in VX isn't
// pressed (usually the next instruction is a jump to skip a code block).
func (v *VM) skipVxKeyNotPressed() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if v.key[v.v[x]] == 0 {
		v.pc += 4
	} else {
		v.pc += 2
	}

	return v.opc & 0xFFFF, nil
}

// handle0xF000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (v *VM) handle0xF000() (uint16, error) {
	switch v.opc & 0x00FF {
	case 0x0007:
		return v.getDelayTimer()

	case 0x0015:
		return v.setDelayTimer()

	case 0x0018:
		return v.setSoundTimer()

	case 0x0029:
		return v.loadFont()

	case 0x0033:
		return v.setBCD()

	case 0x0065:
		return v.regLoad()

	default:
		return v.opc & 0xFFFF, errors.New("TODO: handle0xF000")
	}
}

// getDelayTimer the delay timer to VX.
func (v *VM) getDelayTimer() (uint16, error) {
	v.v[(v.opc&0x0F00)>>8] = v.delayTimer
	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// setDelayTimer sets the delay timer to VX.
func (v *VM) setDelayTimer() (uint16, error) {
	v.delayTimer = v.v[(v.opc&0x0F00)>>8]
	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// setDelayTimer sets the sound timer to VX.
func (v *VM) setSoundTimer() (uint16, error) {
	v.soundTimer = v.v[(v.opc&0x0F00)>>8]
	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// loadFont sets i to the location of the sprite for the character in VX.
// Characters 0-F (in hexadecimal) are represented by a 4x5 font.
func (v *VM) loadFont() (uint16, error) {
	v.i = uint16(v.v[(v.opc&0x0F00)>>8]) * 5
	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// setBCD stores the binary code decimal representation of VX with the most
// significant of three digits at the address in i, the middle digit at i plus
// 1, and the least significant digit at i plus 2. (in other words, take the
// decimal representation of VX, place the hundreds digit in mem at location
// in i, the tens digit at location i+1, and the ones digit at location i+2).1
func (v *VM) setBCD() (uint16, error) {
	x := (v.opc & 0x0F00) >> 8

	v.mem[v.i] = v.v[x] / 100          // Hundreds.
	v.mem[v.i+1] = (v.v[x] / 10) % 10  // Tens.
	v.mem[v.i+2] = (v.v[x] % 100) % 10 // Ones.
	v.pc += 2

	return v.opc & 0xFFFF, nil
}

// fillV0Vx stores V0 to VX (including VX) in mem starting at address i. The
// offset from i is increased by 1 for each value written, but i itself is left
// unmodified.
func (v *VM) regLoad() (uint16, error) {
	for i := uint16(0); i <= (v.opc&0x0F00)>>8; i++ {
		v.v[i] = v.mem[v.i+i]
	}
	v.pc += 2

	return v.opc & 0xFFFF, nil
}
