package core

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/n-ulricksen/chip8/display"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	memorySize         uint16 = 4096
	memoryProgramBegin uint16 = 0x200
	chip8frequency            = 60 * 8
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
	w, h := display.Chip8Width, display.Chip8Height
	return &Chip8{
		mem:      make([]byte, memorySize),
		cpu:      NewCPU(),
		display:  make([]uint8, w*h),
		renderer: display.NewDisplayRenderer(),
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
		c.mem[int(memoryProgramBegin)+i] = data
	}
}

// Run begins execution of program instructions.
func (c *Chip8) Run() {
	vBlankTime := chip8frequency / display.VBlankFreq
	cycle := 0

	for {
		cycle++

		if cycle > vBlankTime {
			cycle = 0
			c.RenderDisplay(&c.display)
		}

		c.Cycle(&c.mem)
		time.Sleep(17 * time.Millisecond)
	}
}

func (c *Chip8) RenderDisplay(disp *[]uint8) {
	c.renderer.SetDrawColor(0, 0, 0, 255)
	c.renderer.Clear()

	c.renderer.SetDrawColor(0, 255, 0, 255)

	for y := int32(0); y < display.Chip8Height; y++ {
		for x := int32(0); x < display.Chip8Width; x++ {
			if (*disp)[y*display.Chip8Width+x] != 0 {
				c.renderer.FillRect(&sdl.Rect{
					X: x * display.DisplayScale,
					Y: y * display.DisplayScale,
					W: display.DisplayScale,
					H: display.DisplayScale,
				})
			}
		}
	}

	c.renderer.Present()
}

// CPU used by the Chip-8 emulator
type CPU struct {
	v      []uint8  // V registers - general purpose
	i      uint16   // I register - general purpose
	pc     uint16   // program counter
	stack  []uint16 // program stack
	sp     uint8    // stack pointer
	dt     uint8    // delay timer
	st     uint8    // sound timer
	opcode Opcode   // 2 bytes representing current opcode
}

const numRegisters = 16
const stackDepth = 16

// NewCPU returns a Chip-8 CPU with cleared registers, and initialized program
// counter.
func NewCPU() *CPU {
	return &CPU{
		v:      make([]uint8, numRegisters),
		i:      0,
		pc:     memoryProgramBegin,
		stack:  make([]uint16, stackDepth),
		sp:     0,
		dt:     0,
		st:     0,
		opcode: 0x0000,
	}
}

// Cycle spins the CPU, executing instructions from RAM.
func (c *Chip8) Cycle(memory *[]uint8) {
	// Load the 2-byte opcode
	bx := (*memory)[c.cpu.pc : c.cpu.pc+2]
	c.cpu.opcode = Opcode(binary.BigEndian.Uint16(bx))

	// Increment the program counter
	c.cpu.pc += 2

	// Execute the instruction
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
		case 0x3:
			c.cpu.Exec8XY3()
		default:
			log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
		}
	case 0xA000:
		c.cpu.ExecANNN()
	case 0xF000:
		switch c.cpu.opcode.nn() {
		case 0x55:
			c.cpu.ExecFX55(memory)
		case 0x65:
			c.cpu.ExecFX65(memory)
		case 0x1E:
			c.cpu.ExecFX1E()
		default:
			log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
		}
	default:
		log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
	}

	//fmt.Printf("CPU: %#v\n\n", *cpu)
}

// Chip-8 instructions found at:
// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#00E0

// 00E0 - CLS
// Clear the display.
func (cpu *CPU) Exec00E0(disp *[]uint8) {
	fmt.Printf("%#x: %#x CLS\n", cpu.pc-2, cpu.opcode)

	for i := range *disp {
		(*disp)[i] = 0
	}
}

// 00EE - RET
// Return from a subroutine.
func (cpu *CPU) Exec00EE() {
	fmt.Printf("%#x: %#x RET\n", cpu.pc-2, cpu.opcode)

	cpu.sp--
	cpu.pc = cpu.stack[cpu.sp]
}

// 1NNN - JP addr
// Set the program counter to NNN.
func (cpu *CPU) Exec1NNN() {
	nnn := cpu.opcode.nnn()

	fmt.Printf("%#x: %#x JP %#v\n", cpu.pc-2, cpu.opcode, nnn)

	cpu.pc = nnn
}

