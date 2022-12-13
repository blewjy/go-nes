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

// SetC will set the C flag if the given condition is true; unset otherwise.
func (cpu *CPU) SetC(condition bool) {
	if condition {
		cpu.c = 1
	} else {
		cpu.c = 0
	}
}

// SetZ will set the Z flag if the given condition is true; unset otherwise.
func (cpu *CPU) SetZ(condition bool) {
	if condition {
		cpu.z = 1
	} else {
		cpu.z = 0
	}
}

func (cpu *CPU) SetI(condition bool) {
	if condition {
		cpu.i = 1
	} else {
		cpu.i = 0
	}
}

func (cpu *CPU) SetD() {

}

func (cpu *CPU) SetB(condition bool) {
	if condition {
		cpu.b = 1
	} else {
		cpu.b = 0
	}
}

func (cpu *CPU) SetU(condition bool) {
	if condition {
		cpu.u = 1
	} else {
		cpu.u = 0
	}
}

// SetV will set the V flag if the given condition is true; unset otherwise.
func (cpu *CPU) SetV(condition bool) {
	if condition {
		cpu.v = 1
	} else {
		cpu.v = 0
	}
}

// SetN will set the N flag if the given value is negative; unset otherwise
func (cpu *CPU) SetN(value uint8) {
	if value&0x80 != 0 {
		cpu.n = 1
	} else {
		cpu.n = 0
	}
}
