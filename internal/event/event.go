package event

import (
	"log"

	"github.com/danmrichards/chip8/internal/chip8"
	"github.com/danmrichards/chip8/internal/sound"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var keys = map[byte]pixelgl.Button{
	0x1: pixelgl.Key1, 0x2: pixelgl.Key2, 0x3: pixelgl.Key3, 0xC: pixelgl.Key4,
	0x4: pixelgl.KeyQ, 0x5: pixelgl.KeyW, 0x6: pixelgl.KeyE, 0xD: pixelgl.KeyR,
	0x7: pixelgl.KeyA, 0x8: pixelgl.KeyS, 0x9: pixelgl.KeyD, 0xE: pixelgl.KeyF,
	0xA: pixelgl.KeyZ, 0x0: pixelgl.KeyX, 0xB: pixelgl.KeyC, 0xF: pixelgl.KeyV,
}

// Handler is responsible for handling input and output for the vm.
type Handler struct {
	window *pixelgl.Window
	vm     *chip8.VM
}

// NewHandler returns a new event handler.
func NewHandler(win *pixelgl.Window, vm *chip8.VM) Handler {
	return Handler{
		window: win,
		vm:     vm,
	}
}

// Handle continually loops while the VM window is open; handling events.
// Events are handled with a non-blocking select. Draw and sound events are
// handled independently with input being treated as the default event to check.
func (h *Handler) Handle() {
	for !h.window.Closed() {
		select {
		case <-h.vm.Draw():
			h.draw()
		case <-h.vm.Beep():
			if err := sound.Beep(); err != nil {
				log.Printf("Error playing beep: %q\n", err)
			}
		default:
			h.input()
		}
	}
}

// input iterates over the keyset, checking if any of them are pressed and
// updates the vm accordingly.
func (h *Handler) input() {
	for i, key := range keys {
		if h.window.Pressed(key) {
			h.vm.KeyDown(i)
		}
	}
}

// draw updates the window based on the current state of the VM graphics array.
func (h *Handler) draw() {
	h.window.Clear(colornames.Black)

	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(0.14, 0.8, 0.26)

	scrW := h.window.Bounds().W()
	scrH := h.window.Bounds().H()

	// Calculate the screen ratio.
	rW, rH := scrW/64, scrH/32

	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if !h.vm.PixelSet((31-y)*64 + x) {
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

	imd.Draw(h.window)
	h.window.Update()
}