// 2NNN - CALL addr
// Call subroutine at NNN.
func (cpu *CPU) Exec2NNN() {
	nnn := cpu.opcode.nnn()

	fmt.Printf("%#x: %#x CALL %#v\n", cpu.pc-2, cpu.opcode, nnn)

	cpu.stack[cpu.sp] = cpu.pc
	cpu.sp++
	cpu.pc = nnn
}

// 3XNN - SE VX, byte
// Skip next instruction if VX == NN.
func (cpu *CPU) Exec3XNN() {
	x := cpu.opcode.x()
	nn := cpu.opcode.nn()

	fmt.Printf("%#x: %#x SE V%d, %#v\n", cpu.pc-2, cpu.opcode, x, nn)
	if cpu.v[x] == nn {
		cpu.pc += 2
	}
}

// 6XNN - LD VX, byte
// Load the value NN into register VX.
func (cpu *CPU) Exec6XNN() {
	x := cpu.opcode.x()
	nn := cpu.opcode.nn()

	fmt.Printf("%#x: %#x LD V%d, %#v\n", cpu.pc-2, cpu.opcode, x, nn)

	cpu.v[x] = nn
}

// 7XNN - ADD VX, byte
// Add the value NN to the value found in register VX, store the result in VX.
func (cpu *CPU) Exec7XNN() {
	x := cpu.opcode.x()
	nn := cpu.opcode.nn()

	fmt.Printf("%#x: %#x ADD V%d, %#v\n", cpu.pc-2, cpu.opcode, x, nn)

	cpu.v[x] += nn
}

// 8XY0 - LD VX, VY
// Store the value of register VY in register VX.
func (cpu *CPU) Exec8XY0() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x LD V%d, V%d\n", cpu.pc-2, cpu.opcode, x, y)

	cpu.v[x] = cpu.v[y]
}

// 8XY3 - XOR VX, VY
// Set VX to VX XOR VY.
func (cpu *CPU) Exec8XY3() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x XOR V%d, V%d\n", cpu.pc-2, cpu.opcode, x, y)

	cpu.v[x] = cpu.v[x] ^ cpu.v[y]
}

// ANNN - LD I, addr
// Store the value of nnn in register I.
func (cpu *CPU) ExecANNN() {
	nnn := cpu.opcode.nnn()

	fmt.Printf("%#x: %#x LD I, %#x\n", cpu.pc-2, cpu.opcode, nnn)

	cpu.i = nnn
}

// FX55 - LD [I], VX
// Store registers V0 through VX in memory starting at location I.
func (cpu *CPU) ExecFX55(memory *[]uint8) {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x LD [I], V%d\n", cpu.pc-2, cpu.opcode, x)

	for i := 0; i <= int(x); i++ {
		(*memory)[int(cpu.i)+i] = cpu.v[i]
	}
	cpu.i = cpu.i + uint16(x) + 1
}

// FX65 - LD VX, [I]
// Load values from memory starting at location I into registers V0 through VX.
func (cpu *CPU) ExecFX65(memory *[]uint8) {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x LD V%d, [I]\n", cpu.pc-2, cpu.opcode, x)

	for i := 0; i <= int(x); i++ {
		cpu.v[i] = (*memory)[int(cpu.i)+i]
	}
	cpu.i = cpu.i + uint16(x) + 1
}

// FX1E - ADD I, VX
// Add the values of I and VX, store the result in I.
func (cpu *CPU) ExecFX1E() {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x ADD I, V%d\n", cpu.pc-2, cpu.opcode, x)

	cpu.i = cpu.i + uint16(cpu.v[x])
}

type Opcode uint16

func (oc Opcode) nnn() uint16 {
	return uint16(oc & 0x0FFF)
}

func (oc Opcode) nn() uint8 {
	return uint8(oc & 0x00FF)
}

func (oc Opcode) n() uint8 {
	return uint8(oc & 0x000F)
}

func (oc Opcode) x() uint8 {
	return uint8((oc & 0x0F00) >> 8)
}

func (oc Opcode) y() uint8 {
	return uint8((oc & 0x00F0) >> 4)
}
