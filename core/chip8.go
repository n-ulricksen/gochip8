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
		fmt.Printf("Loading %v into ROM\n", data)
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
		v:  make([]uint8, 8),
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
	opcode := binary.BigEndian.Uint16(bx)
	fmt.Printf("opcode: %#x\n", opcode)

	// Increment the program counter
	cpu.pc += 2

	// Decode the opcode
	op := decodeOp(opcode)

	fmt.Printf("op: %#x\n", op)
	fmt.Println()

	// Execute the instruction
	switch op {
	case opLd:
		cpu.ExecLd()
	case opAdd:
		cpu.ExecAdd()
	case opLdr:
		cpu.ExecLdr()
	}
}

type Op uint16

const (
	opLd  Op = 0x6000
	opAdd Op = 0x7000
	opLdr Op = 0x8000
)

// decodeOp verifies and returns the given 2-bytes as a valid Chip-8 opcode.
func decodeOp(opcode uint16) Op {
	op := opcode & 0xF000

	switch Op(opcode) {
	case opLd:
	case opAdd:
	case opLdr:
	default:
		log.Fatalf("Invalid opcode: %#v\n", opcode)
	}

	return Op(op)
}

func decodeOpNNN(opcode uint16) uint16 {
	return opcode & 0x0FFF
}

func decodeOpNN(opcode uint16) uint8 {
	return uint8(opcode & 0x00FF)
}

func decodeOpN(opcode uint16) uint8 {
	return uint8(opcode & 0x000F)
}

func decodeOpX(opcode uint16) uint8 {
	return uint8((opcode & 0x0F00) >> 8)
}

func decodeOpY(opcode uint16) uint8 {
	return uint8((opcode & 0x00F0) >> 4)
}
