package nes

type Instruction uint8

const (
	ADC Instruction = iota // add with carry
	AND                    // and (with accumulator)
	ASL                    // arithmetic shift left
	BCC                    // branch on carry clear
	BCS                    // branch on carry set
	BEQ                    // branch on equal (zero set)
	BIT                    // bit test
	BMI                    // branch on minus (negative set)
	BNE                    // branch on not equal (zero clear)
	BPL                    // branch on plus (negative clear)
	BRK                    // break / interrupt
	BVC                    // branch on overflow clear
	BVS                    // branch on overflow set
	CLC                    // clear carry
	CLD                    // clear decimal
	CLI                    // clear interrupt disable
	CLV                    // clear overflow
	CMP                    // compare (with accumulator)
	CPX                    // compare with X
	CPY                    // compare with Y
	DEC                    // decrement
	DEX                    // decrement X
	DEY                    // decrement Y
	EOR                    // exclusive or (with accumulator)
	INC                    // increment
	INX                    // increment X
	INY                    // increment Y
	JMP                    // jump
	JSR                    // jump subroutine
	LDA                    // load accumulator
	LDX                    // load X
	LDY                    // load Y
	LSR                    // logical shift right
	NOP                    // no operation
	ORA                    // or with accumulator
	PHA                    // push accumulator
	PHP                    // push processor status (SR)
	PLA                    // pull accumulator
	PLP                    // pull processor status (SR)
	ROL                    // rotate left
	ROR                    // rotate right
	RTI                    // return from interrupt
	RTS                    // return from subroutine
	SBC                    // subtract with carry
	SEC                    // set carry
	SED                    // set decimal
	SEI                    // set interrupt disable
	STA                    // store accumulator
	STX                    // store X
	STY                    // store Y
	TAX                    // transfer accumulator to X
	TAY                    // transfer accumulator to Y
	TSX                    // transfer stack pointer to X
	TXA                    // transfer X to accumulator
	TXS                    // transfer X to stack pointer
	TYA                    // transfer Y to accumulator
)

func (i Instruction) IsBranch() bool {
	branchInstructions := map[Instruction]bool{BCC: true, BCS: true, BEQ: true, BMI: true, BNE: true, BPL: true, BVC: true, BVS: true}
	if _, ok := branchInstructions[i]; ok {
		return true
	}
	return false
}

type InstructionFunc func(mode AddressMode, addr uint16) bool

type InstructionInfo struct {
	addrMode     AddressMode
	addrModeFunc AddressModeFunc
	inst         Instruction
	instFunc     InstructionFunc
	instSize     uint8
	instCycles   uint8
}

func (c *CPU) getInstructionInfo(opcode uint8) InstructionInfo {
	opInfo := opcodeToInfo[opcode]
	addrModeFunc := c.addressModeToAddressModeFunc[opInfo.addrMode]
	instFunc := c.instToInstFunc[opInfo.inst]
	return InstructionInfo{
		addrMode:     opInfo.addrMode,
		addrModeFunc: addrModeFunc,
		inst:         opInfo.inst,
		instFunc:     instFunc,
		instSize:     opInfo.instSize,
		instCycles:   opInfo.instCycles,
	}
}

// ADC - Add Memory to Accumulator with Carry
//
//	A + M + C -> A, C
//
//	N Z C I D V
//	+ + + - - +
func (c *CPU) adc(mode AddressMode, addr uint16) bool {
	A := c.a
	M := c.Read(addr)
	C := c.c
	c.a = A + M + C

	c.SetN(c.a)
	c.SetZ(c.a == 0)
	c.SetC(uint16(A)+uint16(M)+uint16(C) > 0xFF)
	c.SetV((A^M)&0x80 == 0 && (A^c.a)&0x80 != 0)

	return true
}

// AND - AND Memory with Accumulator
//
//	A AND M -> A
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) and(mode AddressMode, addr uint16) bool {
	A := c.a
	M := c.Read(addr)
	c.a = A & M

	c.SetN(c.a)
	c.SetZ(c.a == 0)

	return true
}

