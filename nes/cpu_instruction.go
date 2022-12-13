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

func (cpu *CPU) getInstructionInfo(opcode uint8) InstructionInfo {
	opInfo := opcodeToInfo[opcode]
	addrModeFunc := cpu.addressModeToAddressModeFunc[opInfo.addrMode]
	instFunc := cpu.instToInstFunc[opInfo.inst]
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
func (cpu *CPU) adc(mode AddressMode, addr uint16) bool {
	A := cpu.a
	M := cpu.Read(addr)
	carry := cpu.GetFlag(C)
	cpu.a = A + M + carry

	cpu.SetFlag(N, IsNegative(cpu.a))
	cpu.SetFlag(Z, IsZero(cpu.a))
	cpu.SetFlag(C, uint16(A)+uint16(M)+uint16(carry) > 0xFF)
	cpu.SetFlag(V, (A^M)&0x80 == 0 && (A^cpu.a)&0x80 != 0)

	return true
}

// AND - AND Memory with Accumulator
//
//	A AND M -> A
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) and(mode AddressMode, addr uint16) bool {
	A := cpu.a
	M := cpu.Read(addr)
	cpu.a = A & M

	cpu.SetFlag(N, IsNegative(cpu.a))
	cpu.SetFlag(Z, IsZero(cpu.a))

	return true
}

// ASL - Shift Left One Bit (Memory or Accumulator)
//
//	C <- [76543210] <- 0
//
//	N Z C I D V
//	+ + + - - -
func (cpu *CPU) asl(mode AddressMode, addr uint16) bool {
	if mode == modeAccu {

		cpu.SetFlag(C, uint16(cpu.a)<<1 > 0xFF)

		cpu.a <<= 1

		cpu.SetFlag(N, IsNegative(cpu.a))
		cpu.SetFlag(Z, IsZero(cpu.a))
	} else {
		M := cpu.Read(addr)

		cpu.SetFlag(C, uint16(M)<<1 > 0xFF)

		M <<= 1

		cpu.SetFlag(N, IsNegative(M))
		cpu.SetFlag(Z, IsZero(M))

		cpu.Write(addr, M)
	}

	return false
}

