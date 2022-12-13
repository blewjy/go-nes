package main

import (
	"fmt"
	"go-nes/emulator"
	"testing"
)

func assert(a interface{}, b interface{}) {
	if a != b {
		panic(fmt.Sprintf("failed assertion: %v (%T) and %v (%T) are not equal", a, a, b, b))
	}
}

func TestRegression1(t *testing.T) {
	fmt.Println("Running TestRegression 1...")

	program := "A20A8E0000A2038E0100AC0000A900186D010088D0FA8D0200EAEAEA"
	startAddr := uint16(0x0000)
	nes := emulator.NewEmulatorWithMode(emulator.Test)

	cpuSnapshot, ramSnapshot := nes.StartWithProgramAsTest(program, startAddr, 38)

	assert(cpuSnapshot.PC, uint16(0x0019))
	assert(cpuSnapshot.A, uint8(0x1E))
	assert(cpuSnapshot.X, uint8(0x03))
	assert(cpuSnapshot.Y, uint8(0x00))
	assert(cpuSnapshot.StackPtr, uint8(0xFD))
	assert(cpuSnapshot.C, uint8(0))
	assert(cpuSnapshot.Z, uint8(1))
	assert(cpuSnapshot.I, uint8(1))
	assert(cpuSnapshot.D, uint8(0))
	assert(cpuSnapshot.B, uint8(0))
	assert(cpuSnapshot.U, uint8(1))
	assert(cpuSnapshot.V, uint8(0))
	assert(cpuSnapshot.N, uint8(0))

	assert(ramSnapshot[0x0000], uint8(0x0A))
	assert(ramSnapshot[0x0001], uint8(0x03))
	assert(ramSnapshot[0x0002], uint8(0x1E))
	assert(ramSnapshot[0x0003], uint8(0x00))
	assert(ramSnapshot[0x0004], uint8(0x00))
	assert(ramSnapshot[0x0005], uint8(0xA2))
	assert(ramSnapshot[0x0006], uint8(0x03))
	assert(ramSnapshot[0x0007], uint8(0x8E))
	assert(ramSnapshot[0x0008], uint8(0x01))
	assert(ramSnapshot[0x0009], uint8(0x00))
	assert(ramSnapshot[0x000A], uint8(0xAC))
	assert(ramSnapshot[0x000B], uint8(0x00))
	assert(ramSnapshot[0x000C], uint8(0x00))
	assert(ramSnapshot[0x000D], uint8(0xA9))
	assert(ramSnapshot[0x000E], uint8(0x00))
	assert(ramSnapshot[0x000F], uint8(0x18))

	assert(ramSnapshot[0x0010], uint8(0x6D))
	assert(ramSnapshot[0x0011], uint8(0x01))
	assert(ramSnapshot[0x0012], uint8(0x00))
	assert(ramSnapshot[0x0013], uint8(0x88))
	assert(ramSnapshot[0x0014], uint8(0xD0))
	assert(ramSnapshot[0x0015], uint8(0xFA))
	assert(ramSnapshot[0x0016], uint8(0x8D))
	assert(ramSnapshot[0x0017], uint8(0x02))
	assert(ramSnapshot[0x0018], uint8(0x00))
	assert(ramSnapshot[0x0019], uint8(0xEA))
	assert(ramSnapshot[0x001A], uint8(0xEA))
	assert(ramSnapshot[0x001B], uint8(0xEA))
	assert(ramSnapshot[0x001C], uint8(0x00))
	assert(ramSnapshot[0x001D], uint8(0x00))
	assert(ramSnapshot[0x001E], uint8(0x00))
	assert(ramSnapshot[0x001F], uint8(0x00))

	fmt.Println("TestRegression1 complete!")
}
