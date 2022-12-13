package emulator

import (
	"go-nes/nes"
)

// StartWithProgramAsTest will load the given program into the emulator, then clock the given number of cycles.
// It will return the resulting PeekCPUResult and the CpuRam as an array of bytes.
func (e *Emulator) StartWithProgramAsTest(program string, startAddr uint16, cycles int) (nes.PeekCPUResult, []byte) {
	if e.Mode != Test {
		panic("Cannot start emulator as test: emulator is not in test mode!")
	}

	e.VM.LoadProgramAsString(program, startAddr)

	for i := 0; i < cycles; i++ {
		e.VM.Step()
	}

	return e.VM.PeekCPU(), e.VM.PeekRAM(0x0000, 0x07FF)
}

func (e *Emulator) StartWithNestestROMAsTest() (nes.PeekCPUResult, []byte) {
	if e.Mode != Test {
		panic("Cannot start emulator as test: emulator is not in test mode!")
	}

	e.VM.LoadROM("roms/nestest.nes")
	e.VM.ForceSetResetVector(0xC000)

	return e.VM.PeekCPU(), e.VM.PeekRAM(0x0000, 0x07FF)
}

func (e *Emulator) ClockAsTest() (nes.PeekCPUResult, []byte) {
	e.VM.Step()
	return e.VM.PeekCPU(), e.VM.PeekRAM(0x0000, 0x07FF)
}

func (e *Emulator) PeekCurrentSnapshot() string {
	return e.VM.PeekCPUSnapshot()
}
