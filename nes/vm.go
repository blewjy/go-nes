package nes

import (
	"encoding/hex"
)

type VM struct {
	bus *Bus
}

func NewVM() *VM {
	vm := &VM{}
	vm.bus = NewBus()
	return vm
}

func (v *VM) ForceSetResetVector(resetVector uint16) {
	v.bus.CpuWrite(0xFFFC, uint8(resetVector))
	v.bus.CpuWrite(0xFFFD, uint8(resetVector>>8))
	v.bus.Reset()
}

func (v *VM) LoadROM(filePath string) {
	v.bus.InsertCartridge(NewCartridge(filePath))
}

// LoadProgramAsString will load the given string as if it were a string of bytes.
// Also sets the given resetVector at 0xFFFC.
func (v *VM) LoadProgramAsString(program string, resetVector uint16) {
	data, err := hex.DecodeString(program)
	if err != nil {
		panic(err)
	}
	offset := resetVector
	for _, b := range data {
		v.bus.CpuWrite(offset, b)
		offset++
	}
	v.bus.CpuWrite(0xFFFC, uint8(resetVector))
	v.bus.CpuWrite(0xFFFD, uint8(resetVector>>8))

	v.bus.Reset()
}

// Step will clock the bus once.
func (v *VM) Step() {
	v.bus.Clock()
}

/** For debugging purposes only **/

// PeekCPUResult is a struct containing all registers of the CPU.
type PeekCPUResult struct {
	A        uint8
	X        uint8
	Y        uint8
	StackPtr uint8
	PC       uint16
	C        uint8
	Z        uint8
	I        uint8
	D        uint8
	B        uint8
	U        uint8
	V        uint8
	N        uint8
}

// PeekCPU returns a snapshot of the CPU registers as a PeekCPUResult.
func (v *VM) PeekCPU() PeekCPUResult {
	return PeekCPUResult{
		A:        v.bus.CPU.a,
		X:        v.bus.CPU.x,
		Y:        v.bus.CPU.y,
		StackPtr: v.bus.CPU.stackPtr,
		PC:       v.bus.CPU.pc,
		C:        v.bus.CPU.c,
		Z:        v.bus.CPU.z,
		I:        v.bus.CPU.i,
		D:        v.bus.CPU.d,
		B:        v.bus.CPU.b,
		U:        v.bus.CPU.u,
		V:        v.bus.CPU.v,
		N:        v.bus.CPU.n,
	}
}

// PeekRAM returns the contents of the CpuRam between the given start and end addresses (inclusive).
func (v *VM) PeekRAM(start, end uint16) []byte {
	var result []byte
	for i := start; i <= end; i++ {
		result = append(result, v.bus.CpuRead(i))
		if i == 0xFFFF {
			break
		}
	}
	return result
}
