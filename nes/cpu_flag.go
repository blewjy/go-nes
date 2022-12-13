package nes

// 7  bit  0
// ---- ----
// NVUB DIZC
// |||| ||||
// |||| |||+- Carry
// |||| ||+-- Zero
// |||| |+--- Interrupt Disable
// |||| +---- Decimal
// ||++------ No CPU effect, see: the B flag
// |+-------- Overflow
// +--------- Negative

type Flag uint8

const (
	C Flag = 1 << 0
	Z      = 1 << 1
	I      = 1 << 2
	D      = 1 << 3
	B      = 1 << 4
	U      = 1 << 5
	V      = 1 << 6
	N      = 1 << 7
)

func (cpu *CPU) GetFlag(flag Flag) uint8 {
	if cpu.p&uint8(flag) > 0 {
		return 1
	} else {
		return 0
	}
}

func (cpu *CPU) SetFlag(flag Flag, set bool) {
	if set {
		cpu.p |= uint8(flag)
	} else {
		cpu.p &= ^uint8(flag)
	}
}

func IsNegative(b uint8) bool {
	return b&0x80 == 0x80
}

func IsZero(b uint8) bool {
	return b == 0
}
