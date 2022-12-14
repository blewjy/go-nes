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
)

type PPU struct {
	Cartridge *Cartridge

	// PPU stuff
	tablePattern [2][4096]uint8
	tableName    [2][1024]uint8
	tablePalette [32]uint8

	// PPU registers
	ppuCtrl   PpuCtrl   // 0x2000
	ppuMask   uint8     // 0x2001
	ppuStatus PpuStatus // 0x2002
	oamAddr   uint8     // 0x2003
	oamData   uint8     // 0x2004
	ppuScroll uint8     // 0x2005
	ppuAddr   uint8     // 0x2006
	ppuData   uint8     // 0x2007

	// PPU Object Attribute Memory (OAM)
	ppuOam [256]uint8

	// ???
	oamDma uint8 // 0x4014

	// PPU helper variables
	addressLatch  uint8
	ppuDataBuffer uint8
	ppuAddress    uint16
	nmi           bool

	// more PPU helpers
	vramAddr     uint16
	tramAddr     uint16
	fineScrollX  uint8
	camPositionX uint8
	camPositionY uint8

	patternShiftHi uint8
	patternShiftLo uint8
	paletteShiftHi uint8
	paletteShiftLo uint8

	spriteActivated   [8]bool
	spriteShiftHi     [8]uint8
	spriteShiftLo     [8]uint8
	spriteAttrInfo    [8]uint8
	spriteCounters    [8]uint8
	spriteRenderCount [8]uint8

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
			screen[i][j] = color.RGBA{255, 0, 0, 255}
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
		data = (uint8(p.ppuStatus & 0xE0)) | (p.ppuDataBuffer & 0x1F) // quirk

		// reset vertical blank
		p.SetVerticalBlank(0)

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

		// auto-increment based on ctrl address increment flag
		if p.GetVramAddrIncrement() == 1 {
			p.ppuAddress += 32
		} else {
			p.ppuAddress += 1
		}
	}

	return data
}

func (p *PPU) CpuWrite(addr uint16, data uint8) {
	switch addr {
	case 0x0000: // Control
		p.ppuCtrl = PpuCtrl(data)
	case 0x0001: // Mask
		p.ppuMask = data
	case 0x0002: // Status
		// you can't write to this register
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
		if p.addressLatch == 0 {
			p.camPositionX = data
			p.addressLatch = 1
		} else {
			p.camPositionY = data
			p.addressLatch = 0
		}
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

		// auto-increment based on ctrl address increment flag
		if p.GetVramAddrIncrement() == 1 {
			p.ppuAddress += 32
		} else {
			p.ppuAddress += 1
		}
	}
}

