package nes

import (
	"fmt"
	"math"
)

type CPU struct {
	bus *Bus

	a  uint8  // Accumulator register
	x  uint8  // X register
	y  uint8  // Y register
	p  uint8  // Status register
	sp uint8  // Stack pointer
	pc uint16 // Program counter

	// cycles???
	cycle int

	// Opcode table
	table [256]OpcodeInfo
}

func NewCPU(bus *Bus) *CPU {
	cpu := &CPU{
		bus: bus,
	}
	cpu.InitOpcodeTable()
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
	return cpu.Read(0x100 | uint16(cpu.sp))
}

func (cpu *CPU) Pull16() uint16 {
	lo := uint16(cpu.Pull())
	hi := uint16(cpu.Pull())
	return hi<<8 | lo
}

func (cpu *CPU) GetStatus() uint8 {
	return cpu.p
}

func (cpu *CPU) PushStatus() {
	cpu.Push(cpu.p)
}

func (cpu *CPU) PullStatus() {
	cpu.p = cpu.Pull()
	cpu.SetFlag(U, true)
	cpu.SetFlag(B, false)
}

func (cpu *CPU) Clock() {
	opcode := cpu.Read(cpu.pc)
	info := cpu.table[opcode]

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
	info := cpu.table[opcode]

	result += fmt.Sprintf("%04X, ", cpu.pc)
	for i := uint8(0); i < info.instSize; i++ {
		result += fmt.Sprintf("%02X ", cpu.Read(cpu.pc+uint16(i)))
	}
	result += fmt.Sprintf("\t\tA: %02X X: %02X Y: %02X P: %02X SP: %02X", cpu.a, cpu.x, cpu.y, cpu.GetStatus(), cpu.sp)

	result += fmt.Sprint("\tCYC: ", cpu.cycle)

	return result
}

func (cpu *CPU) PeekDisassembly() map[uint16]string {
	disassembly := map[uint16]string{}

	currentAddr := uint16(0)
	prevAddr := uint16(0)

	for currentAddr >= prevAddr {
		prevAddr = currentAddr

		opcode := cpu.Read(currentAddr)
		info := cpu.table[opcode]

		instruction := info.inst.ToString()

		switch info.instSize {
		case 0:
			// do nothing
		case 1:
			// do nothing, inst already added to string
		case 2:
			// 1 operand
			instruction += fmt.Sprintf(" %02X", cpu.Read(currentAddr+1))
		case 3:
			// 2 operand
			instruction += fmt.Sprintf(" %02X", cpu.Read(currentAddr+1)) + fmt.Sprintf(" %02X", cpu.Read(currentAddr+2))
		default:
			panic("unexpected instruction size")
		}

		instruction += " (" + info.addrMode.ToString() + ")"

		disassembly[currentAddr] = instruction

		currentAddr += uint16(math.Max(float64(info.instSize), 1))

	}

	return disassembly
}

func (cpu *CPU) GetAdditionalCycles(info OpcodeInfo, addrInfo AddressInfo) int {
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
