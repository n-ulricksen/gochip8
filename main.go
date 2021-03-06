package main

import (
	"fmt"

	"github.com/n-ulricksen/chip8/core"
)

// The path to the ROM used to test our emulator.
var rompath string = "./roms/CONNECT4"

func main() {
	chip8 := core.NewChip8()

	fmt.Printf("Loading ROM from %s\n", rompath)
	chip8.LoadRom(rompath)

	fmt.Printf("Chip-8 machine: %#v\n", chip8)

	fmt.Println("Starting program")

	chip8.Run()

}
