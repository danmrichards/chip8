package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

	rom   string
	debug bool
)

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

	// TODO: Setup input.

	vm = chip8.New()
	vm.Debug = debug

	rom, err := os.Open(rom)
	if err != nil {
		log.Fatalln("Could not open rom:", err)
	}

	if err := vm.Load(rom); err != nil {
		log.Fatal("Could not load rom:", err)
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

		// TODO: Store the key press state.
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

func drawScreen() {
	window.Clear(colornames.Black)

	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(1, 1, 1)

	screenWidth := window.Bounds().W()
	screenHeight := window.Bounds().H()
	width, height := screenWidth/64, screenHeight/32

	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if vm.PixelSet((31-y)*64 + x) {
				imd.Push(pixel.V(width*float64(x), height*float64(y)))
				imd.Push(pixel.V(width*float64(x)+width, height*float64(y)+height))
				imd.Rectangle(0)
			}
		}
	}

	imd.Draw(window)
	window.Update()
}