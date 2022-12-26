// Reference document: https://www.masswerk.at/6502/6502_instruction_set.html

package nes

type AddressMode uint8

const (
	modeNone AddressMode = iota
	modeAccu
	modeAbso
	modeAbsX
	modeAbsY
	modeImmd
	modeImpl
	modeIndi
	modeXInd
	modeIndY
	modeRela
	modeZpag
	modeZpgX
	modeZpgY
)

func (am AddressMode) ToString() string {
	switch am {
	case modeNone:
		return "---"
	case modeAccu:
		return "A"
	case modeAbso:
		return "abs"
	case modeAbsX:
		return "abs,X"
	case modeAbsY:
		return "abs,Y"
	case modeImmd:
		return "#"
	case modeImpl:
		return "impl"
	case modeIndi:
		return "ind"
	case modeXInd:
		return "X,ind"
	case modeIndY:
		return "ind,Y"
	case modeRela:
		return "rel"
	case modeZpag:
		return "zpg"
	case modeZpgX:
		return "zpg,X"
	case modeZpgY:
		return "zpg,Y"
	}
	return ""
}

type AddressInfo struct {
	mode    AddressMode
	address uint16
	crossed bool // whether the page boundaries were crossed
}

type AddressModeFunc func() AddressInfo

// IsCrossed returns true if the old and new 16-bit address belongs to different pages.
// The page of a given 16-bit address refers to the hi-byte.
func IsCrossed(old, new uint16) bool {
	return old&0xFF00 != new&0xFF00
}

func (cpu *CPU) None() AddressInfo {
	panic("Unsupported opcode!")
}

// Accu (Accumulator) - operand is AC (implied single byte instruction)
func (cpu *CPU) Accu() AddressInfo {
	return AddressInfo{
		mode: modeAccu,
	}
}

// Abso (absolute) - operand is address $HHLL
//
// 16-bit address words are little endian, lo(w)-byte first, followed by the hi(gh)-byte.
// (An assembler will use a human-readable, big-endian notation as in $HHLL.)
func (cpu *CPU) Abso() AddressInfo {
	addr := cpu.Read16(cpu.pc + 1)
	return AddressInfo{
		mode:    modeAbso,
		address: addr,
	}
}

// AbsX (absolute, X-indexed) - operand is address; effective address is address incremented by X with carry
func (cpu *CPU) AbsX() AddressInfo {
	baseAddr := cpu.Read16(cpu.pc + 1)
	addr := baseAddr + uint16(cpu.x)
	return AddressInfo{
		mode:    modeAbsX,
		address: addr,
		crossed: IsCrossed(baseAddr, addr),
	}
}

// AbsY (absolute, Y-indexed) - operand is address; effective address is address incremented by Y with carry
func (cpu *CPU) AbsY() AddressInfo {
	baseAddr := cpu.Read16(cpu.pc + 1)
	addr := baseAddr + uint16(cpu.y)
	return AddressInfo{
		mode:    modeAbsY,
		address: addr,
		crossed: IsCrossed(baseAddr, addr),
	}
}

// Immd (immediate) - operand is byte BB
func (cpu *CPU) Immd() AddressInfo {
	return AddressInfo{
		mode:    modeImmd,
		address: cpu.pc + 1,
	}
}

// Impl (implied) - operand implied
func (cpu *CPU) Impl() AddressInfo {
	return AddressInfo{
		mode: modeImpl,
	}
}

// Indi (indirect) - operand is address; effective address is contents of word at address: C.w($HHLL)
func (cpu *CPU) Indi() AddressInfo {
	var addr uint16
	pointer := cpu.Read16(cpu.pc + 1)
	// simulate the 6502 bug - if pointer is at page boundary, the hi-byte will actually not have its page incremented
	if pointer&0x00FF == 0x00FF {
		lo := uint16(cpu.Read(pointer))
		hi := uint16(cpu.Read(pointer & 0xFF00))
		addr = hi<<8 | lo
	} else {
		addr = cpu.Read16(pointer)
	}
	return AddressInfo{
		mode:    modeIndi,
		address: addr,
	}
}

// XInd (X-indexed, indirect) - operand is zero-page address; effective address is word in (LL + X, LL + X + 1), inc. without carry: C.w($00LL + X)
func (cpu *CPU) XInd() AddressInfo {
	var addr uint16
	baseAddr := cpu.Read(cpu.pc + 1)
	absAddr := uint16(baseAddr) + uint16(cpu.x)
	pointer := absAddr & 0x00FF
	// simulate the 6502 bug - if pointer is at page boundary, the hi-byte will actually not have its page incremented
	if pointer&0x00FF == 0x00FF {
		lo := uint16(cpu.Read(pointer))
		hi := uint16(cpu.Read(pointer & 0xFF00))
		addr = hi<<8 | lo
	} else {
		addr = cpu.Read16(pointer)
	}
	return AddressInfo{
		mode:    modeXInd,
		address: addr,
	}
}

// IndY (indirect, Y-indexed) - operand is zero-page address; effective address is word in (LL, LL + 1) incremented by Y with carry: C.w($00LL) + Y
func (cpu *CPU) IndY() AddressInfo {
	var baseAddr uint16
	pointer := uint16(cpu.Read(cpu.pc + 1))
	// simulate the 6502 bug - if pointer is at page boundary, the hi-byte will actually not have its page incremented
	if pointer&0x00FF == 0x00FF {
		lo := uint16(cpu.Read(pointer))
		hi := uint16(cpu.Read(pointer & 0xFF00))
		baseAddr = hi<<8 | lo
	} else {
		baseAddr = cpu.Read16(pointer)
	}
	addr := baseAddr + uint16(cpu.y)
	return AddressInfo{
		mode:    modeIndY,
		address: addr,
		crossed: IsCrossed(baseAddr, addr),
	}
}

// Rela (relative) - branch target is PC + signed offset BB
func (cpu *CPU) Rela() AddressInfo {
	var addr uint16
	offset := uint16(cpu.Read(cpu.pc + 1))
	baseAddr := cpu.pc + 2
	if offset < 0x80 {
		addr = baseAddr + offset
	} else {
		addr = baseAddr + offset - 0x100
	}
	return AddressInfo{
		mode:    modeRela,
		address: addr,
		crossed: IsCrossed(baseAddr, addr),
	}
}

// Zpag (zero-page) - operand is zero-page address (hi-byte is zero, address = $00LL)
func (cpu *CPU) Zpag() AddressInfo {
	addr := uint16(cpu.Read(cpu.pc + 1))
	return AddressInfo{
		mode:    modeZpag,
		address: addr,
	}
}

// ZpgX (zero-page, X-indexed) - operand is zero-page address; effective address is address incremented by X without carry
func (cpu *CPU) ZpgX() AddressInfo {
	addr := uint16(cpu.Read(cpu.pc+1) + cpu.x)
	return AddressInfo{
		mode:    modeZpgX,
		address: addr,
	}
}

// ZpgY (zero-page, Y-indexed) - operand is zero-page address; effective address is address incremented by Y without carry
func (cpu *CPU) ZpgY() AddressInfo {
	addr := uint16(cpu.Read(cpu.pc+1) + cpu.y)
	return AddressInfo{
		mode:    modeZpgY,
		address: addr,
	}
}
