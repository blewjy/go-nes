package emulator

import (
	"fmt"
	"go-nes/nes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	windowWidth  = 1280
	windowHeight = 1240
	windowScale  = 3

	cpuClockSpeed = 1789773
)

var (
	screenImage               = ebiten.NewImage(256, 240)
	paletteTableImage         = ebiten.NewImage(256, 8)
	paletteSelectedBoxImage   = ebiten.NewImage(32, 8)
	patternTableImage         = ebiten.NewImage(128, 128)
	disassemblyHighlightImage = ebiten.NewImage(144, 12)

	debugPatternId = 0
)

type Mode string

const (
	Normal     Mode = "normal"
	Test       Mode = "test"
	Automation Mode = "automation"
)

type State string

const (
	Init     State = "init"
	Running  State = "running"
	Stepping State = "stepping"
	Paused   State = "paused"
)

type Emulator struct {
	// Core
	VM *nes.VM

	// Settings
	Mode  Mode
	State State

	// Debugging controls
	IsKeyPressed bool

	// Debugging info
	Disassembly map[uint16]string
}

func NewEmulator() *Emulator {
	return &Emulator{
		VM: nes.NewVM(),

		Mode:  Normal,
		State: Init,
	}
}

func NewEmulatorWithMode(mode Mode) *Emulator {
	return &Emulator{
		VM: nes.NewVM(),

		Mode:  mode,
		State: Init,
	}
}

// Update will run at exactly 60Hz
func (e *Emulator) Update() error {
	ebiten.SetWindowTitle(fmt.Sprintf("NES Emulator in Go! TPS: %v FPS: %v", ebiten.ActualTPS(), ebiten.ActualFPS()))
	switch e.State {
	case Init:
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			e.State = Stepping
		}
	case Running:
		for i := 0; i < cpuClockSpeed/60; i++ {
			e.VM.Step()
		}

		if ebiten.IsKeyPressed(ebiten.KeyP) {
			e.State = Paused
		}
		if ebiten.IsKeyPressed(ebiten.KeyF1) {
			e.State = Stepping
		}

		if !e.IsKeyPressed && ebiten.IsKeyPressed(ebiten.KeyRight) {
			e.IsKeyPressed = true
			debugPatternId = (debugPatternId + 1) % 28
			fmt.Println("debugPatternId", debugPatternId)
		}
		if !e.IsKeyPressed && ebiten.IsKeyPressed(ebiten.KeyLeft) {
			e.IsKeyPressed = true
			debugPatternId = (debugPatternId - 1) % 28
			if debugPatternId < 0 {
				debugPatternId = 27
			}
			fmt.Println("debugPatternId", debugPatternId)
		}

		if !ebiten.IsKeyPressed(ebiten.KeyRight) && !ebiten.IsKeyPressed(ebiten.KeyLeft) {
			e.IsKeyPressed = false
		}
	case Stepping:
		if !e.IsKeyPressed && ebiten.IsKeyPressed(ebiten.KeySpace) {
			e.IsKeyPressed = true
			e.VM.Step()
		}
		if !e.IsKeyPressed && ebiten.IsKeyPressed(ebiten.KeyF) {
			e.IsKeyPressed = true
			e.VM.StepFrame()
		}
		if !ebiten.IsKeyPressed(ebiten.KeySpace) && !ebiten.IsKeyPressed(ebiten.KeyF) {
			e.IsKeyPressed = false
		}
		if ebiten.IsKeyPressed(ebiten.KeyF2) {
			e.State = Running
		}
	case Paused:
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			e.State = Running
		}

	}

	return nil
}

func (e *Emulator) Draw(screen *ebiten.Image) {
	e.DrawScreenAt(screen, 8, 8)
	e.DrawPaletteTableAt(screen, 8, 256)
	e.DrawPatternTableAt(screen, 8, 272)
	e.DrawStateAt(screen, 272, 8)
	e.DrawCpuAt(screen, 272, 36)
	e.DrawDisassemblyAt(screen, 272, 128)
}

func (e *Emulator) DrawDisassemblyAt(screen *ebiten.Image, x, y int) {
	cpu := e.VM.PeekCPU()

	ebitenutil.DebugPrintAt(screen, "Disassembly:", x, y)

	yOffset := 10

	// print middle
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y+yOffset*12+2+4))
	screen.DrawImage(disassemblyHighlightImage, op)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("0x%04X: %s", cpu.PC, e.Disassembly[cpu.PC]), x, y+yOffset*12+4)

	// move up
	for i, o := 0, uint16(1); i < yOffset-1; i, o = i+1, o+1 {
		for {
			if _, ok := e.Disassembly[cpu.PC-o]; !ok {
				o++
			} else {
				break
			}
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("0x%04X: %s", cpu.PC-o, e.Disassembly[cpu.PC-o]), x, y+(yOffset-1-i)*12+4)
	}

	// move down
	for i, o := 0, uint16(1); i < yOffset-1; i, o = i+1, o+1 {
		for {
			if _, ok := e.Disassembly[cpu.PC+o]; !ok {
				o++
			} else {
				break
			}
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("0x%04X: %s", cpu.PC+o, e.Disassembly[cpu.PC+o]), x, y+(yOffset+1+i)*12+4)
	}

}

