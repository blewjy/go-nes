package nes

import "fmt"

type CPU struct {
	bus *Bus

	a  uint8  // Accumulator register
	x  uint8  // X register
	y  uint8  // Y register
	p  uint8  // Status register
	sp uint8  // Stack pointer
	pc uint16 // Program counter

	// Status registers
	c uint8
	z uint8
	i uint8
	d uint8
	b uint8
	u uint8
	v uint8
	n uint8

	// cycles???
	cycle int

	// emulated maps (must be init-ed)
	addressModeToAddressModeFunc map[AddressMode]AddressModeFunc
	instToInstFunc               map[Instruction]InstructionFunc
}

func NewCPU(bus *Bus) *CPU {
	cpu := &CPU{
		bus: bus,
	}
	cpu.addressModeToAddressModeFunc = map[AddressMode]AddressModeFunc{
		modeAccu: cpu.A, modeAbso: cpu.abs, modeAbsX: cpu.absX, modeAbsY: cpu.absY,
		modeImmd: cpu.immd, modeImpl: cpu.impl, modeIndi: cpu.ind, modeXInd: cpu.xInd,
		modeIndY: cpu.indY, modeRela: cpu.rel, modeZpag: cpu.zpg, modeZpgX: cpu.zpgX,
		modeZpgY: cpu.zpgY,
	}
	cpu.instToInstFunc = map[Instruction]InstructionFunc{
		ADC: cpu.adc, AND: cpu.and, ASL: cpu.asl, BCC: cpu.bcc, BCS: cpu.bcs, BEQ: cpu.beq, BIT: cpu.bit, BMI: cpu.bmi, BNE: cpu.bne, BPL: cpu.bpl, BRK: cpu.brk, BVC: cpu.bvc, BVS: cpu.bvs, CLC: cpu.clc,
		CLD: cpu.cld, CLI: cpu.cli, CLV: cpu.clv, CMP: cpu.cmp, CPX: cpu.cpx, CPY: cpu.cpy, DEC: cpu.dec, DEX: cpu.dex, DEY: cpu.dey, EOR: cpu.eor, INC: cpu.inc, INX: cpu.inx, INY: cpu.iny, JMP: cpu.jmp,
		JSR: cpu.jsr, LDA: cpu.lda, LDX: cpu.ldx, LDY: cpu.ldy, LSR: cpu.lsr, NOP: cpu.nop, ORA: cpu.ora, PHA: cpu.pha, PHP: cpu.php, PLA: cpu.pla, PLP: cpu.plp, ROL: cpu.rol, ROR: cpu.ror, RTI: cpu.rti,
		RTS: cpu.rts, SBC: cpu.sbc, SEC: cpu.sec, SED: cpu.sed, SEI: cpu.sei, STA: cpu.sta, STX: cpu.stx, STY: cpu.sty, TAX: cpu.tax, TAY: cpu.tay, TSX: cpu.tsx, TXA: cpu.txa, TXS: cpu.txs, TYA: cpu.tya,
	}
	return cpu
}

func (cpu *CPU) Reset() {
	// Set PC
	cpu.pc = cpu.Read16(0xFFFC)

	// Reset registers
	cpu.a = 0
	cpu.x = 0
	cpu.y = 0
	cpu.sp = 0xFD

	// Reset status register
	cpu.c = 0
	cpu.z = 0
	cpu.i = 1
	cpu.d = 0
	cpu.b = 0
	cpu.u = 1
	cpu.v = 0
	cpu.n = 0

	cpu.p = 0x24

	// Reset cycle
	cpu.cycle = 7
}

// Read will read 1 byte (8 bits) from the given address.
func (cpu *CPU) Read(addr uint16) uint8 {
	return cpu.bus.CpuRead(addr)
}

// Write will write 1 byte of data to the given address.
func (cpu *CPU) Write(addr uint16, data uint8) {
	cpu.bus.CpuWrite(addr, data)
}

