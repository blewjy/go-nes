package nes

import (
	"encoding/hex"
	"image/color"
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

func (v *VM) Reset() {
	v.bus.Reset()
}

// Step will clock the bus once.
func (v *VM) Step() {
	v.bus.Clock()
}

// StepFrame will clock the bus until 1 frame is complete
func (v *VM) StepFrame() {
	for !v.bus.PPU.frameComplete {
		v.bus.Clock()
	}
	v.bus.PPU.frameComplete = false
}

func (v *VM) GetScreen() [256][240]color.Color {
	return v.bus.PPU.GetScreen()
}

func (v *VM) GetPPUNametable() [1024]uint8 {
	return v.bus.PPU.tableName[0]
}

func (v *VM) GetPatternTableDisplay(tableIndex, paletteId int) [128][128]color.Color {
	return v.bus.PPU.GetPatternTableDisplay(tableIndex, paletteId)
}

func (v *VM) GetPaletteDisplay() [32]color.Color {
	return v.bus.PPU.GetPaletteDisplay()
}

/** For debugging purposes only **/

// PeekCPUResult is a struct containing all registers of the CPU.
type PeekCPUResult struct {
	A        uint8
	X        uint8
	Y        uint8
	StackPtr uint8
	PC       uint16
	P        uint8
	Cycle    int
}

// PeekCPU returns a snapshot of the CPU registers as a PeekCPUResult.
func (v *VM) PeekCPU() PeekCPUResult {
	return PeekCPUResult{
		A:        v.bus.CPU.a,
		X:        v.bus.CPU.x,
		Y:        v.bus.CPU.y,
		StackPtr: v.bus.CPU.sp,
		PC:       v.bus.CPU.pc,
		P:        v.bus.CPU.GetStatus(),
		Cycle:    v.bus.CPU.cycle,
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

func (v *VM) PeekCPUSnapshot() string {
	return v.bus.CPU.PeekCurrentSnapshot()
}

func (v *VM) PeekDisassembly() map[uint16]string {
	return v.bus.CPU.PeekDisassembly()
}
