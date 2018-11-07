package main

import (
	"io"
	"io/ioutil"
	"log"
	"time"
)

type chip8 struct {
	// Stores the current opcode.
	opc uint16

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

	// Clock will run at 60Hz to keep the cycles at the correct speed.
	clock *time.Ticker

	// True if the screen should be drawn.
	draw bool

	// Each supported opcode has handler func.
	handlers map[uint16]opcodeHandler

	// TODO: debug mode bool.
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

	c.clock = time.NewTicker(time.Second / 60)

	c.registerHandlers()
}

// cycle emulates one clock cycle of the Chip8 CPU.
func (c *chip8) cycle() error {
	// Set the current opcode. The opcodes are two bytes long so we get two
	// of them and merge together.
	c.opc = uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])

	// Handle the opcode.
	if err := c.handle(); err != nil {
		return err
	}

	// Use the ticker to slow down emulation cycles to a realistic speed.
	select {
	case <-c.clock.C:
		c.updateTimers()
	default:
	}

	return nil
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

// updateTimers updates the chip8 timers dispatching any additional events
// based on the timer values.
func (c *chip8) updateTimers() {
	if c.delayTimer > 0 {
		c.delayTimer--
	}
	if c.soundTimer > 0 {
		if c.soundTimer == 1 {
			// TODO: Actually beep!
			log.Println("BEEP!")
		}
		c.soundTimer--
	}
}

func newChip8() *chip8 {
	c := &chip8{}
	c.reset()

	return c
}
