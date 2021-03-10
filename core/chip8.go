package core

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	memorySize             uint16 = 4096
	programEntryOffset     uint16 = 0x200
	characterSpritesOffset uint16 = 0x100
	characterSpriteBytes          = 5
	chip8frequency                = 60 * 8
)

// The Chip8 emulator
type Chip8 struct {
	mem       []byte // RAM
	cpu       *CPU
	display   []uint8 // emulator display
	keys      []uint8 // current state of each key
	renderer  *sdl.Renderer
	isRunning bool
}

// NewChip8 creates a new Chip8 emulator with 4KB RAM.
func NewChip8() *Chip8 {
	w, h := Chip8Width, Chip8Height

	// Initialize memory.
	memory := make([]byte, memorySize)
	copy(memory[characterSpritesOffset:], characterSprites)

	// Initialize SDL.
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatal("Unable to initialize SDL", err)
	}

	return &Chip8{
		mem:       memory,
		cpu:       NewCPU(),
		display:   make([]uint8, w*h),
		keys:      make([]uint8, 16),
		renderer:  NewDisplayRenderer(),
		isRunning: true,
	}
}

// LoadRom loads a Chip-8 ROM from the specified path into the Chip-8 RAM.
func (c *Chip8) LoadRom(path string) {
	// Load rom from file
	romdata, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error opening ROM file %s\n%v\n", path, err)
	}

	fmt.Println("ROM loading...")

	// Load rom data into RAM
	for i, data := range romdata {
		c.mem[int(programEntryOffset)+i] = data
	}
}

// Run begins execution of program instructions.
func (c *Chip8) Run() {
	defer c.renderer.Destroy()
	defer sdl.Quit()

	vBlankTime := chip8frequency / VBlankFreq
	cycles := 0

	for c.isRunning {
		cycles++

		c.cycle()

		if cycles > vBlankTime {
			cycles = 0
			c.renderDisplay()

			c.cpu.decrementTimers()
		}

		c.pollSdlEvents()

		time.Sleep(5 * time.Millisecond) // remove this
	}
}

// renderDisplay presents the current display to the screen via the SDL2 renderer.
func (c *Chip8) renderDisplay() {
	c.renderer.SetDrawColor(0, 0, 0, 255)
	c.renderer.Clear()

	c.renderer.SetDrawColor(0, 255, 0, 255)

	for y := int32(0); y < Chip8Height; y++ {
		for x := int32(0); x < Chip8Width; x++ {
			if c.display[y*Chip8Width+x] != 0 {
				c.renderer.FillRect(&sdl.Rect{
					X: x * DisplayScale,
					Y: y * DisplayScale,
					W: DisplayScale,
					H: DisplayScale,
				})
			}
		}
	}

	c.renderer.Present()
}

// pollSdlEvents checks for keyboard events.
func (c *Chip8) pollSdlEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			c.isRunning = false
		case *sdl.KeyboardEvent:
			scancode := t.Keysym.Scancode
			switch t.Type {
			case sdl.KEYDOWN:
				if i, ok := keybinds[int(scancode)]; ok {
					c.keys[i] = 1
				}
			case sdl.KEYUP:
				if i, ok := keybinds[int(scancode)]; ok {
					c.keys[i] = 0
				}
			}
		}
	}
}

// cycle spins the CPU, executing instructions from RAM.
func (c *Chip8) cycle() {
	c.getNextInstruction()

	// Increment the program counter
	c.cpu.pc += 2

	// Execute the instruction
	c.executeInstruction()

	// Debug: print mem
	//for i, b := range c.mem {
	//fmt.Printf("%#x: %#x\t", i, b)
	//if i%8 == 0 {
	//fmt.Println()
	//}
	//}
	//fmt.Println()
}

// getNextInstruction loads the next 2 byte instruction into the CPU from memory.
func (c *Chip8) getNextInstruction() {
	bx := c.mem[c.cpu.pc : c.cpu.pc+2]
	c.cpu.opcode = Opcode(binary.BigEndian.Uint16(bx))
}

// executeInstruction executes the appropriate instruction based on the opcode
// currently loaded into the CPU.
func (c *Chip8) executeInstruction() {
	switch c.cpu.opcode & 0xF000 {
	case 0x0000:
		switch c.cpu.opcode.nnn() {
		case 0x0E0:
			c.cpu.Exec00E0(&c.display)
		case 0x0EE:
			c.cpu.Exec00EE()
		default:
			c.invalidOpcode()
		}
	case 0x1000:
		c.cpu.Exec1NNN()
	case 0x2000:
		c.cpu.Exec2NNN()
	case 0x3000:
		c.cpu.Exec3XNN()
	case 0x4000:
		c.cpu.Exec4XNN()
	case 0x5000:
		c.cpu.Exec5XY0()
	case 0x6000:
		c.cpu.Exec6XNN()
	case 0x7000:
		c.cpu.Exec7XNN()
	case 0x8000:
		switch c.cpu.opcode.n() {
		case 0x0:
			c.cpu.Exec8XY0()
		case 0x1:
			c.cpu.Exec8XY1()
		case 0x2:
			c.cpu.Exec8XY2()
		case 0x3:
			c.cpu.Exec8XY3()
		case 0x4:
			c.cpu.Exec8XY4()
		case 0x5:
			c.cpu.Exec8XY5()
		case 0x6:
			c.cpu.Exec8XY6()
		case 0x7:
			c.cpu.Exec8XY7()
		case 0xE:
			c.cpu.Exec8XYE()
		default:
			c.invalidOpcode()
		}
	case 0xA000:
		c.cpu.ExecANNN()
	case 0xC000:
		c.cpu.ExecCXNN()
	case 0xD000:
		c.cpu.ExecDXYN(&c.mem, &c.display)
	case 0xE000:
		switch c.cpu.opcode.nn() {
		case 0x9E:
			c.cpu.ExecEX9E(c.keys)
		case 0xA1:
			c.cpu.ExecEXA1(c.keys)
		default:
			c.invalidOpcode()
		}
	case 0xF000:
		switch c.cpu.opcode.nn() {
		case 0x07:
			c.cpu.ExecFX07()
		case 0x0A:
			c.cpu.ExecFX0A(c.keys)
		case 0x15:
			c.cpu.ExecFX15()
		case 0x1E:
			c.cpu.ExecFX1E()
		case 0x29:
			c.cpu.ExecFX29(&c.mem)
		case 0x33:
			c.cpu.ExecFX33(&c.mem)
		case 0x55:
			c.cpu.ExecFX55(&c.mem)
		case 0x65:
			c.cpu.ExecFX65(&c.mem)
		default:
			c.invalidOpcode()
		}
	default:
		c.invalidOpcode()
	}
}

// invalidOpcode prints an error message displaying the invalid opcode held in
// the cpu, and then terminates execution of the emulator.
func (c *Chip8) invalidOpcode() {
	log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
}
