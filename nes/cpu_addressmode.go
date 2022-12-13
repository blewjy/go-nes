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

type AddressInfo struct {
	mode    AddressMode
	address uint16
	crossed bool // whether the page boundaries were crossed
}

type AddressModeFunc func() AddressInfo

var (
	addressModeToAddressModeFunc = map[AddressMode]AddressModeFunc{}
)

// IsCrossed returns true if the old and new 16-bit address belongs to different pages.
// The page of a given 16-bit address refers to the hi-byte.
func IsCrossed(old, new uint16) bool {
	return old&0xFF00 != new&0xFF00
}

// A (Accumulator) - operand is AC (implied single byte instruction)
func (cpu *CPU) A() AddressInfo {
	return AddressInfo{
		mode: modeAccu,
	}
}

// abs (absolute) - operand is address $HHLL
//
// 16-bit address words are little endian, lo(w)-byte first, followed by the hi(gh)-byte.
// (An assembler will use a human-readable, big-endian notation as in $HHLL.)
func (cpu *CPU) abs() AddressInfo {
	addr := cpu.Read16(cpu.pc + 1)
	return AddressInfo{
		mode:    modeAbso,
		address: addr,
	}
}

// absX (absolute, X-indexed) - operand is address; effective address is address incremented by X with carry
func (cpu *CPU) absX() AddressInfo {
	baseAddr := cpu.Read16(cpu.pc + 1)
	addr := baseAddr + uint16(cpu.x)
	return AddressInfo{
		mode:    modeAbsX,
		address: addr,
		crossed: IsCrossed(baseAddr, addr),
	}
}

// absY (absolute, Y-indexed) - operand is address; effective address is address incremented by Y with carry
func (cpu *CPU) absY() AddressInfo {
	baseAddr := cpu.Read16(cpu.pc + 1)
	addr := baseAddr + uint16(cpu.y)
	return AddressInfo{
		mode:    modeAbsY,
		address: addr,
		crossed: IsCrossed(baseAddr, addr),
	}
}

// immd (immediate) - operand is byte BB
func (cpu *CPU) immd() AddressInfo {
	return AddressInfo{
		mode:    modeImmd,
		address: cpu.pc + 1,
	}
}

// impl (implied) - operand implied
func (cpu *CPU) impl() AddressInfo {
	return AddressInfo{
		mode: modeImpl,
	}
}

// ind (indirect) - operand is address; effective address is contents of word at address: C.w($HHLL)
func (cpu *CPU) ind() AddressInfo {
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

// xInd (X-indexed, indirect) - operand is zero-page address; effective address is word in (LL + X, LL + X + 1), inc. without carry: C.w($00LL + X)
func (cpu *CPU) xInd() AddressInfo {
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

// indY (indirect, Y-indexed) - operand is zero-page address; effective address is word in (LL, LL + 1) incremented by Y with carry: C.w($00LL) + Y
func (cpu *CPU) indY() AddressInfo {
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

// rel (relative) - branch target is PC + signed offset BB
func (cpu *CPU) rel() AddressInfo {
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

// zpg (zero-page) - operand is zero-page address (hi-byte is zero, address = $00LL)
func (cpu *CPU) zpg() AddressInfo {
	addr := uint16(cpu.Read(cpu.pc + 1))
	return AddressInfo{
		mode:    modeZpag,
		address: addr,
	}
}

// zpgX (zero-page, X-indexed) - operand is zero-page address; effective address is address incremented by X without carry
func (cpu *CPU) zpgX() AddressInfo {
	addr := uint16(cpu.Read(cpu.pc+1) + cpu.x)
	return AddressInfo{
		mode:    modeZpgX,
		address: addr,
	}
}

// zpgX (zero-page, Y-indexed) - operand is zero-page address; effective address is address incremented by Y without carry
func (cpu *CPU) zpgY() AddressInfo {
	addr := uint16(cpu.Read(cpu.pc+1) + cpu.y)
	return AddressInfo{
		mode:    modeZpgY,
		address: addr,
	}
}
