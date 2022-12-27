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

type PpuCtrl uint8

func (r PpuCtrl) GetBaseNametable() uint8 {
	result := uint8(r) & 0x03
	mustAssert(result, 0, 1, 2, 3)
	return result
}

func (r PpuCtrl) GetVramAddrIncrement() uint8 {
	result := (uint8(r) & 0x04) >> 2
	mustAssert(result, 0, 1)
	return result
}

func (r PpuCtrl) GetSpritePatternTableAddr() uint8 {
	result := (uint8(r) & 0x08) >> 3
	mustAssert(result, 0, 1)
	return result
}

func (r PpuCtrl) GetBackgroundPatternTableAddr() uint8 {
	result := (uint8(r) & 0x10) >> 4
	mustAssert(result, 0, 1)
	return result
}

func (r PpuCtrl) GetSpriteSize() uint8 {
	result := (uint8(r) & 0x20) >> 5
	mustAssert(result, 0, 1)
	return result
}

func (r PpuCtrl) GetPpuMasterSlaveSelect() uint8 {
	result := (uint8(r) & 0x40) >> 6
	mustAssert(result, 0, 1)
	return result
}

func (r PpuCtrl) GetNmiIndicator() uint8 {
	result := (uint8(r) & 0x80) >> 7
	mustAssert(result, 0, 1)
	return result
}
