package nes

type Mapper0 struct {
	prgRomBanks uint8
	chrRomBanks uint8
}

func NewMapper0(prgRomBanks, chrRomBanks uint8) *Mapper0 {
	return &Mapper0{
		prgRomBanks: prgRomBanks,
		chrRomBanks: chrRomBanks,
	}
}

func (m *Mapper0) CpuMapRead(addr uint16) (uint16, bool) {
	if addr >= 0x8000 {
		if m.prgRomBanks > 1 {
			return addr - 0x8000, true
		} else {
			if addr >= 0xC000 {
				return addr - 0xC000, true
			} else {
				return addr - 0x8000, true
			}
		}
	}
	return 0, false
}

func (m *Mapper0) CpuMapWrite(addr uint16) (uint16, bool) {
	if addr >= 0x8000 {
		if m.prgRomBanks > 1 {
			return addr - 0x8000, true
		} else {
			if addr >= 0xC000 {
				return addr - 0xC000, true
			} else {
				return addr - 0x8000, true
			}
		}
	}
	return 0, false
}

func (m *Mapper0) PpuMapRead(addr uint16) (uint16, bool) {
	if addr <= 0x1FFF {
		return addr, true
	}
	return 0, false

}

func (m *Mapper0) PpuMapWrite(addr uint16) (uint16, bool) {
	return 0, false
}
