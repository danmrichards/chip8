package main

import (
	"log"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

// TODO: Much better package structure.

func main() {
	log.SetFlags(log.LstdFlags)

	pixelgl.Run(func() {
		// TODO: Flags for rom and debug mode.

		// TODO: Abstract graphics to package/interface.
		cfg := pixelgl.WindowConfig{
			Title:  "chip8",
			Bounds: pixel.R(0, 0, 1024, 768),
			VSync:  true,
		}
		window, err := pixelgl.NewWindow(cfg)
		if err != nil {
			log.Fatal("Could not create window:", err)
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
		for !window.Closed() {
			if window.Pressed(pixelgl.KeyEscape) {
				break
			}

			// Emulate a cycle.
			if err = c8.cycle(); err != nil {
				log.Fatal(err)
			}

			if c8.draw {
				updateDisp(c8, window)
				c8.draw = false
			}

			window.UpdateInput()

			// TODO: Store the key press state.
		}
	})
}

func updateDisp(c8 *chip8, window *pixelgl.Window) {
	window.Clear(colornames.Black)

	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(1, 1, 1)

	screenWidth := window.Bounds().W()
	screenHeight := window.Bounds().H()
	width, height := screenWidth/64, screenHeight/32

	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			// TODO: Move pixel check into c8.
			if c8.disp[(31-y)*64+x] == 1 {
				imd.Push(pixel.V(width*float64(x), height*float64(y)))
				imd.Push(pixel.V(width*float64(x)+width, height*float64(y)+height))
				imd.Rectangle(0)
			}
		}
	}

	imd.Draw(window)
	window.Update()
}