func (p *PPU) PpuOamWrite(offset, data uint8) {
	p.ppuOam[offset] = data
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

		} else if addr >= 0x2000 && addr <= 0x3EFF {
			// name tables
			if p.Cartridge.mirrorMode == Vertical {
				if addr >= 0x2000 && addr <= 0x23FF {
					data = p.tableName[0][addr&0x03FF]
				} else if addr >= 0x2400 && addr <= 0x27FF {
					data = p.tableName[1][addr&0x03FF]
				} else if addr >= 0x2800 && addr <= 0x3BFF {
					data = p.tableName[0][addr&0x03FF]
				} else if addr >= 0x3C00 && addr <= 0x3EFF {
					data = p.tableName[1][addr&0x03FF]
				}
			} else if p.Cartridge.mirrorMode == Horizontal {
				if addr >= 0x2000 && addr <= 0x23FF {
					data = p.tableName[0][addr&0x03FF]
				} else if addr >= 0x2400 && addr <= 0x27FF {
					data = p.tableName[0][addr&0x03FF]
				} else if addr >= 0x2800 && addr <= 0x3BFF {
					data = p.tableName[1][addr&0x03FF]
				} else if addr >= 0x3C00 && addr <= 0x3EFF {
					data = p.tableName[1][addr&0x03FF]
				}
			}

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

		} else if addr >= 0x2000 && addr <= 0x3EFF {
			// name tables
			if p.Cartridge.mirrorMode == Vertical {
				if addr >= 0x2000 && addr <= 0x23FF {
					p.tableName[0][addr&0x03FF] = data
				} else if addr >= 0x2400 && addr <= 0x27FF {
					p.tableName[1][addr&0x03FF] = data
				} else if addr >= 0x2800 && addr <= 0x3BFF {
					p.tableName[0][addr&0x03FF] = data
				} else if addr >= 0x3C00 && addr <= 0x3EFF {
					p.tableName[1][addr&0x03FF] = data
				}
			} else if p.Cartridge.mirrorMode == Horizontal {
				if addr >= 0x2000 && addr <= 0x23FF {
					p.tableName[0][addr&0x03FF] = data
				} else if addr >= 0x2400 && addr <= 0x27FF {
					p.tableName[0][addr&0x03FF] = data
				} else if addr >= 0x2800 && addr <= 0x3BFF {
					p.tableName[1][addr&0x03FF] = data
				} else if addr >= 0x3C00 && addr <= 0x3EFF {
					p.tableName[1][addr&0x03FF] = data
				}
			}

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

func (p *PPU) GetAttributeTable(tableIndex int) []uint8 {
	return p.tableName[tableIndex][960:]
}

func (p *PPU) Clock() {

	if p.scanline == 261 && p.cycle == 1 {
		// set vertical blank
		p.SetVerticalBlank(0)
	}

	if p.scanline == 261 && p.cycle >= 329 && p.cycle <= 336 {
		p.fetchNextTileData()
	}

	if p.scanline >= 0 && p.scanline <= 239 {
		if (p.cycle >= 1 && p.cycle <= 256) || (p.cycle >= 329 && p.cycle <= 336) {
			p.fetchNextTileData()
		}
	}

	if p.scanline == 241 && p.cycle == 1 {
		// set vertical blank
		p.SetVerticalBlank(1)

		// emit nmi if needed
		if p.GetNmiIndicator() == 1 {
			p.nmi = true
		}
	}

	p.cycle++
	if p.cycle > 340 {
		p.cycle = 0
		p.scanline++
		if p.scanline > 261 {
			p.scanline = 0
			p.frameComplete = true
		}
	}
}

func (p *PPU) fetchNextTileData() {
	x := p.cycle - 1
	y := p.scanline
	nextX := x + 1
	nextY := y

	if nextX > 255 {
		nextX = 0
		nextY += 1
		if nextY == 262 {
			nextY = 0
		}
		if nextY > 239 {
			return
		}
	}

	if nextX > 255 || nextY > 239 {
		panic("invalid nextX or nextY")
	}

	nextTileX := nextX / 8
	nextTileY := nextY / 8

	//mustAssertInt(nextTileX, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31)
	//mustAssertInt(nextTileY, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29)

	nextTilePixelOffsetX := nextX % 8
	nextTilePixelOffsetY := nextY % 8

	//mustAssertInt(nextTilePixelOffsetX, 0, 1, 2, 3, 4, 5, 6, 7)
	//mustAssertInt(nextTilePixelOffsetY, 0, 1, 2, 3, 4, 5, 6, 7)

	nextTileNametableIndex := uint16(nextTileX) + uint16(nextTileY)*32
	nextTileNametableByte := p.PpuRead(0x2000 + (0x400 * uint16(p.GetBaseNametable())) + nextTileNametableIndex)

	nextTilePatternTableByteOffset := uint16(nextTileNametableByte) * 16

	if nextTilePatternTableByteOffset > 4080 {
		panic("invalid nextTilePatternTableByteOffset")
	}

	nextTilePixelByteOffsetLsb := nextTilePatternTableByteOffset + uint16(nextTilePixelOffsetY)
	nextTilePixelByteOffsetMsb := nextTilePatternTableByteOffset + uint16(nextTilePixelOffsetY) + 8

	if nextTilePixelByteOffsetLsb >= 0x1000 || nextTilePixelByteOffsetMsb >= 0x1000 {
		panic("invalid next pixel byte offset")
	}

	nextTilePixelByteLsb := p.PpuRead(0x1000*uint16(p.GetBackgroundPatternTableAddr()) + nextTilePixelByteOffsetLsb)
	nextTilePixelByteMsb := p.PpuRead(0x1000*uint16(p.GetBackgroundPatternTableAddr()) + nextTilePixelByteOffsetMsb)

	nextTilePixelBitLsb := (nextTilePixelByteLsb >> (7 - nextTilePixelOffsetX)) & 1
	nextTilePixelBitMsb := (nextTilePixelByteMsb >> (7 - nextTilePixelOffsetX)) & 1

	nextTilePixelBits := nextTilePixelBitMsb<<1 | nextTilePixelBitLsb

	// attribute table info
	nextAttrTileX := nextTileX / 4
	nextAttrTileY := nextTileY / 4

	//mustAssertInt(nextAttrTileX, 0, 1, 2, 3, 4, 5, 6, 7)
	//mustAssertInt(nextAttrTileY, 0, 1, 2, 3, 4, 5, 6, 7)

	nextAttrTileByteOffset := nextAttrTileX + nextAttrTileY*8

	if nextAttrTileByteOffset > 63 {
		panic("invalid nextAttrTileByteOffset")
	}

	nextAttrTileByte := p.PpuRead(0x23C0 + uint16(nextAttrTileByteOffset))
	nextAttrTileInnerOffsetX := (nextTileX % 4) / 2
	nextAttrTileInnerOffsetY := (nextTileY % 4) / 2

	var nextAttrTileInfo uint8
	if nextAttrTileInnerOffsetX == 0 && nextAttrTileInnerOffsetY == 0 {
		nextAttrTileInfo = (nextAttrTileByte >> 0) & 0x3
	} else if nextAttrTileInnerOffsetX == 1 && nextAttrTileInnerOffsetY == 0 {
		nextAttrTileInfo = (nextAttrTileByte >> 2) & 0x3
	} else if nextAttrTileInnerOffsetX == 0 && nextAttrTileInnerOffsetY == 1 {
		nextAttrTileInfo = (nextAttrTileByte >> 4) & 0x3
	} else if nextAttrTileInnerOffsetX == 1 && nextAttrTileInnerOffsetY == 1 {
		nextAttrTileInfo = (nextAttrTileByte >> 6) & 0x3
	}

	mustAssert(nextAttrTileInfo, 0, 1, 2, 3)

	nextAttrTileInfoLsb := (nextAttrTileInfo >> 0) & 0x1
	nextAttrTileInfoMsb := (nextAttrTileInfo >> 1) & 0x1

	nextAttrTileInfo = nextAttrTileInfoMsb<<1 | nextAttrTileInfoLsb

	paletteByteOffset := uint16(nextAttrTileInfo)<<2 | uint16(nextTilePixelBits)

	paletteByteAddr := 0x3F00 + paletteByteOffset

	colorIndex := p.PpuRead(paletteByteAddr)

	/** sprites **/

	if x == 0 {

		p.spriteActivated = [8]bool{false, false, false, false, false, false, false, false}
		p.spriteShiftLo = [8]uint8{0, 0, 0, 0, 0, 0, 0, 0}
		p.spriteShiftHi = [8]uint8{0, 0, 0, 0, 0, 0, 0, 0}
		p.spriteCounters = [8]uint8{255, 255, 255, 255, 255, 255, 255, 255}
		p.spriteRenderCount = [8]uint8{0, 0, 0, 0, 0, 0, 0, 0}

		spriteIndex := 0

		for i := 0; i < 256; i += 4 {
			spritePositionY := p.ppuOam[i] - 1
			spriteTileIndexNumber := p.ppuOam[i+1]
			spriteTileAttributes := p.ppuOam[i+2]
			spritePositionX := p.ppuOam[i+3]

			//fmt.Printf("row %d: sprite pos (%d, %d), sprite index (%d), sprite attr (%d)\n", y, spritePositionX, spritePositionY, spriteTileIndexNumber, spriteTileAttributes)

			if uint8(y) >= spritePositionY && uint8(y) <= spritePositionY+7 {
				spritePixelRowOffsetFromTop := uint8(y) - spritePositionY

				spritePatternTableByteOffset := uint16(spriteTileIndexNumber) * 16

				if spritePatternTableByteOffset > 4080 {
					panic("invalid spritePatternTableByteOffset")
				}

				spritePixelByteOffsetLsb := spritePatternTableByteOffset + uint16(spritePixelRowOffsetFromTop)
				spritePixelByteOffsetMsb := spritePatternTableByteOffset + uint16(spritePixelRowOffsetFromTop) + 8

				if spritePixelByteOffsetLsb >= 0x1000 || spritePixelByteOffsetMsb >= 0x1000 {
					panic("invalid next sprite pixel byte offset")
				}

				p.spriteShiftLo[spriteIndex] = p.PpuRead(0x1000*uint16(p.GetSpritePatternTableAddr()) + spritePixelByteOffsetLsb)
				p.spriteShiftHi[spriteIndex] = p.PpuRead(0x1000*uint16(p.GetSpritePatternTableAddr()) + spritePixelByteOffsetMsb)
				p.spriteAttrInfo[spriteIndex] = spriteTileAttributes
				p.spriteCounters[spriteIndex] = spritePositionX + 1

				spriteIndex++
			}

			if spriteIndex > 7 {
				break
			}
		}
	}

	done := false
	for i := 0; i < 8; i++ {
		if p.spriteCounters[i] == 255 {
			continue
		}

		if p.spriteCounters[i] != 0 {
			p.spriteCounters[i]--
		}

		if p.spriteCounters[i] <= 0 {
			p.spriteActivated[i] = true
		}

		if !done && p.spriteActivated[i] && p.spriteRenderCount[i] < 8 {
			spritePixelBitLsb := (p.spriteShiftLo[i] & 0x80) >> 7
			spritePixelBitMsb := (p.spriteShiftHi[i] & 0x80) >> 7
			p.spriteShiftLo[i] <<= 1
			p.spriteShiftHi[i] <<= 1

			spritePixelBits := spritePixelBitMsb<<1 | spritePixelBitLsb

			spritePaletteByteOffset := uint16(p.spriteAttrInfo[i])<<2 | uint16(spritePixelBits)
			spritePaletteByteAddr := 0x3F10 + spritePaletteByteOffset
			colorIndex = p.PpuRead(spritePaletteByteAddr)

			done = true

			p.spriteRenderCount[i]++
		}
	}

	// for each 8 sprite
	// each cycle, decrement sprite's x position
	// if sprite x == 0, render the earliest pixel in the 8 pixels shift register, then shift that register

	p.screen[nextX][nextY] = p.colorPalette[colorIndex]
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

			if pixelBits >= 4 {
				panic("no no no no nooo~")
			}

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