// BCC - Branch on Carry Clear
//
//	branch on C = 0
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) bcc(mode AddressMode, addr uint16) bool {
	if cpu.GetFlag(C) == 0 {
		cpu.pc = addr
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
func (cpu *CPU) bcs(mode AddressMode, addr uint16) bool {
	if cpu.GetFlag(C) == 1 {
		cpu.pc = addr
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
func (cpu *CPU) beq(mode AddressMode, addr uint16) bool {
	if cpu.GetFlag(Z) == 1 {
		cpu.pc = addr
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
func (cpu *CPU) bit(mode AddressMode, addr uint16) bool {
	A := cpu.a
	M := cpu.Read(addr)

	cpu.SetFlag(N, IsNegative(M))
	cpu.SetFlag(Z, IsZero(A&M))
	cpu.SetFlag(V, M&0x40 == 0x40)

	return false
}

// BMI - Branch on Result Minus
//
//	branch on N = 1
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) bmi(mode AddressMode, addr uint16) bool {
	if cpu.GetFlag(N) == 1 {
		cpu.pc = addr
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
func (cpu *CPU) bne(mode AddressMode, addr uint16) bool {
	if cpu.GetFlag(Z) == 0 {
		cpu.pc = addr
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
func (cpu *CPU) bpl(mode AddressMode, addr uint16) bool {
	if cpu.GetFlag(N) == 0 {
		cpu.pc = addr
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
func (cpu *CPU) brk(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(I, true)
	cpu.Push16(cpu.pc)
	cpu.SetFlag(B, true)
	cpu.PushStatus()
	cpu.SetFlag(B, false)
	cpu.pc = cpu.Read16(0xFFFE)

	return false
}

// BVC - Branch on Overflow Clear
//
//	branch on V = 0
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) bvc(mode AddressMode, addr uint16) bool {
	if cpu.GetFlag(V) == 0 {
		cpu.pc = addr
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
func (cpu *CPU) bvs(mode AddressMode, addr uint16) bool {
	if cpu.GetFlag(V) == 1 {
		cpu.pc = addr
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
func (cpu *CPU) clc(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(C, false)

	return false
}

// CLD - Clear Decimal Mode
//
//	0 -> D
//
//	N Z C I D V
//	- - - - 0 -
func (cpu *CPU) cld(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(D, false)

	return false
}

// CLD - Clear Interrupt Disable Bit
//
//	0 -> I
//
//	N Z C I D V
//	- - - 0 - -
func (cpu *CPU) cli(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(I, false)

	return false
}

// CLD - Clear Overflow Flag
//
//	0 -> V
//
//	N Z C I D V
//	- - - - - 0
func (cpu *CPU) clv(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(V, false)

	return false
}

// CMP - Compare Memory with Accumulator
//
//	A - M
//
//	N Z C I D V
//	+ + + - - -
func (cpu *CPU) cmp(mode AddressMode, addr uint16) bool {
	A := cpu.a
	M := cpu.Read(addr)
	if A < M {
		cpu.SetFlag(Z, false)
		cpu.SetFlag(C, false)
		cpu.SetFlag(N, IsNegative(A-M))
	} else if A > M {
		cpu.SetFlag(Z, false)
		cpu.SetFlag(C, true)
		cpu.SetFlag(N, IsNegative(A-M))
	} else {
		cpu.SetFlag(Z, true)
		cpu.SetFlag(C, true)
		cpu.SetFlag(N, IsNegative(A-M))
	}
	return true
}

// CPX - Compare Memory and Index X
//
//	X - M
//
//	N Z C I D V
//	+ + + - - -
func (cpu *CPU) cpx(mode AddressMode, addr uint16) bool {
	X := cpu.x
	M := cpu.Read(addr)
	if X < M {
		cpu.SetFlag(Z, false)
		cpu.SetFlag(C, false)
		cpu.SetFlag(N, IsNegative(X-M))
	} else if X > M {
		cpu.SetFlag(Z, false)
		cpu.SetFlag(C, true)
		cpu.SetFlag(N, IsNegative(X-M))
	} else {
		cpu.SetFlag(Z, true)
		cpu.SetFlag(C, true)
		cpu.SetFlag(N, IsNegative(X-M))
	}
	return true
}

// CPY - Compare Memory and Index Y
//
//	Y - M
//
//	N Z C I D V
//	+ + + - - -
func (cpu *CPU) cpy(mode AddressMode, addr uint16) bool {
	Y := cpu.y
	M := cpu.Read(addr)
	if Y < M {
		cpu.SetFlag(Z, false)
		cpu.SetFlag(C, false)
		cpu.SetFlag(N, IsNegative(Y-M))
	} else if Y > M {
		cpu.SetFlag(Z, false)
		cpu.SetFlag(C, true)
		cpu.SetFlag(N, IsNegative(Y-M))
	} else {
		cpu.SetFlag(Z, true)
		cpu.SetFlag(C, true)
		cpu.SetFlag(N, IsNegative(Y-M))
	}
	return true
}

// DEC - Decrement Memory by One
//
//	M - 1 -> M
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) dec(mode AddressMode, addr uint16) bool {
	M := cpu.Read(addr)
	M -= 1
	cpu.Write(addr, M)

	cpu.SetFlag(N, IsNegative(M))
	cpu.SetFlag(Z, IsZero(M))

	return false
}

// DEX - Decrement Index X by One
//
//	X - 1 -> X
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) dex(mode AddressMode, addr uint16) bool {
	cpu.x--

	cpu.SetFlag(N, IsNegative(cpu.x))
	cpu.SetFlag(Z, IsZero(cpu.x))

	return false
}

// DEY - Decrement Index Y by One
//
//	Y - 1 -> Y
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) dey(mode AddressMode, addr uint16) bool {
	cpu.y--

	cpu.SetFlag(N, IsNegative(cpu.y))
	cpu.SetFlag(Z, IsZero(cpu.y))

	return false
}

// EOR - Exclusive-OR Memory with Accumulator
//
//	A EOR M -> A
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) eor(mode AddressMode, addr uint16) bool {
	A := cpu.a
	M := cpu.Read(addr)
	A ^= M
	cpu.a = A

	cpu.SetFlag(N, IsNegative(A))
	cpu.SetFlag(Z, IsZero(A))

	return true
}

// INC - Increment Memory by One
//
//	M + 1 -> M
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) inc(mode AddressMode, addr uint16) bool {
	M := cpu.Read(addr)
	M += 1
	cpu.Write(addr, M)

	cpu.SetFlag(N, IsNegative(M))
	cpu.SetFlag(Z, IsZero(M))

	return false
}

// INX - Increment Index X by One
//
//	X + 1 -> X
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) inx(mode AddressMode, addr uint16) bool {
	cpu.x++

	cpu.SetFlag(N, IsNegative(cpu.x))
	cpu.SetFlag(Z, IsZero(cpu.x))

	return false
}

// INY - Increment Index Y by One
//
//	Y + 1 -> Y
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) iny(mode AddressMode, addr uint16) bool {
	cpu.y++

	cpu.SetFlag(N, IsNegative(cpu.y))
	cpu.SetFlag(Z, IsZero(cpu.y))

	return false
}

// JMP - Jump to New Location
//
//	(PC+1) -> PCL
//	(PC+2) -> PCH
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) jmp(mode AddressMode, addr uint16) bool {
	cpu.pc = addr

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
func (cpu *CPU) jsr(mode AddressMode, addr uint16) bool {
	cpu.Push16(cpu.pc - 1) // ?!?!
	cpu.pc = addr

	return false
}

// LDA - Load Accumulator with Memory
//
//	M -> A
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) lda(mode AddressMode, addr uint16) bool {
	cpu.a = cpu.Read(addr)

	cpu.SetFlag(N, IsNegative(cpu.a))
	cpu.SetFlag(Z, IsZero(cpu.a))

	return true
}

// LDX - Load Index X with Memory
//
//	M -> X
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) ldx(mode AddressMode, addr uint16) bool {
	cpu.x = cpu.Read(addr)

	cpu.SetFlag(N, IsNegative(cpu.x))
	cpu.SetFlag(Z, IsZero(cpu.x))

	return true
}

// LDY - Load Index Y with Memory
//
//	M -> Y
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) ldy(mode AddressMode, addr uint16) bool {
	cpu.y = cpu.Read(addr)

	cpu.SetFlag(N, IsNegative(cpu.y))
	cpu.SetFlag(Z, IsZero(cpu.y))

	return true
}

// LSR - Shift One Bit Right (Memory or Accumulator)
//
//	0 -> [76543210] -> C
//
//	N Z C I D V
//	0 + + - - -
func (cpu *CPU) lsr(mode AddressMode, addr uint16) bool {
	if mode == modeAccu {
		cpu.SetFlag(C, cpu.a&1 > 0)
		cpu.a >>= 1
		cpu.SetFlag(N, IsNegative(cpu.a))
		cpu.SetFlag(Z, IsZero(cpu.a))
	} else {
		M := cpu.Read(addr)
		cpu.SetFlag(C, M&1 > 0)
		M >>= 1
		cpu.SetFlag(N, IsNegative(M))
		cpu.SetFlag(Z, IsZero(M))
		cpu.Write(addr, M)
	}

	return false
}

// NOP - No Operation
func (cpu *CPU) nop(mode AddressMode, addr uint16) bool {
	return false
}

// ORA - OR Memory with Accumulator
//
//	A OR M -> A
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) ora(mode AddressMode, addr uint16) bool {
	cpu.a |= cpu.Read(addr)

	cpu.SetFlag(N, IsNegative(cpu.a))
	cpu.SetFlag(Z, IsZero(cpu.a))

	return true
}

// PHA - Push Accumulator on Stack
//
//	push A
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) pha(mode AddressMode, addr uint16) bool {
	cpu.Push(cpu.a)

	return false
}

// PHP - Push Processor Status on Stack
//
//	push SR
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) php(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(U, true)
	cpu.SetFlag(B, true)
	cpu.PushStatus()
	cpu.SetFlag(B, false)

	return false
}

// PLA - Pull Accumulator from Stack
//
//	pull A
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) pla(mode AddressMode, addr uint16) bool {
	cpu.a = cpu.Pull()

	cpu.SetFlag(N, IsNegative(cpu.a))
	cpu.SetFlag(Z, IsZero(cpu.a))

	return false
}

