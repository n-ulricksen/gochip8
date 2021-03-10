package main

import (
	"flag"
	"fmt"

	"github.com/n-ulricksen/chip8/core"
)

// The path to the ROM used to test our emulator.
var (
	rompath  string = "./roms/CONNECT4"
	testpath string = "./test/BC_test.ch8"
)

var flagtest bool

func init() {
	flag.BoolVar(&flagtest, "t", false, "Load the emulator test ROM")
	flag.Parse()
}

func main() {
	chip8 := core.NewChip8()

	if flagtest {
		fmt.Printf("Loading test ROM from %s\n", testpath)
		chip8.LoadRom(testpath)
	} else {
		fmt.Printf("Loading ROM from %s\n", rompath)
		chip8.LoadRom(rompath)
	}

	fmt.Println("Starting program...")
	fmt.Println()

	chip8.Run()
}
