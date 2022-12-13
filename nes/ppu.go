package nes

type PPU struct {
	Cartridge *Cartridge

	// PPU bus devices
	tableName    [2][1024]uint8
	tablePalette [32]uint8
}

func NewPPU() *PPU {
	return &PPU{}
}

func (p *PPU) CpuRead(addr uint16) uint8 {
	data := uint8(0x00)

	switch addr {
	case 0x0000: // Control
	case 0x0001: // Mask
	case 0x0002: // Status
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
	case 0x0006: // PPU Address
	case 0x0007: // PPU Data
	}

	return data
}

func (p *PPU) CpuWrite(addr uint16, data uint8) {
	switch addr {
	case 0x0000: // Control
	case 0x0001: // Mask
	case 0x0002: // Status
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
	case 0x0006: // PPU Address
	case 0x0007: // PPU Data
	}
}

func (p *PPU) PpuRead(addr uint16) uint8 {
	return 0
}

func (p *PPU) PpuWrite(addr uint16, data uint8) {}

func (p *PPU) ConnectCartridge(cartridge *Cartridge) {
	p.Cartridge = cartridge
}

func (p *PPU) Clock() {

}