// PLP - Pull Processor Status from Stack
//
//	pull SR
//
//	N Z C I D V
//	from stack
func (cpu *CPU) plp(mode AddressMode, addr uint16) bool {
	cpu.PullStatus()

	return false
}

// ROL - Rotate One Bit Left (Memory or Accumulator)
//
//	C <- [76543210] <- C
//
//	N Z C I D V
//	+ + + - - -
func (cpu *CPU) rol(mode AddressMode, addr uint16) bool {
	if mode == modeAccu {
		carry := cpu.GetFlag(C)
		cpu.SetFlag(C, cpu.a>>7 == 1)
		cpu.a = cpu.a<<1 | carry

		cpu.SetFlag(N, IsNegative(cpu.a))
		cpu.SetFlag(Z, IsZero(cpu.a))
	} else {
		carry := cpu.GetFlag(C)
		M := cpu.Read(addr)
		cpu.SetFlag(C, M>>7 == 1)
		M = M<<1 | carry

		cpu.SetFlag(N, IsNegative(M))
		cpu.SetFlag(Z, IsZero(M))
		cpu.Write(addr, M)
	}

	return false
}

// ROR - Rotate One Bit Right (Memory or Accumulator)
//
//	C -> [76543210] -> C
//
//	N Z C I D V
//	+ + + - - -
func (cpu *CPU) ror(mode AddressMode, addr uint16) bool {
	if mode == modeAccu {
		carry := cpu.GetFlag(C)
		cpu.SetFlag(C, cpu.a&1 == 1)
		cpu.a = cpu.a>>1 | carry<<7

		cpu.SetFlag(N, IsNegative(cpu.a))
		cpu.SetFlag(Z, IsZero(cpu.a))
	} else {
		carry := cpu.GetFlag(C)
		M := cpu.Read(addr)
		cpu.SetFlag(C, M&1 == 1)
		M = M>>1 | carry<<7

		cpu.SetFlag(N, IsNegative(M))
		cpu.SetFlag(Z, IsZero(M))
		cpu.Write(addr, M)
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
func (cpu *CPU) rti(mode AddressMode, addr uint16) bool {
	cpu.PullStatus()
	cpu.pc = cpu.Pull16()

	return false
}

// RTS - Return from Subroutine
//
//	pull PC, PC+1 -> PC
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) rts(mode AddressMode, addr uint16) bool {
	cpu.pc = cpu.Pull16() + 1

	return false
}

// SBC - Subtract Memory from Accumulator with Borrow
//
//	A - M - C -> A
//
//	N Z C I D V
//	+ + + - - +
func (cpu *CPU) sbc(mode AddressMode, addr uint16) bool {
	A := uint16(cpu.a)
	M := uint16(cpu.Read(addr)) ^ 0x00FF
	carry := uint16(cpu.GetFlag(C))
	temp := A + M + carry
	cpu.a = uint8(temp)

	cpu.SetFlag(N, IsNegative(cpu.a))
	cpu.SetFlag(Z, IsZero(cpu.a))
	cpu.SetFlag(C, A+M+carry > 0xFF)
	cpu.SetFlag(V, (A^M)&0x80 == 0 && (A^temp)&0x80 != 0)

	return true
}

// SEC - Set Carry Flag
//
//	1 -> C
//
//	N Z C I D V
//	- - 1 - - -
func (cpu *CPU) sec(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(C, true)

	return false
}

// SED - Set Decimal Flag
//
//	1 -> D
//
//	N Z C I D V
//	- - - - 1 -
func (cpu *CPU) sed(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(D, true)

	return false
}

// SEI - Set Interrupt Disable Status
//
//	1 -> I
//
//	N Z C I D V
//	- - - 1 - -
func (cpu *CPU) sei(mode AddressMode, addr uint16) bool {
	cpu.SetFlag(I, true)

	return false
}

// STA - Store Accumulator in Memory
//
//	A -> M
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) sta(mode AddressMode, addr uint16) bool {
	cpu.Write(addr, cpu.a)

	return false
}

