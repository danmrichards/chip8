package main

import (
	"io"
	"io/ioutil"
)

// opcode represents a Chip8 opcode which are 2 bytes long.
type opcode uint16

type chip8 struct {
	// Stores the current opcode.
	opc opcode

	// Chip8 has 4K of memory. The first 512 bytes are reserved
	// for the the interpreter, fonts etc. Programs are expected to start at
	// 0x200. See map:
	//
	// 0x000 -> 0x1FF - Chip 8 interpreter (contains font set in emu)
	// 0x050 -> 0x0A0 - Used for the built in 4x5 pixel font set (0-F)
	// 0x200 -> 0xFFF - Program ROM and work RAM
	memory [4096]byte

	// CPU registers. The Chip 8 has 15 8-bit general purpose registers named
	// V0,V1...VE. The 16th register (VF) is used  for the 'carry flag'.
	V [16]byte

	// Index register.
	I uint16

	// Program counter.
	pc uint16

	// Chip8 display resolution is 64x32 pixels in monochrome. Drawing is done
	// in XOR mode and if a pixel is turned off as a result of drawing, the VF
	// register is set. This is used for collision detection.
	disp [64 * 32]byte

	// Interrupts and hardware registers. The Chip 8 has none, but there are two
	// timer registers that count at 60 Hz. When set above zero they will count
	// down to zero.
	delayTimer byte // Intended for timing game events, can be set or read.
	soundTimer byte // Used for sound events. Beep when value is non-zero.

	// The Chip8 instruction set has opcodes that allow the program to jump to
	// addresses or call subroutines. Consequently we need a stack to remember
	// locations before a jump is performed. Chip8 has 16 levels of stack.
	stack [16]uint16

	// The stack pointer is used to remember which level of stack is being used.
	sp uint16

	// Chip 8 has a HEX based keypad (0x0-0xF).
	key [16]byte
}

// reset initialises the Chip8 registers and memory.
func (c *chip8) reset() {
	c.opc = 0                // Reset current opcode.
	c.memory = [4096]byte{}  // Clear memory
	c.V = [16]byte{}         // Clear registers V0-VF
	c.I = 0                  // Reset the index register.
	c.pc = 0x200             // Program counter starts at 0x200.
	c.sp = 0                 // Reset the stack pointer.
	c.disp = [64 * 32]byte{} // Clear display
	c.stack = [16]uint16{}   // Clear stack

	// Load the font set into memory.
	for i := 0; i < 80; i++ {
		c.memory[i] = fontset[i]
	}

	// Reset timers
	c.delayTimer, c.soundTimer = 0, 0
}

// cycle emulates one clock cycle of the Chip8 CPU.
func (c *chip8) cycle() {
	// TODO: fetch opcode
	// TODO: decode opcode
	// TODO: execute opcode
	// TODO: update timers
}

// load loads the contents of rom into memory.
func (c *chip8) load(rom io.Reader) error {
	data, err := ioutil.ReadAll(rom)
	if err != nil {
		return err
	}

	// Load byte into memory, offset by 512 bytes (0x200).
	for i := 0; i < len(data); i++ {
		c.memory[i+0x200] = data[i]
	}

	return nil
}

func newChip8() *chip8 {
	c := &chip8{}
	c.reset()

	return c
}
