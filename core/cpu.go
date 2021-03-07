package core

import "fmt"

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
		pc:     memoryProgramBegin,
		stack:  make([]uint16, stackDepth),
		sp:     0,
		dt:     0,
		st:     0,
		opcode: 0x0000,
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