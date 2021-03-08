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
	chip8frequency                = 60 * 8
)

// The Chip8 emulator
type Chip8 struct {
	mem      []byte
	cpu      *CPU
	display  []uint8
	renderer *sdl.Renderer
}

// NewChip8 creates a new Chip8 emulator with 4KB RAM.
func NewChip8() *Chip8 {
	w, h := Chip8Width, Chip8Height

	memory := make([]byte, memorySize)
	copy(memory[characterSpritesOffset:], characterSprites)

	return &Chip8{
		mem:      memory,
		cpu:      NewCPU(),
		display:  make([]uint8, w*h),
		renderer: NewDisplayRenderer(),
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
	vBlankTime := chip8frequency / VBlankFreq
	cycle := 0

	for {
		cycle++

		if cycle > vBlankTime {
			cycle = 0
			c.RenderDisplay()
		}

		c.Cycle()
		time.Sleep(17 * time.Millisecond) // remove this
	}
}

// RenderDisplay presents the current display to the screen via the SDL2 renderer.
func (c *Chip8) RenderDisplay() {
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

// Cycle spins the CPU, executing instructions from RAM.
func (c *Chip8) Cycle() {
	c.getNextInstruction()

	// Increment the program counter
	c.cpu.pc += 2

	// Execute the instruction
	c.executeInstruction()

	//fmt.Printf("CPU: %#v\n\n", *cpu)
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
			log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
		}
	case 0x1000:
		c.cpu.Exec1NNN()
	case 0x2000:
		c.cpu.Exec2NNN()
	case 0x3000:
		c.cpu.Exec3XNN()
	case 0x6000:
		c.cpu.Exec6XNN()
	case 0x7000:
		c.cpu.Exec7XNN()
	case 0x8000:
		switch c.cpu.opcode.n() {
		case 0x0:
			c.cpu.Exec8XY0()
		case 0x2:
			c.cpu.Exec8XY2()
		case 0x3:
			c.cpu.Exec8XY3()
		case 0x6:
			c.cpu.Exec8XY6()
		case 0xE:
			c.cpu.Exec8XYE()
		default:
			log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
		}
	case 0xA000:
		c.cpu.ExecANNN()
	case 0xD000:
		c.cpu.ExecDXYN(&c.mem, &c.display)
	case 0xF000:
		switch c.cpu.opcode.nn() {
		case 0x55:
			c.cpu.ExecFX55(&c.mem)
		case 0x65:
			c.cpu.ExecFX65(&c.mem)
		case 0x1E:
			c.cpu.ExecFX1E()
		default:
			log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
		}
	default:
		log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
	}
}
