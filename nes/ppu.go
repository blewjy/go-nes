package nes

import (
	"image/color"
	"math/rand"
)

type PPU struct {
	Cartridge *Cartridge

	// PPU stuff
	tablePattern [2][4096]uint8
	tableName    [2][1024]uint8
	tablePalette [32]uint8

	// ???
	scanline      int
	cycle         int
	frameComplete bool

	// All possible colors the NES can display
	colorPalette [0x40]color.Color

	// Temporary variable??
	screen [256][240]color.Color
}

func NewPPU() *PPU {
	screen := [256][240]color.Color{}
	for i := 0; i < 256; i++ {
		screen[i] = [240]color.Color{}
		for j := 0; j < 240; j++ {
			screen[i][j] = color.Black
		}
	}
	return &PPU{
		screen:       screen,
		colorPalette: nesColorPalette,
	}
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
	data, ok := p.Cartridge.PpuRead(addr)
	if !ok {
		if addr >= 0x3F00 && addr <= 0x3FFF {
			addr &= 0x001F
			if addr == 0x0010 {
				addr = 0x0000
			}
			if addr == 0x0014 {
				addr = 0x0004
			}
			if addr == 0x0018 {
				addr = 0x0008
			}
			if addr == 0x001C {
				addr = 0x000C
			}
			data = p.tablePalette[addr]
		}
	}
	return data
}

func (p *PPU) PpuWrite(addr uint16, data uint8) {
	ok := p.Cartridge.PpuWrite(addr, data)
	if !ok {
		if addr >= 0x3F00 && addr <= 0x3FFF {
			addr &= 0x001F
			if addr == 0x0010 {
				addr = 0x0000
			}
			if addr == 0x0014 {
				addr = 0x0004
			}
			if addr == 0x0018 {
				addr = 0x0008
			}
			if addr == 0x001C {
				addr = 0x000C
			}
			p.tablePalette[addr] = data
		}
	}
}

func (p *PPU) ConnectCartridge(cartridge *Cartridge) {
	p.Cartridge = cartridge
}

func (p *PPU) GetScreen() [256][240]color.Color {
	return p.screen
}

func (p *PPU) Clock() {
	// At each clock of the PPU, we will render a pixel to the screen at the current scanline and cycle.
	// Cycle is like the X coordinate and scanline is like the Y coordinate
	p.screen[p.cycle%256][p.scanline%240] = p.colorPalette[rand.Intn(64)]

	p.cycle++
	if p.cycle >= 341 {
		p.cycle = 0
		p.scanline++
		if p.scanline >= 261 {
			p.scanline = 0
			p.frameComplete = true
		}
	}
}

func (p *PPU) GetPatternTableDisplay(tableIndex, paletteId int) [128][128]color.Color {
	display := [128][128]color.Color{}
	for i := 0; i < 128; i++ {
		display[i] = [128]color.Color{}
		for j := 0; j < 128; j++ {
			tileX := j / 8
			tileY := i / 8
			tileByteOffset := uint16(tileX*16 + tileY*256)
			pixelX := j % 8
			pixelY := i % 8
			lsbPixelByteOffset := tileByteOffset + uint16(pixelY)
			msbPixelByteOffset := tileByteOffset + uint16(pixelY) + 8

			lsbPixelByte := p.PpuRead(0x1000*uint16(tableIndex) + lsbPixelByteOffset)
			msbPixelByte := p.PpuRead(0x1000*uint16(tableIndex) + msbPixelByteOffset)

			lsbPixelBit := (lsbPixelByte >> (7 - pixelX)) & 1
			msbPixelBit := (msbPixelByte >> (7 - pixelX)) & 1

			pixelBits := msbPixelBit<<1 | lsbPixelBit

			paletteByteOffset := 0x3F00 + uint16(paletteId)<<2 + uint16(pixelBits)

			colorIndex := p.PpuRead(paletteByteOffset)
			colorIndex = pixelBits * 21

			display[i][j] = p.colorPalette[colorIndex]
		}
	}
	return display
}
