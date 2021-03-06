package core

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
)

const memorySize uint16 = 4096
const memoryProgramBegin uint16 = 0x200

// The Chip8 emulator
type Chip8 struct {
	mem []byte
	cpu *CPU
}

// NewChip8 creates a new Chip8 emulator with 4KB RAM.
func NewChip8() *Chip8 {
	return &Chip8{
		mem: make([]byte, memorySize),
		cpu: NewCPU(),
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
	for i := 0; i < 50; i++ {
		c.cpu.Cycle(&c.mem)
	}
}

// CPU used by the Chip-8 emulator
type CPU struct {
	v  []uint8 // V registers - general purpose
	i  uint16  // I register - general purpose
	pc uint16  // program counter
	sp uint8   // stack pointer
	dt uint8   // delay timer
	st uint8   // sound timer
}

// NewCPU returns a Chip-8 CPU with cleared registers, and initialized program
// counter.
func NewCPU() *CPU {
	return &CPU{
		v:  make([]uint8, 16),
		i:  0,
		pc: memoryProgramBegin,
		sp: 0,
		dt: 0,
		st: 0,
	}
}

// Cycle spins the CPU, executing instructions from RAM.
func (cpu *CPU) Cycle(memory *[]uint8) {
	// Load the 2-byte opcode
	bx := (*memory)[cpu.pc : cpu.pc+2]
	opcode := Opcode(binary.BigEndian.Uint16(bx))

	// Increment the program counter
	cpu.pc += 2

	// Decode the opcode
	op := opcode.op()

	// Execute the instruction
	switch op {
	case op1NNN:
		cpu.Exec1NNN(opcode)
	case op6XNN:
		cpu.Exec6XNN(opcode)
	case op7XNN:
		cpu.Exec7XNN(opcode)
	case op8XY0:
		cpu.Exec8XY0(opcode)
	case opANNN:
		cpu.ExecANNN(opcode)
	case opFXmm:
		mm := opcode.nn()
		switch mm {
		case 0x55:
			cpu.ExecFX55(opcode, memory)
		case 0x65:
			cpu.ExecFX65(opcode, memory)
		default:
			log.Fatalf("Invalid opcode: %#v\n", opcode)
		}
	default:
		log.Fatalf("Invalid opcode: %#v\n", opcode)
	}

	fmt.Printf("CPU: %#v\n\n", *cpu)
}

// Chip-8 instructions found at:
// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#00E0

// 1NNN - JP addr
// Set the program counter to NNN.
func (cpu *CPU) Exec1NNN(opcode Opcode) {
	nnn := opcode.nnn()

	fmt.Printf("%#x: %#x JP %#v\n", cpu.pc-2, opcode, nnn)

	cpu.pc = nnn
}

// 6XNN - LD VX, byte
// Load the value NN into register VX.
func (cpu *CPU) Exec6XNN(opcode Opcode) {
	x := opcode.x()
	nn := opcode.nn()

	fmt.Printf("%#x: %#x LD V%d, %#v\n", cpu.pc-2, opcode, x, nn)

	cpu.v[x] = nn
}

// 7XNN - ADD VX, byte
// Add the value NN to the value found in register VX, store the result in VX.
func (cpu *CPU) Exec7XNN(opcode Opcode) {
	x := opcode.x()
	nn := opcode.nn()

	fmt.Printf("%#x: %#x ADD V%d, %#v\n", cpu.pc-2, opcode, x, nn)

	cpu.v[x] += nn
}

// 8XY0 - LD VX, VY
// Store the value of register VY in register VX.
func (cpu *CPU) Exec8XY0(opcode Opcode) {
	x := opcode.x()
	y := opcode.y()

	fmt.Printf("%#x: %#x LD V%d, V%d\n", cpu.pc-2, opcode, x, y)

	cpu.v[x] = cpu.v[y]
}

// ANNN - LD I, addr
// Store the value of nnn in register I.
func (cpu *CPU) ExecANNN(opcode Opcode) {
	nnn := opcode.nnn()

	fmt.Printf("%#x: %#x LD I, %#x\n", cpu.pc-2, opcode, nnn)

	cpu.i = nnn
}

// FX55 - LD [I], VX
// Store registers V0 through VX in memory starting at location I.
func (cpu *CPU) ExecFX55(opcode Opcode, memory *[]uint8) {
	x := opcode.x()

	fmt.Printf("%#x: %#x LD [I], V%d\n", cpu.pc-2, opcode, x)
	for i := 0; i <= int(x); i++ {
		(*memory)[int(cpu.i)+i] = cpu.v[i]
	}
	cpu.i = cpu.i + uint16(x) + 1
}

// FX65 - LD VX, [I]
// Load values from memory starting at location I into registers V0 through VX.
func (cpu *CPU) ExecFX65(opcode Opcode, memory *[]uint8) {
	x := opcode.x()

	fmt.Printf("%#x: %#x LD V%d, [I]\n", cpu.pc-2, opcode, x)

	for i := 0; i <= int(x); i++ {
		cpu.v[i] = (*memory)[int(cpu.i)+i]
	}
	cpu.i = cpu.i + uint16(x) + 1
}

type Opcode uint16

// Available Chip-8 operations
const (
	op1NNN Opcode = 0x1000
	op6XNN Opcode = 0x6000
	op7XNN Opcode = 0x7000
	op8XY0 Opcode = 0x8000
	opANNN Opcode = 0xA000
	opFXmm Opcode = 0xF000
)

// decodeOp verifies and returns the given 2-bytes as a valid Chip-8 opcode.
func (oc Opcode) op() Opcode {
	return oc & 0xF000
}

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
