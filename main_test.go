package main

import (
	"bufio"
	"fmt"
	"go-nes/emulator"
	"go-nes/nes"
	"os"
	"strconv"
	"strings"
	"testing"
)

func assert(a interface{}, b interface{}) {
	if a != b {
		panic(fmt.Sprintf("failed assertion: %v (%T) and %v (%T) are not equal", a, a, b, b))
	}
}

func TestBasic(t *testing.T) {
	fmt.Println("Running TestBasic...")

	program := "A20A8E0000A2038E0100AC0000A900186D010088D0FA8D0200EAEAEA"
	startAddr := uint16(0x0000)
	nes := emulator.NewEmulatorWithMode(emulator.Test)

	cpuSnapshot, ramSnapshot := nes.StartWithProgramAsTest(program, startAddr, 38)

	assert(cpuSnapshot.PC, uint16(0x0019))
	assert(cpuSnapshot.A, uint8(0x1E))
	assert(cpuSnapshot.X, uint8(0x03))
	assert(cpuSnapshot.Y, uint8(0x00))
	assert(cpuSnapshot.StackPtr, uint8(0xFD))
	assert(cpuSnapshot.P, uint8(0x26))

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

	fmt.Println("TestBasic complete!")
}

func TestNestest(t *testing.T) {
	fmt.Println("Running TestNestest...")

	// Prepare the nestest.txt log
	results := parseNestestLog()

	nes := emulator.NewEmulatorWithMode(emulator.Test)

	cpuSnapshot, _ := nes.StartWithNestestROMAsTest()
	for i := 0; i < 1000; i++ {
		fmt.Print(nes.PeekCurrentSnapshot())
		assert(cpuSnapshot, results[i])
		fmt.Println("\t\u2713")
		cpuSnapshot, _ = nes.ClockAsTest()
	}
	fmt.Println("TestNestest complete!")

}

func parseNestestLog() []nes.PeekCPUResult {
	file, err := os.Open("roms/nestest.txt")
	if err != nil {
		panic("Failed to open nestest.txt!")
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text []string
	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	file.Close()

	var results []nes.PeekCPUResult

	for _, line := range text {

		result := nes.PeekCPUResult{}

		var v uint64

		// PC
		PC := strings.Split(line, " ")[0]
		v, err = strconv.ParseUint(PC, 16, 16)
		if err != nil {
			panic(err)
		}
		result.PC = uint16(v)

		// A
		A := strings.Split(strings.Split(line, "A:")[1], " ")[0]
		v, err = strconv.ParseUint(A, 16, 8)
		if err != nil {
			panic(err)
		}
		result.A = uint8(v)

		// X
		X := strings.Split(strings.Split(line, "X:")[1], " ")[0]
		v, err = strconv.ParseUint(X, 16, 8)
		if err != nil {
			panic(err)
		}
		result.X = uint8(v)

		// Y
		Y := strings.Split(strings.Split(line, "Y:")[1], " ")[0]
		v, err = strconv.ParseUint(Y, 16, 8)
		if err != nil {
			panic(err)
		}
		result.Y = uint8(v)

		// P
		P := strings.Split(strings.Split(line, "P:")[1], " ")[0]
		v, err = strconv.ParseUint(P, 16, 8)
		if err != nil {
			panic(err)
		}
		result.P = uint8(v)

		// SP
		SP := strings.Split(strings.Split(line, "SP:")[1], " ")[0]
		v, err = strconv.ParseUint(SP, 16, 8)
		if err != nil {
			panic(err)
		}
		result.StackPtr = uint8(v)

		// CYC
		CYC := strings.Split(strings.Split(line, "CYC:")[1], " ")[0]
		v, err = strconv.ParseUint(CYC, 10, 64)
		if err != nil {
			panic(err)
		}
		result.Cycle = int(v)

		results = append(results, result)
	}

	return results
}
