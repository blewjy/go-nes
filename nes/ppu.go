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
	return 0
}

func (p *PPU) PpuWrite(addr uint16, data uint8) {}

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
