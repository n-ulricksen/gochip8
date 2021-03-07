package core

type Opcode uint16

// nnn returns the 12 lowest bits of the opcode.
func (oc Opcode) nnn() uint16 {
	return uint16(oc & 0x0FFF)
}

// nn returns the 8 lowest bits of the opcode.
func (oc Opcode) nn() uint8 {
	return uint8(oc & 0x00FF)
}

// n returns the 4 lowest bits of the opcode.
func (oc Opcode) n() uint8 {
	return uint8(oc & 0x000F)
}

// x returns the lower 4 bits of the high byte of the instruction
func (oc Opcode) x() uint8 {
	return uint8((oc & 0x0F00) >> 8)
}

// y returns the upper 4 bits of the low byte of the instruction
func (oc Opcode) y() uint8 {
	return uint8((oc & 0x00F0) >> 4)
}
