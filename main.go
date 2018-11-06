package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	sig  = make(chan os.Signal, 1)
	loop = true
)

func main() {
	log.SetFlags(log.LstdFlags)

	// Handle sig for exit. TODO: Replace with SDL exit.
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		loop = false
	}()

	// TODO: Setup graphics.
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		log.Fatal(err)
	}

	// TODO: Setup input.

	c8 := newChip8()

	// TODO: Load rom from flag.
	rom, err := os.Open("PONG")
	if err != nil {
		log.Fatalln("Could not open rom:", err)
	}

	if err := c8.load(rom); err != nil {
		log.Fatal("Could not load rom:", err)
	}

	// Emulation loop.
	for loop {
		select {
		case <-sig:
			break
		default:
		}

		// Emulate a cycle.
		if err = c8.cycle(); err != nil {
			log.Fatal(err)
		}

		// TODO: Update screen if the draw flag is set.

		// TODO: Store the key press state.
	}
}
