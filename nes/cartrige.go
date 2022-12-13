// Header Reference: https://www.nesdev.org/wiki/INES

package nes

import (
	"fmt"
	"os"
)

type Cartridge struct {
	prgRomData []byte
	chrRomData []byte

	variant     iNESVariant
	mapperId    uint8
	prgRomBanks uint8
	chrRomBanks uint8

	mapper Mapper
}

func NewCartridge(filePath string) *Cartridge {
	// Init cartridge
	cartridge := &Cartridge{}

	// Open ROM file
	f, err := os.Open(filePath)
	if err != nil {
		panic(fmt.Sprintf("Error while reading cartridge file: %v", err))
	}
	defer f.Close()

	// Parse header
	headerData := make([]byte, 16)
	n, err := f.Read(headerData)
	if err != nil {
		panic(fmt.Sprintf("Error while reading header data: %v", err))
	}
	if n != 16 {
		panic(fmt.Sprintf("Expected header size (16) not equal to number of bytes read (%v)", n))
	}
	type header struct {
		name           [4]uint8 // Constant $4E $45 $53 $1A [78 69 83 26] ("NES" followed by MS-DOS end-of-file)
		prgRomSize     uint8    // Size of PRG ROM in 16 KB units
		chrRomSize     uint8    // Size of CHR ROM in 8 KB units (Value 0 means the board uses CHR RAM)
		mapper         uint8    // Flags 6 - Mapper, mirroring, battery, trainer
		mapper2        uint8    // Flags 7 - Mapper, VS/Playchoice, NES 2.0
		prgRamSize     uint8    // Flags 8 - PRG-RAM size (rarely used extension)
		tvSystem       uint8    // Flags 9 - TV system (rarely used extension)
		tvSystemPrgRam uint8    // Flags 10 - TV system, PRG-RAM presence (unofficial, rarely used extension)
		unused         [5]uint8 // Unused padding (should be filled with zero, but some rippers put their name across bytes 7-15)
	}
	h := header{
		name:           [4]uint8{headerData[0], headerData[1], headerData[2], headerData[3]},
		prgRomSize:     headerData[4],
		chrRomSize:     headerData[5],
		mapper:         headerData[6],
		mapper2:        headerData[7],
		prgRamSize:     headerData[8],
		tvSystem:       headerData[9],
		tvSystemPrgRam: headerData[10],
		unused:         [5]uint8{headerData[11], headerData[12], headerData[13], headerData[14], headerData[15]},
	}

	// Parse the mapperId ID
	cartridge.mapperId = (h.mapper2>>4)<<4 | h.mapper>>4

	// Parse the variant of the iNES file
	cartridge.variant = parseVariant(headerData)

	switch cartridge.variant {
	case Archaic, NES2:
		panic("Variant not supported yet!")
	case iNES:
		cartridge.prgRomBanks = h.prgRomSize
		cartridge.chrRomBanks = h.chrRomSize

		// TODO: Right now, we are just assuming that trainer data is not there. We should actually check it via Flag 6.

		// Read PRG ROM data
		prgRomSize := int(cartridge.prgRomBanks) * 16384
		cartridge.prgRomData = make([]byte, prgRomSize)
		n, err = f.Read(cartridge.prgRomData)
		if err != nil {
			panic(fmt.Sprintf("Error while reading PRG ROM data: %v", err))
		}
		if n != prgRomSize {
			panic(fmt.Sprintf("Expected PRG ROM size (%v) not equal to number of bytes read (%v)", prgRomSize, n))
		}

		// Read CHR ROM data
		chrRomSize := int(cartridge.chrRomBanks) * 8192
		cartridge.chrRomData = make([]byte, chrRomSize)
		n, err = f.Read(cartridge.chrRomData)
		if err != nil {
			panic(fmt.Sprintf("Error while reading CHR ROM data: %v", err))
		}
		if n != chrRomSize {
			panic(fmt.Sprintf("Expected CHR ROM size (%v) not equal to number of bytes read (%v)", chrRomSize, n))
		}

		// Load mapperId
		switch cartridge.mapperId {
		case 0:
			cartridge.mapper = NewMapper0(cartridge.prgRomBanks, cartridge.chrRomBanks)
		default:
			panic(fmt.Sprintf("Mapper %v not yet supported!", cartridge.mapperId))
		}
	}

	return cartridge
}

type iNESVariant uint8

const (
	Archaic iNESVariant = iota
	iNES
	NES2
)

func parseVariant(data []byte) iNESVariant {
	return iNES
}

func (c *Cartridge) CpuRead(addr uint16) (uint8, bool) {
	mappedAddr, ok := c.mapper.CpuMapRead(addr)
	if !ok {
		return 0, false
	}
	return c.prgRomData[mappedAddr], true
}

func (c *Cartridge) CpuWrite(addr uint16, data uint8) bool {
	mappedAddr, ok := c.mapper.CpuMapWrite(addr)
	if !ok {
		return false
	}
	c.prgRomData[mappedAddr] = data
	return true
}

func (c *Cartridge) PpuRead(addr uint16) (uint8, bool) {
	mappedAddr, ok := c.mapper.PpuMapRead(addr)
	if !ok {
		return 0, false
	}
	return c.chrRomData[mappedAddr], true
}

func (c *Cartridge) PpuWrite(addr uint16, data uint8) bool {
	mappedAddr, ok := c.mapper.PpuMapWrite(addr)
	if !ok {
		return false
	}
	c.chrRomData[mappedAddr] = data
	return true
}
