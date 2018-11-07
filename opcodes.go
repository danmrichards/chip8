package main

import (
	"github.com/pkg/errors"
)

// opcodeHandler handles the opcode opc and returns an error if it occurred.
type opcodeHandler func() error

// registerHandlers registers up the map of opcode handlers
func (c *chip8) registerHandlers() {
	// TODO: names for opcodes.
	c.handlers = map[uint16]opcodeHandler{
		0x0000: c.handle0x0000,
		0x1000: c.jump,
		0x2000: c.callSub,
		0x3000: c.skipVxNN,
		0x6000: c.setVx,
		0x7000: c.incVx,
		0x9000: c.skipVxNotVy,
		0xA000: c.setAddress,
		0xD000: c.drawDisp,
		0xF000: c.handle0xF000,
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
		return errHandleOpcode{
			opc: c.opc,
			err: errors.New("unsupported opcode"),
		}
	}

	// TODO: debug mode support. Log opcode and handler name.

	return h()
}

// handle0x0000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (c *chip8) handle0x0000() error {
	switch {
	case c.opc&0x00FF == 0x00E0:
		return c.clrDisp()

	case c.opc&0x00FF == 0x00EE:
		return c.subRet()

	case c.opc&0xF000 == 0x0000:
		return c.callSys()

	default:
		return errHandleOpcode{
			opc: c.opc,
			err: errors.New("unknown opcode [0x0000]"),
		}
	}
}

// callSys calls RCA 1802 program at address NNN. Not necessary for most ROMs.
func (c *chip8) callSys() error {
	return errHandleOpcode{
		opc: c.opc,
		err: errors.New("TODO: callSys"),
	}
}

// clrDisp clears the display.
func (c *chip8) clrDisp() error {
	return errHandleOpcode{
		opc: c.opc,
		err: errors.New("TODO: clrDisp"),
	}
}

// subRet returns from a subroutine.
func (c *chip8) subRet() error {
	// Return to the program counter stored in the stack (adding 2 for the
	// next instruction as usual).
	c.pc = c.stack[c.sp] + 2
	c.sp--

	return nil
}

// jump jumps to address NNN.
func (c *chip8) jump() error {
	// Jump to NNN.
	c.pc = c.opc & 0x0FFF

	return nil
}

// callSub calls subroutine at NNN.
func (c *chip8) callSub() error {
	// Store the current program counter temporarily while we jump to
	// the subroutine. Incrementing the stack pointer to prevent overwrite.
	c.sp++
	c.stack[c.sp] = c.pc

	// Jump to NNN.
	c.pc = c.opc & 0x0FFF
	return nil
}

// skipVxNN skips the next instruction if VX equals NN. Usually the next instruction
// is a jump to skip a code block.
func (c *chip8) skipVxNN() error {
	x := (c.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(c.opc & 0x00FF) // Get the last 2 chars.

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if c.V[x] == nn {
		c.pc += 4
	} else {
		c.pc += 2
	}

	return nil
}

// setVx sets CPU register Vx (where x is A-E) to NN.
func (c *chip8) setVx() error {
	x := (c.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(c.opc & 0x00FF) // Get the last 2 chars.

	c.V[x] = nn
	c.pc += 2

	return nil
}

// incVx adds NN to CPU register Vx.
func (c *chip8) incVx() error {
	x := (c.opc & 0x0F00) >> 8 // Reverse the shift.
	nn := byte(c.opc & 0x00FF) // Get the last 2 chars.

	c.V[x] += nn
	c.pc += 2

	return nil
}

// skipVxNotVy skips the next instruction if VX doesn't equal VY. Usually the
// next instruction is a jump to skip a code block.
func (c *chip8) skipVxNotVy() error {
	x := (c.opc & 0x0F00) >> 8
	y := (c.opc & 0x00F0) >> 4

	// Skip the next instruction by increasing the program counter by 4
	// instead of the usual 2.
	if c.V[x] != c.V[y] {
		c.pc += 4
	} else {
		c.pc += 2
	}

	return nil
}

// setAddress sets the index register to the address NNN.
func (c *chip8) setAddress() error {
	c.I = c.opc & 0x0FFF
	c.pc += 2 // Increase by 2 because each instruction is 2 bytes long.
	return nil
}

// drawDisp draws a sprite at coordinate (VX, VY) that has a width of 8 pixels
// and a height of N pixels. Each row of 8 pixels is read as bit-coded starting
// from memory location I; I doesn't change after the execution of this
// instruction. VF is set to 1 if any screen pixels are flipped from set to
// unset when the sprite is drawn, and to 0 if that doesn't happen.
func (c *chip8) drawDisp() error {
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

	return nil
}

// handle0xF000 performs additional opcode parsing to determine the correct
// action. Codes in this range cannot rely on the first 4 bits.
func (c *chip8) handle0xF000() error {
	return errHandleOpcode{
		opc: c.opc,
		err: errors.New("TODO: handle0xF000"),
	}
}
