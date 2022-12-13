package nes

type Mapper interface {
	CpuMapRead(addr uint16) (uint16, bool)
	CpuMapWrite(addr uint16) (uint16, bool)
	PpuMapRead(addr uint16) (uint16, bool)
	PpuMapWrite(addr uint16) (uint16, bool)
}
