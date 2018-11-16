package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/danmrichards/chip8/internal/chip8"
	"github.com/danmrichards/chip8/internal/sound"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var (
	vm     *chip8.VM
	window *pixelgl.Window

	// TODO: Abstract this.
	keys = map[byte]pixelgl.Button{
		0x1: pixelgl.Key1, 0x2: pixelgl.Key2, 0x3: pixelgl.Key3, 0xC: pixelgl.Key4,
		0x4: pixelgl.KeyQ, 0x5: pixelgl.KeyW, 0x6: pixelgl.KeyE, 0xD: pixelgl.KeyR,
		0x7: pixelgl.KeyA, 0x8: pixelgl.KeyS, 0x9: pixelgl.KeyD, 0xE: pixelgl.KeyF,
		0xA: pixelgl.KeyZ, 0x0: pixelgl.KeyX, 0xB: pixelgl.KeyC, 0xF: pixelgl.KeyV,
	}

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

	var err error
	window, err = pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatal("Could not create window:", err)
	}

	vm = chip8.New()
	vm.Debug = debug

	rom, err := os.Open(rom)
	if err != nil {
		log.Fatalln("Could not open ROM:", err)
	}

	if err := vm.Load(rom); err != nil {
		log.Fatal("Could not load ROM:", err)
	}

	go eventHandler()

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

		inputHandler()

		// A bit dirty, but block the next cycle until a tick. This prevents
		// the emulator from running too quickly.
		<-tick.C
	}
}

func eventHandler() {
	for !window.Closed() {
		select {
		case <-vm.Draw():
			drawScreen()
		case <-vm.Beep():
			if err := sound.Beep(); err != nil {
				log.Printf("Error playing beep: %q\n", err)
			}
		default:
		}
	}
}

// TODO: Abstract this.
func drawScreen() {
	window.Clear(colornames.Black)

	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(0.14, 0.8, 0.26)

	scrW := window.Bounds().W()
	scrH := window.Bounds().H()

	// Calculate the screen ratio.
	rW, rH := scrW/64, scrH/32

	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if !vm.PixelSet((31-y)*64 + x) {
				continue
			}

			// Scale the pixel co-ords.
			sX := rW * float64(x)
			sY := rH * float64(y)

			imd.Push(pixel.V(sX, sY))
			imd.Push(pixel.V(sX+rW, sY+rH))
			imd.Rectangle(0)
		}
	}

	imd.Draw(window)
	window.Update()
}

func inputHandler() {
	for i, key := range keys {
		if window.Pressed(key) {
			vm.KeyDown(i)
		}
	}
}
