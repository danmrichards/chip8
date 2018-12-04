package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/danmrichards/chip8/internal/chip8"
	"github.com/danmrichards/chip8/internal/event"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var (
	vm *chip8.VM

	rom   string
	debug bool
)

const cycleRate = 300

func main() {
	log.SetFlags(log.LstdFlags)

	flag.StringVar(&rom, "rom", "", "Path to the ROM file to load")
	flag.BoolVar(&debug, "debug", false, "Run the emulator in debug mode")
	flag.Parse()

	// Validate the ROM flag.
	if rom == "" {
		fmt.Println("ROM flag is required")
		os.Exit(1)
	}
	if _, err := os.Stat(rom); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("ROM %q does not exist", rom)
			os.Exit(1)
		} else {
			log.Fatal(err)
		}
	}

	pixelgl.Run(run)
}

func run() {
	tick := time.NewTicker(time.Second / cycleRate)
	defer tick.Stop()

	cfg := pixelgl.WindowConfig{
		Title:  "chip8",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}

	window, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatal("Could not create event:", err)
	}

	vm = chip8.New()
	vm.Debug = debug

	eh := event.NewHandler(window, vm)

	rom, err := os.Open(rom)
	if err != nil {
		log.Fatalln("Could not open ROM:", err)
	}

	if err := vm.Load(rom); err != nil {
		log.Fatal("Could not load ROM:", err)
	}

	// Handle input, screen and sound events.
	go eh.Handle()

	// Emulation loop.
	for !window.Closed() {
		window.UpdateInput()

		if window.Pressed(pixelgl.KeyEscape) {
			break
		}

		// Emulate a cycle.
		if err = vm.Cycle(); err != nil {
			log.Fatal(err)
		}

		// A bit dirty, but block the next cycle until a tick. This prevents
		// the emulator from running too quickly.
		<-tick.C
	}
}
