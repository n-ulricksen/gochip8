package core

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	memorySize             uint16 = 4096
	programEntryOffset     uint16 = 0x200
	characterSpritesOffset uint16 = 0x100
	characterSpriteBytes          = 5
	chip8frequency                = 60 * 8
	fontpath                      = "./fonts/DotGothic16-Regular.ttf"
	fontsize                      = 32
)

// The Chip8 emulator
type Chip8 struct {
	mem       []byte // RAM
	cpu       *CPU
	display   []uint8 // emulator display
	keys      []uint8 // current state of each key
	renderer  *sdl.Renderer
	font      *ttf.Font
	isRunning bool
	isDebug   bool
	ophistory []string // history of cpu ops: `address: op, mneumonic`
	opindex   int      // insertion point in ophistory for next op
}

// NewChip8 creates a new Chip8 emulator with 4KB RAM.
func NewChip8(debug bool) *Chip8 {
	w, h := Chip8Width, Chip8Height

	// Initialize memory.
	memory := make([]byte, memorySize)
	copy(memory[characterSpritesOffset:], characterSprites)

	// Initialize SDL.
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatal("Unable to initialize SDL\n", err)
	}
	if err := ttf.Init(); err != nil {
		log.Fatal("Unable to initialize TTF\n", err)
	}

	// Load font.
	font, err := ttf.OpenFont(fontpath, fontsize)
	if err != nil {
		log.Fatal("Unable to load font\n", err)
	}

	return &Chip8{
		mem:       memory,
		cpu:       NewCPU(),
		display:   make([]uint8, w*h),
		keys:      make([]uint8, 16),
		renderer:  NewDisplayRenderer(debug),
		font:      font,
		isRunning: true,
		isDebug:   debug,
		ophistory: make([]string, 100),
		opindex:   0,
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
	defer c.font.Close()
	defer ttf.Quit()
	defer sdl.Quit()

	lastDrawTime := time.Now()
	vBlankTime := chip8frequency / VBlankFreq
	cycles := 0

	for c.isRunning {
		cycles++

		c.cycle()

		if cycles > vBlankTime {
			cycles = 0
			c.renderDisplay()

			// delay every 8 cycles to keep CPU steady
			elapsed := time.Now().Sub(lastDrawTime)
			timePerCycles := (time.Duration(vBlankTime) * time.Second / time.Duration(chip8frequency))
			time.Sleep(timePerCycles - elapsed)
			lastDrawTime = time.Now()

			c.cpu.decrementTimers()
		}

		c.pollSdlEvents()
	}
}

// renderDisplay presents the current display to the screen via the SDL2 renderer.
func (c *Chip8) renderDisplay() {
	c.renderer.SetDrawColor(0, 0, 0, 255)
	c.renderer.Clear()

	c.renderer.SetDrawColor(0, 255, 200, 255)

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

	if c.isDebug {
		c.renderDebugDisplay()
	}

	c.renderer.Present()
}

func (c *Chip8) renderDebugDisplay() {
	c.renderer.SetDrawColor(50, 50, 50, 255)
	debugRect := &sdl.Rect{X: 0, Y: EmulatorHeight, W: EmulatorWidth, H: DebugHeight}
	c.renderer.FillRect(debugRect)

	drawcolor := sdl.Color{R: 255, G: 0, B: 180, A: 255}
	surface, err := c.font.RenderUTF8Solid("testing123: hello world", drawcolor)
	if err != nil {
		log.Fatal(err)
	}
	defer surface.Free()

	texture, err := c.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		log.Fatal(err)
	}
	defer texture.Destroy()

	x := int32(0)
	y := int32(EmulatorHeight + 10)
	w := surface.W
	h := surface.H
	c.renderer.Copy(texture, nil, &sdl.Rect{X: x, Y: y, W: w, H: h})
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

// addOpHistoryItem adds an operation string to the Chip-8 ophistory slice at
// at the appropriate index.
func (c *Chip8) addOpHistoryItem(op string) {
	fmt.Print(op)
	c.ophistory[c.opindex] = op
	c.opindex = (c.opindex + 1) % len(c.ophistory)
}

