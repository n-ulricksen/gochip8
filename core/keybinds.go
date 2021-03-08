package core

import "github.com/veandco/go-sdl2/sdl"

var keybinds = map[int]uint8{
	sdl.SCANCODE_7:         0x1,
	sdl.SCANCODE_8:         0x2,
	sdl.SCANCODE_9:         0x3,
	sdl.SCANCODE_0:         0xc,
	sdl.SCANCODE_U:         0x4,
	sdl.SCANCODE_I:         0x5,
	sdl.SCANCODE_O:         0x6,
	sdl.SCANCODE_P:         0xd,
	sdl.SCANCODE_J:         0x7,
	sdl.SCANCODE_K:         0x8,
	sdl.SCANCODE_L:         0x9,
	sdl.SCANCODE_SEMICOLON: 0xe,
	sdl.SCANCODE_M:         0xa,
	sdl.SCANCODE_COMMA:     0x0,
	sdl.SCANCODE_PERIOD:    0xb,
	sdl.SCANCODE_SLASH:     0xf,
}