// Read16 will read 2 bytes (16 bits) from the given address.
// 16-bit address words are little endian, lo(w)-byte first, followed by the hi(gh)-byte.
// (An assembler will use a human-readable, big-endian notation as in $HHLL)
func (cpu *CPU) Read16(addr uint16) uint16 {
	lo := uint16(cpu.Read(addr))
	hi := uint16(cpu.Read(addr + 1))
	return hi<<8 | lo
}

func (cpu *CPU) Write16(addr, data uint16) {
	hi := uint8(data >> 8)
	lo := uint8(data & 0xFF)
	cpu.Write(addr, lo)
	cpu.Write(addr+1, hi)
}

func (cpu *CPU) Push(data uint8) {
	//fmt.Printf("Pushing to 0x%02X to stack at 0x%04X\n", data, 0x100|uint16(cpu.sp))
	cpu.Write(0x100|uint16(cpu.sp), data)
	cpu.sp--
}

func (cpu *CPU) Push16(data uint16) {
	hi := uint8(data >> 8)
	lo := uint8(data & 0xFF)
	cpu.Push(hi)
	cpu.Push(lo)
}

func (cpu *CPU) Pull() uint8 {
	cpu.sp++
	data := cpu.Read(0x100 | uint16(cpu.sp))
	//fmt.Printf("Pulling data 0x%02X from stack at 0x%04X\n", data, 0x100|uint16(cpu.sp))
	return data
}

func (cpu *CPU) Pull16() uint16 {
	lo := uint16(cpu.Pull())
	hi := uint16(cpu.Pull())
	return hi<<8 | lo
}

func (cpu *CPU) GetStatus() uint8 {
	return cpu.p

	N := cpu.n << 7
	V := cpu.v << 6
	U := cpu.u << 5
	B := cpu.b << 4
	D := cpu.d << 3
	I := cpu.i << 2
	Z := cpu.z << 1
	C := cpu.c << 0
	return N | V | U | B | D | I | Z | C
}

func (cpu *CPU) PushStatus() {
	cpu.Push(cpu.GetStatus())
}

func (cpu *CPU) PullStatus() {
	status := cpu.Pull()
	cpu.n = status >> 7 & 0x1
	cpu.v = status >> 6 & 0x1
	cpu.u = 1
	cpu.b = 0
	cpu.d = status >> 3 & 0x1
	cpu.i = status >> 2 & 0x1
	cpu.z = status >> 1 & 0x1
	cpu.c = status >> 0 & 0x1

	cpu.p = status
	cpu.SetFlag(U, true)
	cpu.SetFlag(B, false)
}

func (cpu *CPU) Clock() {
	opcode := cpu.Read(cpu.pc)
	info := cpu.getInstructionInfo(opcode)

	addrInfo := info.addrModeFunc()

	cpu.pc += uint16(info.instSize)

	hasAdditionalCycles := info.instFunc(addrInfo.mode, addrInfo.address)
	if hasAdditionalCycles {
		cpu.cycle += cpu.GetAdditionalCycles(info, addrInfo)
	}

	cpu.cycle += int(info.instCycles)
	// todo: return number of cycles??
}

func (cpu *CPU) PeekCurrentSnapshot() string {
	result := ""

	opcode := cpu.Read(cpu.pc)
	info := cpu.getInstructionInfo(opcode)

	result += fmt.Sprintf("%04X, ", cpu.pc)
	for i := uint8(0); i < info.instSize; i++ {
		result += fmt.Sprintf("%02X ", cpu.Read(cpu.pc+uint16(i)))
	}
	result += fmt.Sprintf("\t\tA: %02X X: %02X Y: %02X P: %02X SP: %02X", cpu.a, cpu.x, cpu.y, cpu.GetStatus(), cpu.sp)

	result += fmt.Sprint("\tCYC: ", cpu.cycle)

	return result
}

func (cpu *CPU) GetAdditionalCycles(info InstructionInfo, addrInfo AddressInfo) int {
	if info.inst.IsBranch() {
		if addrInfo.crossed {
			return 2
		} else {
			return 1
		}
	} else {
		if addrInfo.crossed {
			return 1
		} else {
			return 0
		}
	}
}
