package nes

import "fmt"

type CPU struct {
	bus *Bus

	// Registers
	a uint8 // Accumulator register
	x uint8 // X register
	y uint8 // Y register

	stackPtr uint8  // Stack pointer
	pc       uint16 // Program counter

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
		modeAccu: cpu.A,
		modeAbso: cpu.abs,
		modeAbsX: cpu.absX,
		modeAbsY: cpu.absY,
		modeImmd: cpu.immd,
		modeImpl: cpu.impl,
		modeIndi: cpu.ind,
		modeXInd: cpu.xInd,
		modeIndY: cpu.indY,
		modeRela: cpu.rel,
		modeZpag: cpu.zpg,
		modeZpgX: cpu.zpgX,
		modeZpgY: cpu.zpgY,
	}
	cpu.instToInstFunc = map[Instruction]InstructionFunc{
		ADC: cpu.adc,
		AND: cpu.and,
		ASL: cpu.asl,
		BCC: cpu.bcc,
		BCS: cpu.bcs,
		BEQ: cpu.beq,
		BIT: cpu.bit,
		BMI: cpu.bmi,
		BNE: cpu.bne,
		BPL: cpu.bpl,
		BRK: cpu.brk,
		BVC: cpu.bvc,
		BVS: cpu.bvs,
		CLC: cpu.clc,
		CLD: cpu.cld,
		CLI: cpu.cli,
		CLV: cpu.clv,
		CMP: cpu.cmp,
		CPX: cpu.cpx,
		CPY: cpu.cpy,
		DEC: cpu.dec,
		DEX: cpu.dex,
		DEY: cpu.dey,
		EOR: cpu.eor,
		INC: cpu.inc,
		INX: cpu.inx,
		INY: cpu.iny,
		JMP: cpu.jmp,
		JSR: cpu.jsr,
		LDA: cpu.lda,
		LDX: cpu.ldx,
		LDY: cpu.ldy,
		LSR: cpu.lsr,
		NOP: cpu.nop,
		ORA: cpu.ora,
		PHA: cpu.pha,
		PHP: cpu.php,
		PLA: cpu.pla,
		PLP: cpu.plp,
		ROL: cpu.rol,
		ROR: cpu.ror,
		RTI: cpu.rti,
		RTS: cpu.rts,
		SBC: cpu.sbc,
		SEC: cpu.sec,
		SED: cpu.sed,
		SEI: cpu.sei,
		STA: cpu.sta,
		STX: cpu.stx,
		STY: cpu.sty,
		TAX: cpu.tax,
		TAY: cpu.tay,
		TSX: cpu.tsx,
		TXA: cpu.txa,
		TXS: cpu.txs,
		TYA: cpu.tya,
	}
	return cpu
}

func (c *CPU) Reset() {
	// Set PC
	c.pc = c.Read16(0xFFFC)

	// Reset registers
	c.a = 0
	c.x = 0
	c.y = 0
	c.stackPtr = 0xFD

	// Reset status register
	c.c = 0
	c.z = 0
	c.i = 1
	c.d = 0
	c.b = 0
	c.u = 1
	c.v = 0
	c.n = 0

	// Reset cycle
	c.cycle = 7
}

// Read will read 1 byte (8 bits) from the given address.
func (c *CPU) Read(addr uint16) uint8 {
	return c.bus.CpuRead(addr)
}

// Write will write 1 byte of data to the given address.
func (c *CPU) Write(addr uint16, data uint8) {
	c.bus.CpuWrite(addr, data)
}

// Read16 will read 2 bytes (16 bits) from the given address.
// 16-bit address words are little endian, lo(w)-byte first, followed by the hi(gh)-byte.
// (An assembler will use a human-readable, big-endian notation as in $HHLL)
func (c *CPU) Read16(addr uint16) uint16 {
	lo := uint16(c.Read(addr))
	hi := uint16(c.Read(addr + 1))
	return hi<<8 | lo
}

func (c *CPU) Write16(addr, data uint16) {
	hi := uint8(data >> 8)
	lo := uint8(data & 0xFF)
	c.Write(addr, lo)
	c.Write(addr+1, hi)
}

func (c *CPU) Push(data uint8) {
	//fmt.Printf("Pushing to 0x%02X to stack at 0x%04X\n", data, 0x100|uint16(c.stackPtr))
	c.Write(0x100|uint16(c.stackPtr), data)
	c.stackPtr--
}

func (c *CPU) Push16(data uint16) {
	hi := uint8(data >> 8)
	lo := uint8(data & 0xFF)
	c.Push(hi)
	c.Push(lo)
}

func (c *CPU) Pull() uint8 {
	c.stackPtr++
	data := c.Read(0x100 | uint16(c.stackPtr))
	//fmt.Printf("Pulling data 0x%02X from stack at 0x%04X\n", data, 0x100|uint16(c.stackPtr))
	return data
}

func (c *CPU) Pull16() uint16 {
	lo := uint16(c.Pull())
	hi := uint16(c.Pull())
	return hi<<8 | lo
}

func (c *CPU) GetStatus() uint8 {
	N := c.n << 7
	V := c.v << 6
	U := c.u << 5
	B := c.b << 4
	D := c.d << 3
	I := c.i << 2
	Z := c.z << 1
	C := c.c << 0
	return N | V | U | B | D | I | Z | C
}

func (c *CPU) PushStatus() {
	c.Push(c.GetStatus())
}

func (c *CPU) PullStatus() {
	status := c.Pull()
	c.n = status >> 7 & 0x1
	c.v = status >> 6 & 0x1
	c.u = 0
	c.b = 0
	c.d = status >> 3 & 0x1
	c.i = status >> 2 & 0x1
	c.z = status >> 1 & 0x1
	c.c = status >> 0 & 0x1
}

func (c *CPU) Clock() {
	opcode := c.Read(c.pc)
	info := c.getInstructionInfo(opcode)

	fmt.Printf("%04X, ", c.pc)
	for i := uint8(0); i < info.instSize; i++ {
		fmt.Printf("%02X ", c.Read(c.pc+uint16(i)))
	}
	fmt.Printf("\t\tA: %02X X: %02X Y: %02X P: %02X SP: %02X", c.a, c.x, c.y, c.GetStatus(), c.stackPtr)

	fmt.Println("\tCYC: ", c.cycle)
	addrInfo := info.addrModeFunc()

	c.pc += uint16(info.instSize)

	hasAdditionalCycles := info.instFunc(addrInfo.mode, addrInfo.address)
	if hasAdditionalCycles {
		c.cycle += c.GetAdditionalCycles(info, addrInfo)
	}

	c.cycle += int(info.instCycles)
	// todo: return number of cycles??
}

func (c *CPU) GetAdditionalCycles(info InstructionInfo, addrInfo AddressInfo) int {
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
