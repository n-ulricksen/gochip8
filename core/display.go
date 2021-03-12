package core

import (
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	VBlankFreq     = 60
	Chip8Width     = 64
	Chip8Height    = 32
	DisplayScale   = 10
	EmulatorWidth  = Chip8Width * DisplayScale
	EmulatorHeight = Chip8Height * DisplayScale
	DebugHeight    = 256
)

func NewDisplayRenderer(debug bool) *sdl.Renderer {
	height := int32(EmulatorHeight)
	if debug {
		height += DebugHeight
	}

	window, err := sdl.CreateWindow("Chip-8 Emulator", sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED, EmulatorWidth, height, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal("NewDisplayRenderer error:", err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_PRESENTVSYNC)

	window.Show()

	return renderer
}
