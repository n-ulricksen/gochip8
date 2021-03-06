package core

import (
	"fmt"
	"io/ioutil"
	"log"
)

const MEMORY_SIZE uint16 = 4096
const MEMORY_PROGRAM_BEGIN uint16 = 0x200

// The Chip8 emulator
type Chip8 struct {
	mem []uint8
}

// NewChip8 creates a new Chip8 emulator with 4KB RAM.
func NewChip8() *Chip8 {
	return &Chip8{
		mem: make([]uint8, MEMORY_SIZE),
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
		c.mem[int(MEMORY_PROGRAM_BEGIN)+i] = data
	}

}