// ASL - Shift Left One Bit (Memory or Accumulator)
//
//	C <- [76543210] <- 0
//
//	N Z C I D V
//	+ + + - - -
func (c *CPU) asl(mode AddressMode, addr uint16) bool {
	if mode == modeAccu {
		c.SetC(uint16(c.a)<<1 > 0xFF)
		c.a <<= 1
		c.SetN(c.a)
		c.SetZ(c.a == 0)
	} else {
		M := c.Read(addr)
		c.SetC(uint16(M)<<1 > 0xFF)
		M <<= 1
		c.SetN(M)
		c.SetZ(M == 0)
		c.Write(addr, M)
	}

	return false
}

// BCC - Branch on Carry Clear
//
//	branch on C = 0
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) bcc(mode AddressMode, addr uint16) bool {
	if c.c == 0 {
		c.pc = addr
		return true
	}

	return false
}

// BCS - Branch on Carry Set
//
//	branch on C = 1
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) bcs(mode AddressMode, addr uint16) bool {
	if c.c == 1 {
		c.pc = addr
		return true
	}

	return false
}

// BEQ - Branch on Result Zero
//
//	branch on Z = 1
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) beq(mode AddressMode, addr uint16) bool {
	if c.z == 1 {
		c.pc = addr
		return true
	}

	return false
}

// BIT - Test Bits in Memory with Accumulator
//
// Bits 7 and 6 of operand are transferred to bit 7 and 6 of SR (N,V);
// the zero-flag is set to the result of operand AND accumulator.
//
//	A AND M, M7 -> N, M6 -> V
//
//	N  Z C I D V
//	M7 + - - - M6
func (c *CPU) bit(mode AddressMode, addr uint16) bool {
	A := c.a
	M := c.Read(addr)
	c.SetN(M)
	c.SetZ(A&M == 0)
	c.SetV(M&0x40 == 0x40)

	return false
}

// BMI - Branch on Result Minus
//
//	branch on N = 1
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) bmi(mode AddressMode, addr uint16) bool {
	if c.n == 1 {
		c.pc = addr
		return true
	}

	return false
}

// BNE - Branch on Result not Zero
//
//	branch on Z = 0
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) bne(mode AddressMode, addr uint16) bool {
	if c.z == 0 {
		c.pc = addr
		return true
	}

	return false
}

// BPL - Branch on Result Plus
//
//	branch on N = 0
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) bpl(mode AddressMode, addr uint16) bool {
	if c.n == 0 {
		c.pc = addr
		return true
	}

	return false
}

// BRK - Force Break
//
// BRK initiates a software interrupt similar to a hardware
// interrupt (IRQ). The return address pushed to the stack is
// PC+2, providing an extra byte of spacing for a break mark
// (identifying a reason for the break.)
// The status register will be pushed to the stack with the break
// flag set to 1. However, when retrieved during RTI or by a PLP
// instruction, the break flag will be ignored.
// The interrupt disable flag is not set automatically.
//
//		interrupt,
//	 push PC+2, push SR
//
//		N Z C I D V
//		- - - 1 - -
func (c *CPU) brk(mode AddressMode, addr uint16) bool {
	c.SetI(true)
	c.Push16(c.pc)
	c.SetB(true)
	c.PushStatus()
	c.SetB(false)
	c.pc = c.Read16(0xFFFE)

	return false
}

// BVC - Branch on Overflow Clear
//
//	branch on V = 0
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) bvc(mode AddressMode, addr uint16) bool {
	if c.v == 0 {
		c.pc = addr
		return true
	}

	return false
}

// BVS - Branch on Overflow Set
//
//	branch on V = 1
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) bvs(mode AddressMode, addr uint16) bool {
	if c.v == 1 {
		c.pc = addr
		return true
	}

	return false
}

// CLC - Clear Carry Flag
//
//	0 -> C
//
//	N Z C I D V
//	- - 0 - - -
func (c *CPU) clc(mode AddressMode, addr uint16) bool {
	c.c = 0

	return false
}

// CLD - Clear Decimal Mode
//
//	0 -> D
//
//	N Z C I D V
//	- - - - 0 -
func (c *CPU) cld(mode AddressMode, addr uint16) bool {
	c.d = 0

	return false
}

