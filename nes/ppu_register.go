package nes

import "fmt"

func mustAssert(actual uint8, possible ...uint8) {
	for _, expected := range possible {
		if actual == expected {
			return
		}
	}
	panic(fmt.Sprintf("assertion failed: actual=%v, possible=%v", actual, possible))
}

func mustAssertInt(actual int, possible ...int) {
	for _, expected := range possible {
		if actual == expected {
			return
		}
	}
	panic(fmt.Sprintf("assertion failed: actual=%v, possible=%v", actual, possible))
}

type PpuCtrl uint8

func (p *PPU) GetBaseNametable() uint8 {
	result := uint8(p.ppuCtrl) & 0x03
	mustAssert(result, 0, 1, 2, 3)
	return result
}

func (p *PPU) GetVramAddrIncrement() uint8 {
	result := (uint8(p.ppuCtrl) & 0x04) >> 2
	mustAssert(result, 0, 1)
	return result
}

func (p *PPU) GetSpritePatternTableAddr() uint8 {
	result := (uint8(p.ppuCtrl) & 0x08) >> 3
	mustAssert(result, 0, 1)
	return result
}

func (p *PPU) GetBackgroundPatternTableAddr() uint8 {
	result := (uint8(p.ppuCtrl) & 0x10) >> 4
	mustAssert(result, 0, 1)
	return result
}

func (p *PPU) GetSpriteSize() uint8 {
	result := (uint8(p.ppuCtrl) & 0x20) >> 5
	mustAssert(result, 0, 1)
	return result
}

func (p *PPU) GetPpuMasterSlaveSelect() uint8 {
	result := (uint8(p.ppuCtrl) & 0x40) >> 6
	mustAssert(result, 0, 1)
	return result
}

func (p *PPU) GetNmiIndicator() uint8 {
	result := (uint8(p.ppuCtrl) & 0x80) >> 7
	mustAssert(result, 0, 1)
	return result
}

type PpuStatus uint8

func (p *PPU) GetSpriteOverflow() uint8 {
	result := (uint8(p.ppuStatus) & 0x20) >> 5
	mustAssert(result, 0, 1)
	return result
}

func (p *PPU) GetSpriteZeroHit() uint8 {
	result := (uint8(p.ppuStatus) & 0x40) >> 6
	mustAssert(result, 0, 1)
	return result
}

func (p *PPU) GetVerticalBlank() uint8 {
	result := (uint8(p.ppuStatus) & 0x80) >> 7
	mustAssert(result, 0, 1)
	return result
}

func (p *PPU) SetVerticalBlank(value uint8) {
	mustAssert(value, 0, 1)
	if value == 0 {
		p.ppuStatus = PpuStatus(uint8(p.ppuStatus) & 0x7F)
	} else if value == 1 {
		p.ppuStatus = PpuStatus(uint8(p.ppuStatus) | 0x80)
	}
}