func (e *Emulator) DrawPaletteTableAt(screen *ebiten.Image, x, y int) {
	display := e.VM.GetPaletteDisplay()

	var pixels []byte
	for row := 0; row < 8; row++ {
		for p := 0; p < 32; p++ {
			for col := 0; col < 8; col++ {
				r, g, b, a := display[p].RGBA()
				pixels = append(pixels, uint8(r))
				pixels = append(pixels, uint8(g))
				pixels = append(pixels, uint8(b))
				pixels = append(pixels, uint8(a))
			}
		}
	}

	paletteTableImage.WritePixels(pixels)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))

	screen.DrawImage(paletteTableImage, op)

	op.GeoM.Translate(float64(debugPatternId)*32, 0)

	screen.DrawImage(paletteSelectedBoxImage, op)
}

func (e *Emulator) DrawPatternTableAt(screen *ebiten.Image, x, y int) {
	display := e.VM.GetPatternTableDisplay(0, debugPatternId)

	var pixels []byte
	for px := 0; px < 128; px++ {
		for py := 0; py < 128; py++ {
			r, g, b, a := display[px][py].RGBA()
			pixels = append(pixels, uint8(r))
			pixels = append(pixels, uint8(g))
			pixels = append(pixels, uint8(b))
			pixels = append(pixels, uint8(a))
		}
	}

	patternTableImage.WritePixels(pixels)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))

	screen.DrawImage(patternTableImage, op)

	display2 := e.VM.GetPatternTableDisplay(1, debugPatternId)

	var pixels2 []byte
	for px := 0; px < 128; px++ {
		for py := 0; py < 128; py++ {
			r, g, b, a := display2[px][py].RGBA()
			pixels2 = append(pixels2, uint8(r))
			pixels2 = append(pixels2, uint8(g))
			pixels2 = append(pixels2, uint8(b))
			pixels2 = append(pixels2, uint8(a))
		}
	}

	patternTableImage.WritePixels(pixels2)

	op.GeoM.Translate(128, 0)

	screen.DrawImage(patternTableImage, op)
}

func (e *Emulator) DrawScreenAt(screen *ebiten.Image, x, y int) {
	vmScreen := e.VM.GetScreen()

	var pixels []byte
	for px := 0; px < 256; px++ {
		for py := 0; py < 240; py++ {
			r, g, b, a := vmScreen[px][py].RGBA()
			pixels = append(pixels, uint8(r))
			pixels = append(pixels, uint8(g))
			pixels = append(pixels, uint8(b))
			pixels = append(pixels, uint8(a))
		}
	}

	screenImage.WritePixels(pixels)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))

	screen.DrawImage(screenImage, op)
}

func (e *Emulator) DrawCpuAt(screen *ebiten.Image, x, y int) {
	cpu := e.VM.PeekCPU()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("PC: 0x%04X", cpu.PC), x, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" A: 0x%02X", cpu.A), x, y+12)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" X: 0x%02X", cpu.X), x, y+24)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" Y: 0x%02X", cpu.Y), x, y+36)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SP: 0x%04X", cpu.StackPtr), x, y+48)
	ebitenutil.DebugPrintAt(screen, "STATUS: N V U B D I Z C", x, y+60)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("        %d %d %d %d %d %d %d %d", cpu.P>>7&1, cpu.P>>6&1, cpu.P>>5&1, cpu.P>>4&1, cpu.P>>3&1, cpu.P>>2&1, cpu.P>>1&1, cpu.P>>0&1), x, y+72)
}

func (e *Emulator) DrawRamAt(screen *ebiten.Image, startAddr, rows uint16, x, y int) {
	ram1 := e.VM.PeekRAM(startAddr, startAddr+rows*16-1)
	for i := 0; i < int(rows); i++ {
		s := fmt.Sprintf("0x%04X: ", startAddr+uint16(i)*16)
		for j := 0; j < 16; j++ {
			s += fmt.Sprintf("%02X ", ram1[i*16+j])
		}
		ebitenutil.DebugPrintAt(screen, s, x, y+i*12)
	}
}

func (e *Emulator) DrawStateAt(screen *ebiten.Image, x, y int) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Mode: %v", e.Mode), x, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("State: %v", e.State), x, y+12)
}

func (e *Emulator) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth / windowScale, outsideHeight / windowScale
}

func (e *Emulator) Start() {
	ebiten.SetWindowTitle("NES Emulator in Go!")
	ebiten.SetWindowSize(windowWidth, windowHeight)

	e.Disassembly = e.VM.PeekDisassembly()
	disassemblyHighlightImage.Fill(color.RGBA{
		R: 0,
		G: 20,
		B: 100,
		A: 255,
	})

	for px := 0; px < 32; px++ {
		for py := 0; py < 8; py++ {
			if px == 0 || px == 31 || py == 0 || py == 7 {
				paletteSelectedBoxImage.Set(px, py, color.White)
			}
		}
	}

	if err := ebiten.RunGame(e); err != nil {
		log.Fatal(err)
	}
}

func (e *Emulator) StartWithROM(filePath string) {
	e.VM.LoadROM(filePath)
	if e.Mode == Automation {
		e.VM.ForceSetResetVector(0xC000)
	}
	e.VM.Reset()
	e.Start()
}

func (e *Emulator) StartWithProgram(program string, startAddr uint16) {
	e.VM.LoadProgramAsString(program, startAddr)
	e.Start()
}