// CLD - Clear Interrupt Disable Bit
//
//	0 -> I
//
//	N Z C I D V
//	- - - 0 - -
func (c *CPU) cli(mode AddressMode, addr uint16) bool {
	c.i = 0

	return false
}

// CLD - Clear Overflow Flag
//
//	0 -> V
//
//	N Z C I D V
//	- - - - - 0
func (c *CPU) clv(mode AddressMode, addr uint16) bool {
	c.v = 0

	return false
}

// CMP - Compare Memory with Accumulator
//
//	A - M
//
//	N Z C I D V
//	+ + + - - -
func (c *CPU) cmp(mode AddressMode, addr uint16) bool {
	A := c.a
	M := c.Read(addr)
	if A < M {
		c.SetZ(false)
		c.SetC(false)
		c.SetN(A - M)
	} else if A > M {
		c.SetZ(false)
		c.SetC(true)
		c.SetN(0)
	} else {
		c.SetZ(true)
		c.SetC(true)
		c.SetN(A - M)
	}
	return true
}

// CPX - Compare Memory and Index X
//
//	X - M
//
//	N Z C I D V
//	+ + + - - -
func (c *CPU) cpx(mode AddressMode, addr uint16) bool {
	X := c.x
	M := c.Read(addr)
	if X < M {
		c.SetZ(false)
		c.SetC(false)
		c.SetN(X - M)
	} else if X > M {
		c.SetZ(false)
		c.SetC(true)
		c.SetN(0)
	} else {
		c.SetZ(true)
		c.SetC(true)
		c.SetN(X - M)
	}
	return true
}

// CPY - Compare Memory and Index Y
//
//	Y - M
//
//	N Z C I D V
//	+ + + - - -
func (c *CPU) cpy(mode AddressMode, addr uint16) bool {
	Y := c.y
	M := c.Read(addr)
	if Y < M {
		c.SetZ(false)
		c.SetC(false)
		c.SetN(Y - M)
	} else if Y > M {
		c.SetZ(false)
		c.SetC(true)
		c.SetN(0)
	} else {
		c.SetZ(true)
		c.SetC(true)
		c.SetN(Y - M)
	}
	return true
}

// DEC - Decrement Memory by One
//
//	M - 1 -> M
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) dec(mode AddressMode, addr uint16) bool {
	M := c.Read(addr)
	M -= 1
	c.Write(addr, M)

	c.SetN(M)
	c.SetZ(M == 0)

	return false
}

// DEX - Decrement Index X by One
//
//	X - 1 -> X
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) dex(mode AddressMode, addr uint16) bool {
	c.x--

	c.SetN(c.x)
	c.SetZ(c.x == 0)

	return false
}

// DEY - Decrement Index Y by One
//
//	Y - 1 -> Y
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) dey(mode AddressMode, addr uint16) bool {
	c.y--

	c.SetN(c.y)
	c.SetZ(c.y == 0)

	return false
}

// EOR - Exclusive-OR Memory with Accumulator
//
//	A EOR M -> A
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) eor(mode AddressMode, addr uint16) bool {
	A := c.a
	M := c.Read(addr)
	A ^= M
	c.a = A

	c.SetN(A)
	c.SetZ(A == 0)

	return true
}

// INC - Increment Memory by One
//
//	M + 1 -> M
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) inc(mode AddressMode, addr uint16) bool {
	M := c.Read(addr)
	M += 1
	c.Write(addr, M)

	c.SetN(M)
	c.SetZ(M == 0)

	return false
}

// INX - Increment Index X by One
//
//	X + 1 -> X
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) inx(mode AddressMode, addr uint16) bool {
	c.x++

	c.SetN(c.x)
	c.SetZ(c.x == 0)

	return false
}

// INY - Increment Index Y by One
//
//	Y + 1 -> Y
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) iny(mode AddressMode, addr uint16) bool {
	c.y++

	c.SetN(c.y)
	c.SetZ(c.y == 0)

	return false
}

// JMP - Jump to New Location
//
//	(PC+1) -> PCL
//	(PC+2) -> PCH
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) jmp(mode AddressMode, addr uint16) bool {
	c.pc = addr

	return false
}

