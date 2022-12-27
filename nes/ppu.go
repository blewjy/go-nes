/*
Info from javidx9's NES emulator video: https://www.youtube.com/watch?v=-THeUXqR3zY

PPU Registers:

0x2000 - PPUCTRL
Responsible for configuring the PPU to render in different ways

0x2001 - PPUMASK
Decides if backgrounds/sprites are being drawn, and what's happening at the edges of the screen

0x2002 - PPUSTATUS
Tells us when it is safe to render

0x2005 - PPUSCROLL
Scrolling information

0x2006 - PPUADDR
Allows CPU to read and write to PPU's address?

0x2007 - PPUDATA
Allows CPU to read and write to PPU's data?


How PPU renders a frame?

Each scanline can be thought of as a row of pixels horizontally across the screen.
Each pixel in that scanline can be thought of as a cycle, and there are 341 cycles per scanline.
However, the NES screen is only 256 pixels wide and 240 pixels high.
Therefore, each scanline will exceed the 256 width of the screen.

In the vertical axis, there are a total of 261 scanlines, which means that it exceeds the height of the screen by 21 lines.
This 21 lines is known as the vertical blank period.
The game needs to know when this period starts, because the CPU uses this area to do some processing to the PPU (because we can't see it on the screen).
It is typically during this period that CPU does some setup on the PPU to prepare for the next frame.

So the vertical blank bit in the PPUSTATUS tells us if we are in screen space (0) or blank space (1).

At the point of hitting the vertical blank, we also may emit an NMI signal to the CPU.
Whether or not this is emitted depends on the flag of the PPUCTRL register.

*/

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

	// PPU registers
	ppuCtrl   uint8 // 0x2000
	ppuMask   uint8 // 0x2001
	ppuStatus uint8 // 0x2002
	oamAddr   uint8 // 0x2003
	oamData   uint8 // 0x2004
	ppuScroll uint8 // 0x2005
	ppuAddr   uint8 // 0x2006
	ppuData   uint8 // 0x2007

	// ???
	oamDma uint8 // 0x4014

	// PPU helper variables
	addressLatch  uint8
	ppuDataBuffer uint8
	ppuAddress    uint16
	nmi           bool

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
		data = (p.ppuStatus & 0xE0) | (p.ppuDataBuffer & 0x1F) // quirk

		// reset vertical blank
		p.ppuStatus &= 0x3F

		// reset address latch
		p.addressLatch = 0

	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
	case 0x0006: // PPU Address
	case 0x0007: // PPU Data
		data = p.ppuDataBuffer
		p.ppuDataBuffer = p.PpuRead(p.ppuAddress)

		// for palette data, don't need to buffer one cycle
		if p.ppuAddress > 0x3F00 {
			data = p.ppuDataBuffer
		}

		// auto-increment
		p.ppuAddress++

	}

	return data
}

func (p *PPU) CpuWrite(addr uint16, data uint8) {
	switch addr {
	case 0x0000: // Control
		p.ppuCtrl = data
	case 0x0001: // Mask
		p.ppuMask = data
	case 0x0002: // Status
		// you can't write to this register
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
	case 0x0006: // PPU Address
		if p.addressLatch == 0 {
			p.ppuAddress = p.ppuAddress&0x00FF | (uint16(data) << 8)
			p.addressLatch = 1
		} else {
			p.ppuAddress = p.ppuAddress&0xFF00 | uint16(data)
			p.addressLatch = 0
		}
	case 0x0007: // PPU Data
		p.PpuWrite(p.ppuAddress, data)

		// auto-increment
		p.ppuAddress++
	}
}

func (p *PPU) PpuRead(addr uint16) uint8 {
	data, ok := p.Cartridge.PpuRead(addr)
	if !ok {
		if addr <= 0x0FFF {
			// pattern table 0
			index := addr & 0x0FFF
			data = p.tablePattern[0][index]

		} else if addr >= 0x1000 && addr <= 0x1FFF {
			// pattern table 1
			index := addr & 0x0FFF
			data = p.tablePattern[1][index]

		} else if addr >= 0x3F00 && addr <= 0x3FFF {
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
		if addr <= 0x0FFF {
			// pattern table 0
			index := addr & 0x0FFF
			p.tablePattern[0][index] = data

		} else if addr >= 0x1000 && addr <= 0x1FFF {
			// pattern table 1
			index := addr & 0x0FFF
			p.tablePattern[1][index] = data

		} else if addr >= 0x3F00 && addr <= 0x3FFF {
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

	if p.scanline == 0 && p.cycle == 1 {
		// set vertical blank
		p.ppuStatus &= 0x7F
	}

	if p.scanline == 241 && p.cycle == 1 {
		// set vertical blank
		p.ppuStatus |= 0x80

		// emit nmi if needed
		if p.ppuCtrl&0x80 > 0 {
			p.nmi = true
		}
	}

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

			display[i][j] = p.colorPalette[colorIndex]
		}
	}
	return display
}

func (p *PPU) GetPaletteDisplay() [32]color.Color {
	display := [32]color.Color{}

	for paletteId := 0; paletteId < 8; paletteId++ {
		for pixel := 0; pixel < 4; pixel++ {
			paletteByteOffset := 0x3F00 + (uint16(paletteId)<<2+uint16(pixel))&0x3F
			colorIndex := p.PpuRead(paletteByteOffset)

			display[paletteId*4+pixel] = p.colorPalette[colorIndex]
		}
	}

	return display
}
