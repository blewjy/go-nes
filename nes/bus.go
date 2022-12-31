package nes

type Bus struct {
	// Devices on the bus
	CPU        *CPU
	PPU        *PPU
	CpuRam     [2048]byte
	Cartridge  *Cartridge
	Controller uint8

	// Internal
	clockCounter uint64 // CPU only

	// Internal controller snapshot
	controllerState uint8
}

func NewBus() *Bus {
	bus := &Bus{}
	bus.CPU = NewCPU(bus)
	bus.CPU.Reset()
	bus.CpuRam = [2048]byte{}
	bus.PPU = NewPPU()
	return bus
}

func (b *Bus) Reset() {
	b.CPU.Reset()
	b.clockCounter = 0
}

func (b *Bus) Clock() {
	b.PPU.Clock()
	b.PPU.Clock()
	b.PPU.Clock()
	b.CPU.Clock()

	if b.PPU.nmi {
		b.PPU.nmi = false
		b.CPU.nmi()
	}

	b.clockCounter++
}

func (b *Bus) CpuRead(addr uint16) uint8 {
	var data uint8
	ok := false
	if b.Cartridge != nil {
		data, ok = b.Cartridge.CpuRead(addr)
	}
	if !ok {
		if addr <= 0x1FFF {
			data = b.CpuRam[addr&0x07FF]
		} else if addr >= 0x2000 && addr <= 0x3FFF {
			data = b.PPU.CpuRead(addr & 0x0007)
		} else if addr >= 0x4016 && addr <= 0x4017 {
			if b.controllerState&0x80 > 0 {
				data = 1
			}
			b.controllerState <<= 1
		}
	}
	return data
}

func (b *Bus) CpuWrite(addr uint16, data uint8) {
	ok := false
	if b.Cartridge != nil {
		ok = b.Cartridge.CpuWrite(addr, data)
	}
	if !ok {
		if addr <= 0x1FFF {
			b.CpuRam[addr&0x07FF] = data
		} else if addr >= 0x2000 && addr <= 0x3FFF {
			b.PPU.CpuWrite(addr&0x0007, data)
		} else if addr >= 0x4016 && addr <= 0x4017 {
			b.controllerState = b.Controller
		}
	}
}

func (b *Bus) InsertCartridge(cartridge *Cartridge) {
	b.Cartridge = cartridge
	b.PPU.ConnectCartridge(cartridge)
}