// STX - Store Index X in Memory
//
//	X -> M
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) stx(mode AddressMode, addr uint16) bool {
	cpu.Write(addr, cpu.x)

	return false
}

// STY - Store Index Y in Memory
//
//	Y -> M
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) sty(mode AddressMode, addr uint16) bool {
	cpu.Write(addr, cpu.y)

	return false
}

// TAX - Transfer Accumulator to Index X
//
//	A -> X
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) tax(mode AddressMode, addr uint16) bool {
	cpu.x = cpu.a

	cpu.SetFlag(N, IsNegative(cpu.x))
	cpu.SetFlag(Z, IsZero(cpu.x))

	return false
}

// TAY - Transfer Accumulator to Index Y
//
//	A -> Y
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) tay(mode AddressMode, addr uint16) bool {
	cpu.y = cpu.a

	cpu.SetFlag(N, IsNegative(cpu.y))
	cpu.SetFlag(Z, IsZero(cpu.y))

	return false
}

// TSX - Transfer Stack Pointer to Index X
//
//	SP -> X
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) tsx(mode AddressMode, addr uint16) bool {
	cpu.x = cpu.sp

	cpu.SetFlag(N, IsNegative(cpu.x))
	cpu.SetFlag(Z, IsZero(cpu.x))

	return false
}

// TXA - Transfer Index X to Accumulator
//
//	X -> A
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) txa(mode AddressMode, addr uint16) bool {
	cpu.a = cpu.x

	cpu.SetFlag(N, IsNegative(cpu.a))
	cpu.SetFlag(Z, IsZero(cpu.a))

	return false
}

// TXS - Transfer Index X to Stack Register
//
//	X -> SP
//
//	N Z C I D V
//	- - - - - -
func (cpu *CPU) txs(mode AddressMode, addr uint16) bool {
	cpu.sp = cpu.x

	return false
}

// TYA - Transfer Index Y to Accumulator
//
//	Y -> A
//
//	N Z C I D V
//	+ + - - - -
func (cpu *CPU) tya(mode AddressMode, addr uint16) bool {
	cpu.a = cpu.y

	cpu.SetFlag(N, IsNegative(cpu.a))
	cpu.SetFlag(Z, IsZero(cpu.a))

	return false
}
