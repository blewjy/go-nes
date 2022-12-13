// Reference document: https://www.masswerk.at/6502/6502_instruction_set.html

package nes

type OpInfo struct {
	addrMode   AddressMode
	inst       Instruction
	instSize   uint8
	instCycles uint8
}

var (
	opcodeToInfo = [256]OpInfo{
		{modeImpl, BRK, 1, 7}, {modeXInd, ORA, 2, 6}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpag, ORA, 2, 3}, {modeZpag, ASL, 2, 5}, {modeNone, NOP, 0, 0}, {modeImpl, PHP, 1, 3}, {modeImmd, ORA, 2, 2}, {modeAccu, ASL, 1, 2}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeAbso, ORA, 3, 4}, {modeAbso, ASL, 3, 6}, {modeNone, NOP, 0, 0},
		{modeRela, BPL, 2, 2}, {modeIndY, ORA, 2, 5}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpgX, ORA, 2, 4}, {modeZpgX, ASL, 2, 6}, {modeNone, NOP, 0, 0}, {modeImpl, CLC, 1, 2}, {modeAbsY, ORA, 3, 4}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeAbsX, ORA, 3, 4}, {modeAbsX, ASL, 3, 7}, {modeNone, NOP, 0, 0},
		{modeAbso, JSR, 3, 6}, {modeXInd, AND, 2, 6}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpag, BIT, 2, 3}, {modeZpag, AND, 2, 3}, {modeZpag, ROL, 2, 5}, {modeNone, NOP, 0, 0}, {modeImpl, PLP, 1, 4}, {modeImmd, AND, 2, 2}, {modeAccu, ROL, 1, 2}, {modeNone, NOP, 0, 0}, {modeAbso, BIT, 3, 4}, {modeAbso, AND, 3, 4}, {modeAbso, ROL, 3, 6}, {modeNone, NOP, 0, 0},
		{modeRela, BMI, 2, 2}, {modeIndY, AND, 2, 5}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpgX, AND, 2, 4}, {modeZpgX, ROL, 2, 6}, {modeNone, NOP, 0, 0}, {modeImpl, SEC, 1, 2}, {modeAbsY, AND, 3, 4}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeAbsX, AND, 3, 4}, {modeAbsX, ROL, 3, 7}, {modeNone, NOP, 0, 0},
		{modeImpl, RTI, 1, 6}, {modeXInd, EOR, 2, 6}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpag, EOR, 2, 3}, {modeZpag, LSR, 2, 5}, {modeNone, NOP, 0, 0}, {modeImpl, PHA, 1, 3}, {modeImmd, EOR, 2, 2}, {modeAccu, LSR, 1, 2}, {modeNone, NOP, 0, 0}, {modeAbso, JMP, 3, 3}, {modeAbso, EOR, 3, 4}, {modeAbso, LSR, 3, 6}, {modeNone, NOP, 0, 0},
		{modeRela, BVC, 2, 2}, {modeIndY, EOR, 2, 5}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpgX, EOR, 2, 4}, {modeZpgX, LSR, 2, 6}, {modeNone, NOP, 0, 0}, {modeImpl, CLI, 1, 2}, {modeAbsY, EOR, 3, 4}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeAbsX, EOR, 3, 4}, {modeAbsX, LSR, 3, 7}, {modeNone, NOP, 0, 0},
		{modeImpl, RTS, 1, 6}, {modeXInd, ADC, 2, 6}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpag, ADC, 2, 3}, {modeZpag, ROR, 2, 5}, {modeNone, NOP, 0, 0}, {modeImpl, PLA, 1, 4}, {modeImmd, ADC, 2, 2}, {modeAccu, ROR, 1, 2}, {modeNone, NOP, 0, 0}, {modeIndi, JMP, 3, 5}, {modeAbso, ADC, 3, 4}, {modeAbso, ROR, 3, 6}, {modeNone, NOP, 0, 0},
		{modeRela, BVS, 2, 2}, {modeIndY, ADC, 2, 5}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpgX, ADC, 2, 4}, {modeZpgX, ROR, 2, 6}, {modeNone, NOP, 0, 0}, {modeImpl, SEI, 1, 2}, {modeAbsY, ADC, 3, 4}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeAbsX, ADC, 3, 4}, {modeAbsX, ROR, 3, 7}, {modeNone, NOP, 0, 0},
		{modeNone, NOP, 0, 0}, {modeXInd, STA, 2, 6}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpag, STY, 2, 3}, {modeZpag, STA, 2, 3}, {modeZpag, STX, 2, 3}, {modeNone, NOP, 0, 0}, {modeImpl, DEY, 1, 2}, {modeNone, NOP, 0, 0}, {modeImpl, TXA, 1, 2}, {modeNone, NOP, 0, 0}, {modeAbso, STY, 3, 4}, {modeAbso, STA, 3, 4}, {modeAbso, STX, 3, 4}, {modeNone, NOP, 0, 0},
		{modeRela, BCC, 2, 2}, {modeIndY, STA, 2, 6}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpgX, STY, 2, 4}, {modeZpgX, STA, 2, 4}, {modeZpgY, STX, 2, 4}, {modeNone, NOP, 0, 0}, {modeImpl, TYA, 1, 2}, {modeAbsY, STA, 3, 5}, {modeImpl, TXS, 1, 2}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeAbsX, STA, 3, 5}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0},
		{modeImmd, LDY, 2, 2}, {modeXInd, LDA, 2, 6}, {modeImmd, LDX, 2, 2}, {modeNone, NOP, 0, 0}, {modeZpag, LDY, 2, 3}, {modeZpag, LDA, 2, 3}, {modeZpag, LDX, 2, 3}, {modeNone, NOP, 0, 0}, {modeImpl, TAY, 1, 2}, {modeImmd, LDA, 2, 2}, {modeImpl, TAX, 1, 2}, {modeNone, NOP, 0, 0}, {modeAbso, LDY, 3, 4}, {modeAbso, LDA, 3, 4}, {modeAbso, LDX, 3, 4}, {modeNone, NOP, 0, 0},
		{modeRela, BCS, 2, 2}, {modeIndY, LDA, 2, 5}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpgX, LDY, 2, 4}, {modeZpgX, LDA, 2, 4}, {modeZpgY, LDX, 2, 4}, {modeNone, NOP, 0, 0}, {modeImpl, CLV, 1, 2}, {modeAbsY, LDA, 3, 4}, {modeImpl, TSX, 1, 2}, {modeNone, NOP, 0, 0}, {modeAbsX, LDY, 3, 4}, {modeAbsX, LDA, 3, 4}, {modeAbsY, LDX, 3, 4}, {modeNone, NOP, 0, 0},
		{modeImmd, CPY, 2, 2}, {modeXInd, CMP, 2, 6}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpag, CPY, 2, 3}, {modeZpag, CMP, 2, 3}, {modeZpag, DEC, 2, 5}, {modeNone, NOP, 0, 0}, {modeImpl, INY, 1, 2}, {modeImmd, CMP, 2, 2}, {modeImpl, DEX, 1, 2}, {modeNone, NOP, 0, 0}, {modeAbso, CPY, 3, 4}, {modeAbso, CMP, 3, 4}, {modeAbso, DEC, 3, 6}, {modeNone, NOP, 0, 0},
		{modeRela, BNE, 2, 2}, {modeIndY, CMP, 2, 5}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpgX, CMP, 2, 4}, {modeZpgX, DEC, 2, 6}, {modeNone, NOP, 0, 0}, {modeImpl, CLD, 1, 2}, {modeAbsY, CMP, 3, 4}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeAbsX, CMP, 3, 4}, {modeAbsX, DEC, 3, 7}, {modeNone, NOP, 0, 0},
		{modeImmd, CPX, 2, 2}, {modeXInd, SBC, 2, 6}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpag, CPX, 2, 3}, {modeZpag, SBC, 2, 3}, {modeZpag, INC, 2, 5}, {modeNone, NOP, 0, 0}, {modeImpl, INX, 1, 2}, {modeImmd, SBC, 2, 2}, {modeImpl, NOP, 1, 2}, {modeNone, NOP, 0, 0}, {modeAbso, CPX, 3, 4}, {modeAbso, SBC, 3, 4}, {modeAbso, INC, 3, 6}, {modeNone, NOP, 0, 0},
		{modeRela, BEQ, 2, 2}, {modeIndY, SBC, 2, 5}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeZpgX, SBC, 2, 4}, {modeZpgX, INC, 2, 6}, {modeNone, NOP, 0, 0}, {modeImpl, SED, 1, 2}, {modeAbsY, SBC, 3, 4}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeNone, NOP, 0, 0}, {modeAbsX, SBC, 3, 4}, {modeAbsX, INC, 3, 7}, {modeNone, NOP, 0, 0},
	}
)