// executeInstruction executes the appropriate instruction based on the opcode
// currently loaded into the CPU.
func (c *Chip8) executeInstruction() {
	var op string

	x := c.cpu.opcode.x()
	y := c.cpu.opcode.y()
	n := c.cpu.opcode.n()
	nn := c.cpu.opcode.nn()
	nnn := c.cpu.opcode.nnn()

	switch c.cpu.opcode & 0xF000 {
	case 0x0000:
		switch nnn {
		case 0x0E0:
			op = fmt.Sprintf("%#x: %#x CLS\n", c.cpu.pc-2, c.cpu.opcode)
			c.cpu.Exec00E0(&c.display)
		case 0x0EE:
			op = fmt.Sprintf("%#x: %#x RET\n", c.cpu.pc-2, c.cpu.opcode)
			c.cpu.Exec00EE()
		default:
			c.invalidOpcode()
		}
	case 0x1000:
		op = fmt.Sprintf("%#x: %#x JP %#v\n", c.cpu.pc-2, c.cpu.opcode, nnn)
		c.cpu.Exec1NNN()
	case 0x2000:
		op = fmt.Sprintf("%#x: %#x CALL %#v\n", c.cpu.pc-2, c.cpu.opcode, nnn)
		c.cpu.Exec2NNN()
	case 0x3000:
		op = fmt.Sprintf("%#x: %#x SE V%d, %#v\n", c.cpu.pc-2, c.cpu.opcode, x, nn)
		c.cpu.Exec3XNN()
	case 0x4000:
		op = fmt.Sprintf("%#x: %#x SNE V%d, %#v\n", c.cpu.pc-2, c.cpu.opcode, x, nn)
		c.cpu.Exec4XNN()
	case 0x5000:
		op = fmt.Sprintf("%#x: %#x SE V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
		c.cpu.Exec5XY0()
	case 0x6000:
		op = fmt.Sprintf("%#x: %#x LD V%d, %#v\n", c.cpu.pc-2, c.cpu.opcode, x, nn)
		c.cpu.Exec6XNN()
	case 0x7000:
		op = fmt.Sprintf("%#x: %#x ADD V%d, %#v\n", c.cpu.pc-2, c.cpu.opcode, x, nn)
		c.cpu.Exec7XNN()
	case 0x8000:
		switch n {
		case 0x0:
			op = fmt.Sprintf("%#x: %#x LD V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XY0()
		case 0x1:
			op = fmt.Sprintf("%#x: %#x OR V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XY1()
		case 0x2:
			op = fmt.Sprintf("%#x: %#x AND V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XY2()
		case 0x3:
			op = fmt.Sprintf("%#x: %#x XOR V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XY3()
		case 0x4:
			op = fmt.Sprintf("%#x: %#x ADD V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XY4()
		case 0x5:
			op = fmt.Sprintf("%#x: %#x SUB V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XY5()
		case 0x6:
			op = fmt.Sprintf("%#x: %#x SHR V%d {, V%d}\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XY6()
		case 0x7:
			op = fmt.Sprintf("%#x: %#x SUBN V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XY7()
		case 0xE:
			op = fmt.Sprintf("%#x: %#x SHL V%d {, V%d}\n", c.cpu.pc-2, c.cpu.opcode, x, y)
			c.cpu.Exec8XYE()
		default:
			c.invalidOpcode()
		}
	case 0x9000:
		op = fmt.Sprintf("%#x: %#x SNE V%d, V%d\n", c.cpu.pc-2, c.cpu.opcode, x, y)
		c.cpu.Exec9XY0()
	case 0xA000:
		op = fmt.Sprintf("%#x: %#x LD I, %#x\n", c.cpu.pc-2, c.cpu.opcode, nnn)
		c.cpu.ExecANNN()
	case 0xC000:
		op = fmt.Sprintf("%#x: %#x RND V%d, byte\n", c.cpu.pc-2, c.cpu.opcode, x)
		c.cpu.ExecCXNN()
	case 0xD000:
		op = fmt.Sprintf("%#x: %#x DRW V%d, V%d, %#x\n", c.cpu.pc-2, c.cpu.opcode, x, y, n)
		c.cpu.ExecDXYN(&c.mem, &c.display)
	case 0xE000:
		switch nn {
		case 0x9E:
			op = fmt.Sprintf("%#x: %#x SKP V%d\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecEX9E(c.keys)
		case 0xA1:
			op = fmt.Sprintf("%#x: %#x SKNP V%d\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecEXA1(c.keys)
		default:
			c.invalidOpcode()
		}
	case 0xF000:
		switch nn {
		case 0x07:
			op = fmt.Sprintf("%#x: %#x LD V%d, DT\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX07()
		case 0x0A:
			op = fmt.Sprintf("%#x: %#x LD V%d, key\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX0A(c.keys)
		case 0x15:
			op = fmt.Sprintf("%#x: %#x LD DT, V%d\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX15()
		case 0x18:
			op = fmt.Sprintf("%#x: %#x LD ST, V%d\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX18()
		case 0x1E:
			op = fmt.Sprintf("%#x: %#x ADD I, V%d\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX1E()
		case 0x29:
			op = fmt.Sprintf("%#x: %#x LD F, V%d\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX29(&c.mem)
		case 0x33:
			op = fmt.Sprintf("%#x: %#x LD B, V%d\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX33(&c.mem)
		case 0x55:
			op = fmt.Sprintf("%#x: %#x LD [I], V%d\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX55(&c.mem)
		case 0x65:
			op = fmt.Sprintf("%#x: %#x LD V%d, [I]\n", c.cpu.pc-2, c.cpu.opcode, x)
			c.cpu.ExecFX65(&c.mem)
		default:
			c.invalidOpcode()
		}
	default:
		c.invalidOpcode()
	}

	c.addOpHistoryItem(op)
}

// invalidOpcode prints an error message displaying the invalid opcode held in
// the cpu, and then terminates execution of the emulator.
func (c *Chip8) invalidOpcode() {
	log.Fatalf("Invalid opcode: %#v\n", c.cpu.opcode)
}
