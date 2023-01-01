package main

import "go-nes/emulator"

func main() {
	//emulator.NewEmulator().StartWithROM("roms/Super Mario Bros. (World).nes")
	//emulator.NewEmulator().StartWithROM("roms/DuckTales (USA).nes")
	//emulator.NewEmulator().StartWithROM("roms/Ice Climber (U).nes")
	emulator.NewEmulator().StartWithROM("roms/Donkey Kong (World) (Rev A).nes")
	//emulator.NewEmulatorWithMode(emulator.Automation).StartWithROM("roms/nestest.nes")
	//emulator.NewEmulator().StartWithROM("roms/nestest.nes")

	//program := "A20A8E0000A2038E0100AC0000A900186D010088D0FA8D0200EAEAEA"
	//startAddr := uint16(0x1000)
	//emulator.NewEmulator().StartWithProgram(program, startAddr)
}
