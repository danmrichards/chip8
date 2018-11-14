package chip8

import (
	"io"
	"io/ioutil"
	"time"
)

// VM is an implementation of the Chip8 virtual machine.
type VM struct {
	Debug bool

	// Stores the current opcode.
	opc uint16

	// Chip8 has 4K of mem. The first 512 bytes are reserved
	// for the the interpreter, fonts etc. Programs are expected to start at
	// 0x200. See map:
	//
	// 0x000 -> 0x1FF - Chip 8 interpreter (contains font set in emu)
	// 0x050 -> 0x0A0 - Used for the built in 4x5 pixel font set (0-F)
	// 0x200 -> 0xFFF - Program ROM and work RAM
	mem [4096]byte

	// CPU registers. The Chip 8 has 15 8-bit general purpose registers named
	// V0,V1...VE. The 16th register (VF) is used  for the 'carry flag'.
	v [16]byte

	// Index register.
	i uint16

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
	keys [16]byte

	// Clock will run at 60Hz to keep the cycles at the correct speed.
	clock *time.Ticker

	// Each supported opcode has handler func.
	handlers map[uint16]opcodeHandler

	// Delivered to when the screen should be drawn.
	drawChan chan struct{}

	// Delivered to when a beep should be made.
	beepChan chan struct{}
}

func New() *VM {
	v := &VM{
		drawChan: make(chan struct{}),
		beepChan: make(chan struct{}),
	}
	v.reset()

	return v
}

// Cycle emulates one clock cycle of the Chip8 CPU.
func (v *VM) Cycle() error {
	// Set the current opcode. The opcodes are two bytes long so we get two
	// of them and merge together.
	v.opc = uint16(v.mem[v.pc])<<8 | uint16(v.mem[v.pc+1])

	// Handle the opcode.
	if err := v.handle(); err != nil {
		return err
	}

	// Use the ticker to slow down emulation cycles to a realistic speed.
	select {
	case <-v.clock.C:
		v.updateTimers()
	default:
	}

	return nil
}

// Load loads the contents of rom into mem.
func (v *VM) Load(rom io.Reader) error {
	data, err := ioutil.ReadAll(rom)
	if err != nil {
		return err
	}

	// Load byte into mem, offset by 512 bytes (0x200).
	for i := 0; i < len(data); i++ {
		v.mem[i+0x200] = data[i]
	}

	return nil
}

// PixelSet returns true if the pixel at i is set.
func (v *VM) PixelSet(i int) bool {
	return v.disp[i] == 1
}

// Draw returns a read-only channel indicating when the screen should be drawn.
func (v *VM) Draw() <-chan struct{} {
	return v.drawChan
}

// Beep returns a read-only channel indicating when a beep should happen.
func (v *VM) Beep() <-chan struct{} {
	return v.beepChan
}

// KeyDown marks key as pressed.
func (v *VM) KeyDown(key uint16) {
	v.keys[key] = 1
}

// updateTimers updates the chip8 timers dispatching any additional events
// based on the timer values.
func (v *VM) updateTimers() {
	if v.delayTimer > 0 {
		v.delayTimer--
	}
	if v.soundTimer > 0 {
		if v.soundTimer == 1 {
			v.beepChan <- struct{}{}
		}
		v.soundTimer--
	}
}

// reset initialises the Chip8 registers and mem.
func (v *VM) reset() {
	v.opc = 0                // Reset current opcode.
	v.mem = [4096]byte{}     // Clear mem
	v.v = [16]byte{}         // Clear registers V0-VF
	v.i = 0                  // Reset the index register.
	v.pc = 0x200             // Program counter starts at 0x200.
	v.sp = 0                 // Reset the stack pointer.
	v.disp = [64 * 32]byte{} // Clear display
	v.stack = [16]uint16{}   // Clear stack

	// Load the font set into mem.
	for i := 0; i < 80; i++ {
		v.mem[i] = fontset[i]
	}

	// Reset timers
	v.delayTimer, v.soundTimer = 0, 0

	v.clock = time.NewTicker(time.Second / 60)

	v.registerHandlers()
}
