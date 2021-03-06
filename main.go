package main

import (
	"fmt"

	"github.com/n-ulricksen/chip8/core"
)

// The path to the ROM used to test our emulator.
var rompath string = "./roms/BLINKY"

func main() {
	chip8 := core.NewChip8()

	fmt.Printf("Loading ROM from %s\n", rompath)
	chip8.LoadRom(rompath)

	fmt.Println("Starting program...")
	fmt.Println()

	chip8.Run()

}
