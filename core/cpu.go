package core

import (
	"fmt"
	"math/rand"
	"time"
)

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

const (
	numRegisters = 16
	stackDepth   = 16
)

// NewCPU returns a Chip-8 CPU with cleared registers, and initialized program
// counter.
func NewCPU() *CPU {
	return &CPU{
		v:      make([]uint8, numRegisters),
		i:      0,
		pc:     programEntryOffset,
		stack:  make([]uint16, stackDepth),
		sp:     0,
		dt:     0,
		st:     0,
		opcode: 0x0000,
	}
}

func init() {
	// "random" numbers needed by 0xCXNN instruction.
	rand.Seed(time.Now().UnixNano())
}

func (cpu *CPU) decrementTimers() {
	if cpu.dt > 0 {
		cpu.dt--
	}
	if cpu.st > 0 {
		cpu.st--
	}
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

// 4XNN - SNE VX, byte
// Skip next instruction if VX != NN.
func (cpu *CPU) Exec4XNN() {
	x := cpu.opcode.x()
	nn := cpu.opcode.nn()

	fmt.Printf("%#x: %#x SNE V%d, %#v\n", cpu.pc-2, cpu.opcode, x, nn)

	if cpu.v[x] != nn {
		cpu.pc += 2
	}
}

// 5XY0 - SE VX, VY
// Skip next instruction if VX == VY.
func (cpu *CPU) Exec5XY0() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x SE V%d, V%d\n", cpu.pc-2, cpu.opcode, x, y)

	if cpu.v[x] == cpu.v[y] {
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

// 8XY2 - AND VX, VY
// Store the result of VX AND VY to register VX.
func (cpu *CPU) Exec8XY2() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x AND V%d, V%d\n", cpu.pc-2, cpu.opcode, x, y)

	cpu.v[x] = cpu.v[x] & cpu.v[y]
}

// 8XY3 - XOR VX, VY
// Set VX to VX XOR VY.
func (cpu *CPU) Exec8XY3() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x XOR V%d, V%d\n", cpu.pc-2, cpu.opcode, x, y)

	cpu.v[x] = cpu.v[x] ^ cpu.v[y]
}

// 8XY4 - ADD VX, VY
// Set VX to result of VX + VY. Set VF = 1 if carry (result > 255).
func (cpu *CPU) Exec8XY4() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x ADD V%d, V%d\n", cpu.pc-2, cpu.opcode, x, y)

	sum16 := uint16(cpu.v[x]) + uint16(cpu.v[y])
	if sum16 > 0xFF {
		cpu.v[0xF] = 1
	} else {
		cpu.v[0xF] = 0
	}

	cpu.v[x] = uint8(sum16)
}

// 8XY5 - SUB VX, VY
// Set VX to result of VX - VY. Set VF = 0 if borrow, else VF = 1.
func (cpu *CPU) Exec8XY5() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x SUB V%d, V%d\n", cpu.pc-2, cpu.opcode, x, y)

	if cpu.v[y] > cpu.v[x] {
		cpu.v[0xF] = 0
	} else {
		cpu.v[0xF] = 0
	}

	cpu.v[x] -= cpu.v[y]
}

// 8XY6 - SHR VX {, VY}
// Store the value of VY shifted right one bit in register VX. Set register VF to
// the least significant bit prior to shift.
func (cpu *CPU) Exec8XY6() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x SHR V%d {, V%d}\n", cpu.pc-2, cpu.opcode, x, y)

	// Set carry flag if needed.
	if cpu.v[x]%2 == 1 {
		cpu.v[0xF] = 1
	} else {
		cpu.v[0xF] = 0
	}
	cpu.v[x] = cpu.v[y] >> 1
}

// 8XYE - SHL VX {, VY}
// Store the value of VY shifted left one bit in register VX. Set register VF to
// the most significant bit prior to shift.
func (cpu *CPU) Exec8XYE() {
	x := cpu.opcode.x()
	y := cpu.opcode.y()

	fmt.Printf("%#x: %#x SHL V%d {, V%d}\n", cpu.pc-2, cpu.opcode, x, y)

	// Set carry flag if needed.
	if cpu.v[x] >= 128 {
		cpu.v[0xF] = 1
	} else {
		cpu.v[0xF] = 0
	}
	cpu.v[x] = cpu.v[y] << 1
}

// ANNN - LD I, addr
// Store the value of nnn in register I.
func (cpu *CPU) ExecANNN() {
	nnn := cpu.opcode.nnn()

	fmt.Printf("%#x: %#x LD I, %#x\n", cpu.pc-2, cpu.opcode, nnn)

	cpu.i = nnn
}