// JMP - Jump to New Location Saving Return Address
//
//	push (PC+2)
//	(PC+1) -> PCL
//	(PC+2) -> PCH
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) jsr(mode AddressMode, addr uint16) bool {
	c.Push16(c.pc - 1) // ?!?!
	c.pc = addr

	return false
}

// LDA - Load Accumulator with Memory
//
//	M -> A
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) lda(mode AddressMode, addr uint16) bool {
	c.a = c.Read(addr)

	c.SetN(c.a)
	c.SetZ(c.a == 0)

	return true
}

// LDX - Load Index X with Memory
//
//	M -> X
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) ldx(mode AddressMode, addr uint16) bool {
	c.x = c.Read(addr)

	c.SetN(c.x)
	c.SetZ(c.x == 0)

	return true
}

// LDY - Load Index Y with Memory
//
//	M -> Y
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) ldy(mode AddressMode, addr uint16) bool {
	c.y = c.Read(addr)

	c.SetN(c.y)
	c.SetZ(c.y == 0)

	return true
}

// LSR - Shift One Bit Right (Memory or Accumulator)
//
//	0 -> [76543210] -> C
//
//	N Z C I D V
//	0 + + - - -
func (c *CPU) lsr(mode AddressMode, addr uint16) bool {
	if mode == modeAccu {
		c.SetC(c.a&1 > 0)
		c.a >>= 1
		c.SetN(c.a)
		c.SetZ(c.a == 0)
	} else {
		M := c.Read(addr)
		c.SetC(M&1 > 0)
		M >>= 1
		c.SetN(M)
		c.SetZ(M == 0)
		c.Write(addr, M)
	}

	return false
}

// NOP - No Operation
func (c *CPU) nop(mode AddressMode, addr uint16) bool {
	return false
}

// ORA - OR Memory with Accumulator
//
//	A OR M -> A
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) ora(mode AddressMode, addr uint16) bool {
	c.a |= c.Read(addr)

	c.SetN(c.a)
	c.SetZ(c.a == 0)

	return true
}

// PHA - Push Accumulator on Stack
//
//	push A
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) pha(mode AddressMode, addr uint16) bool {
	c.Push(c.a)

	return false
}

// PHP - Push Processor Status on Stack
//
//	push SR
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) php(mode AddressMode, addr uint16) bool {
	c.SetU(true)
	c.SetB(true)
	c.PushStatus()
	c.SetB(false)

	return false
}

// PLA - Pull Accumulator from Stack
//
//	pull A
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) pla(mode AddressMode, addr uint16) bool {
	c.a = c.Pull()

	c.SetN(c.a)
	c.SetZ(c.a == 0)

	return false
}

// PLP - Pull Processor Status from Stack
//
//	pull SR
//
//	N Z C I D V
//	from stack
func (c *CPU) plp(mode AddressMode, addr uint16) bool {
	c.PullStatus()

	return false
}

// ROL - Rotate One Bit Left (Memory or Accumulator)
//
//	C <- [76543210] <- C
//
//	N Z C I D V
//	+ + + - - -
func (c *CPU) rol(mode AddressMode, addr uint16) bool {
	if mode == modeAccu {
		C := c.c
		c.c = c.a >> 7
		c.a = c.a<<1 | C
		c.SetN(c.a)
		c.SetZ(c.a == 0)
	} else {
		C := c.c
		M := c.Read(addr)
		c.c = M >> 7
		M = M<<1 | C
		c.SetN(M)
		c.SetZ(M == 0)
		c.Write(addr, M)
	}

	return false
}

// ROR - Rotate One Bit Right (Memory or Accumulator)
//
//	C -> [76543210] -> C
//
//	N Z C I D V
//	+ + + - - -
func (c *CPU) ror(mode AddressMode, addr uint16) bool {
	if mode == modeAccu {
		C := c.c
		c.c = c.a & 1
		c.a = c.a>>1 | C<<7
		c.SetN(c.a)
		c.SetZ(c.a == 0)
	} else {
		C := c.c
		M := c.Read(addr)
		c.c = M & 1
		M = M>>1 | C<<7
		c.SetN(M)
		c.SetZ(M == 0)
		c.Write(addr, M)
	}

	return false
}

