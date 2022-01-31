package main

import (
	"flag"
	"fmt"

	"github.com/n-ulricksen/chip8/core"
)

// The path to the ROM used to test our emulator.
var (
	testpath string = "./test/BC_test.ch8"
)

var (
	flagtest  bool
	flagdebug bool
	rompath   string
)

func init() {
	flag.BoolVar(&flagtest, "t", false, "Load the emulator test ROM")
	flag.BoolVar(&flagdebug, "d", false, "Print debug info to the screen")
	flag.StringVar(&rompath, "p", "./roms/TETRIS", "Specify the path of the ROM to load")
	flag.Parse()
}

func main() {
	chip8 := core.NewChip8(flagdebug)

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