// CXNN - RND VX, byte
// Set VX to the result of (rand(0-255) AND NN)
func (cpu *CPU) ExecCXNN() {
	x := cpu.opcode.x()
	nn := cpu.opcode.nn()

	fmt.Printf("%#x: %#x RND V%d, byte\n", cpu.pc-2, cpu.opcode, x)

	// set v[x] to (rand(0xFF) & NN)
	cpu.v[x] = uint8(rand.Intn(256)) & nn
}

// DXYN - DRW VX, VY, nibble
// Display an n-byte sprite starting at memory location I, at display location
// (VX, VY). Set VF if collision occurs. Sprites are XORed into the existing
// display.
func (cpu *CPU) ExecDXYN(memory *[]uint8, display *[]uint8) {
	x := cpu.opcode.x()
	y := cpu.opcode.y()
	n := cpu.opcode.n()

	fmt.Printf("%#x: %#x DRW V%d, V%d, %#x\n", cpu.pc-2, cpu.opcode, x, y, n)

	spriteMem := (*memory)[cpu.i:]

	for iy := uint8(0); iy < n; iy++ {
		for ix := uint8(0); ix < 8; ix++ {
			xpos := int(cpu.v[x]) + int(ix)
			ypos := int(cpu.v[y]) + int(iy)
			if xpos >= Chip8Width || ypos >= Chip8Height {
				continue
			}

			// XOR sprite to the display.
			oldpixel := (*display)[ypos*Chip8Width+xpos]
			newpixel := (spriteMem[iy] >> (7 - ix)) & 0x01
			(*display)[ypos*Chip8Width+xpos] ^= newpixel

			// Set carry flag if any pixels are changed to unset.
			flipped := uint8(0)
			if oldpixel == 1 && newpixel == 1 {
				flipped = 1
			}
			cpu.v[0xF] = flipped
		}
	}
}

// EX9E - SKP VX
// Skip next instruction if VX key is pressed.
func (cpu *CPU) ExecEX9E(keys []uint8) {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x SKP V%d\n", cpu.pc-2, cpu.opcode, x)

	if keys[int(cpu.v[x])] == 1 {
		cpu.pc += 2
	}
}

// EXA1 - SKNP VX
// Skip next instruction if VX key is not pressed.
func (cpu *CPU) ExecEXA1(keys []uint8) {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x SKNP V%d\n", cpu.pc-2, cpu.opcode, x)

	if keys[int(cpu.v[x])] == 0 {
		cpu.pc += 2
	}
}

// FX07 - LD VX, DT
// Set VX to the value of the delay timer.
func (cpu *CPU) ExecFX07() {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x LD V%d, DT\n", cpu.pc-2, cpu.opcode, x)

	cpu.v[x] = cpu.dt
}

// FX0A - LD VX, key
// Wait for a key press, store the value of the key in VX.
func (cpu *CPU) ExecFX0A(keys []uint8) {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x LD V%d, key\n", cpu.pc-2, cpu.opcode, x)

	var pressed uint8 = 0xFF
	for i, keystate := range keys {
		if keystate == 1 {
			pressed = uint8(i)
			break
		}
	}

	if pressed == 0xFF {
		// Block by decrementing the program counter.
		cpu.pc -= 2
	} else {
		cpu.v[x] = pressed
	}
}

// FX15 - LD DT, VX
// Set the delay timer to the value of VX.
func (cpu *CPU) ExecFX15() {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x LD DT, V%d\n", cpu.pc-2, cpu.opcode, x)

	cpu.dt = cpu.v[x]
}

// FX1E - ADD I, VX
// Add the values of I and VX, store the result in I.
func (cpu *CPU) ExecFX1E() {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x ADD I, V%d\n", cpu.pc-2, cpu.opcode, x)

	cpu.i = cpu.i + uint16(cpu.v[x])
}

// FX29 - LD F, VX
// Set I to the location of the sprite data corresponding to value of VX.
func (cpu *CPU) ExecFX29(memory *[]uint8) {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x LD F, V%d\n", cpu.pc-2, cpu.opcode, x)

	cpu.i = characterSpritesOffset + uint16(cpu.v[x])*characterSpriteBytes
}

// FX33 - LD B, VX
// Store the binary representation of VX in memory at I, I+1, I+2 (hunreds, tens, ones).
func (cpu *CPU) ExecFX33(memory *[]uint8) {
	x := cpu.opcode.x()

	fmt.Printf("%#x: %#x LD B, V%d\n", cpu.pc-2, cpu.opcode, x)

	(*memory)[cpu.i] = cpu.v[x] / 100
	(*memory)[cpu.i+1] = (cpu.v[x] % 100) / 10
	(*memory)[cpu.i+2] = cpu.v[x] % 10
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