// RTI - Return from Interrupt
//
// The status register is pulled with the break flag
// and bit 5 ignored. Then PC is pulled from the stack.
//
//	pull SR, pull PC
//
//	N Z C I D V
//	from stack
func (c *CPU) rti(mode AddressMode, addr uint16) bool {
	c.PullStatus()
	c.pc = c.Pull16()

	return false
}

// RTS - Return from Subroutine
//
//	pull PC, PC+1 -> PC
//
//	N Z C I D V
//	from stack
func (c *CPU) rts(mode AddressMode, addr uint16) bool {
	c.pc = c.Pull16() + 1

	return false
}

// SBC - Subtract Memory from Accumulator with Borrow
//
//	A - M - C -> A
//
//	N Z C I D V
//	+ + + - - +
func (c *CPU) sbc(mode AddressMode, addr uint16) bool {
	A := c.a
	M := c.Read(addr)
	C := c.c
	c.a = A - M - (1 - C)

	c.SetN(c.a)
	c.SetZ(c.a == 0)
	c.SetC(uint16(A)-uint16(M)-uint16(1-C) >= 0)
	c.SetV((A^M)&0x80 != 0 && (A^c.a)&0x80 != 0)

	return true
}

// SEC - Set Carry Flag
//
//	1 -> C
//
//	N Z C I D V
//	- - 1 - - -
func (c *CPU) sec(mode AddressMode, addr uint16) bool {
	c.c = 1

	return false
}

// SED - Set Decimal Flag
//
//	1 -> D
//
//	N Z C I D V
//	- - - - 1 -
func (c *CPU) sed(mode AddressMode, addr uint16) bool {
	c.d = 1

	return false
}

// SEI - Set Interrupt Disable Status
//
//	1 -> I
//
//	N Z C I D V
//	- - - 1 - -
func (c *CPU) sei(mode AddressMode, addr uint16) bool {
	c.i = 1

	return false
}

// STA - Store Accumulator in Memory
//
//	A -> M
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) sta(mode AddressMode, addr uint16) bool {
	c.Write(addr, c.a)

	return false
}

// STX - Store Index X in Memory
//
//	X -> M
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) stx(mode AddressMode, addr uint16) bool {
	c.Write(addr, c.x)

	return false
}

// STY - Store Index Y in Memory
//
//	Y -> M
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) sty(mode AddressMode, addr uint16) bool {
	c.Write(addr, c.y)

	return false
}

// TAX - Transfer Accumulator to Index X
//
//	A -> X
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) tax(mode AddressMode, addr uint16) bool {
	c.x = c.a

	c.SetN(c.x)
	c.SetZ(c.x == 0)

	return false
}

// TAY - Transfer Accumulator to Index Y
//
//	A -> Y
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) tay(mode AddressMode, addr uint16) bool {
	c.y = c.a

	c.SetN(c.y)
	c.SetZ(c.y == 0)

	return false
}

// TSX - Transfer Stack Pointer to Index X
//
//	SP -> X
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) tsx(mode AddressMode, addr uint16) bool {
	c.x = c.stackPtr

	c.SetN(c.x)
	c.SetZ(c.x == 0)

	return false
}

// TXA - Transfer Index X to Accumulator
//
//	X -> A
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) txa(mode AddressMode, addr uint16) bool {
	c.a = c.x

	c.SetN(c.a)
	c.SetZ(c.a == 0)

	return false
}

// TXS - Transfer Index X to Stack Register
//
//	X -> SP
//
//	N Z C I D V
//	- - - - - -
func (c *CPU) txs(mode AddressMode, addr uint16) bool {
	c.stackPtr = c.x

	return false
}

// TYA - Transfer Index Y to Accumulator
//
//	Y -> A
//
//	N Z C I D V
//	+ + - - - -
func (c *CPU) tya(mode AddressMode, addr uint16) bool {
	c.a = c.y

	c.SetN(c.a)
	c.SetZ(c.a == 0)

	return false
}
