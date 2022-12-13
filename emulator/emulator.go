package emulator

import (
	"fmt"
	"go-nes/nes"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	windowWidth  = 1024
	windowHeight = 960
	windowScale  = 3
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
	switch e.State {
	case Init:
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			e.State = Stepping
		}
	case Running:
		e.VM.Step()

		if ebiten.IsKeyPressed(ebiten.KeyP) {
			e.State = Paused
		}
		if ebiten.IsKeyPressed(ebiten.KeyF1) {
			e.State = Stepping
		}
	case Stepping:
		if !e.IsKeyPressed && ebiten.IsKeyPressed(ebiten.KeySpace) {
			e.IsKeyPressed = true
			e.VM.Step()
		}
		if !ebiten.IsKeyPressed(ebiten.KeySpace) {
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
	e.DrawCpuAt(screen, 0, 0)
	e.DrawRamAt(screen, 0x0000, 10, 0, 108)
	e.DrawStateAt(screen, 0, 252)
}

func (e *Emulator) DrawCpuAt(screen *ebiten.Image, x, y int) {
	cpu := e.VM.PeekCPU()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("PC: 0x%04X", cpu.PC), x, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" A: 0x%02X", cpu.A), x, y+12)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" X: 0x%02X", cpu.X), x, y+24)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" Y: 0x%02X", cpu.Y), x, y+36)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SP: 0x%04X", cpu.StackPtr), x, y+48)
	ebitenutil.DebugPrintAt(screen, "STATUS: N V U B D I Z C", x, y+60)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("        %d %d %d %d %d %d %d %d", cpu.P>>7|1, cpu.P>>7|1, cpu.P>>7|1, cpu.P>>7|1, cpu.P>>7|1, cpu.P>>7|1, cpu.P>>7|1, cpu.P>>7|1), x, y+72)

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
	if err := ebiten.RunGame(e); err != nil {
		log.Fatal(err)
	}
}

func (e *Emulator) StartWithROM(filePath string) {
	e.VM.LoadROM(filePath)
	if e.Mode == Automation {
		e.VM.ForceSetResetVector(0xC000)
	}
	e.Start()
}

func (e *Emulator) StartWithProgram(program string, startAddr uint16) {
	e.VM.LoadProgramAsString(program, startAddr)
	e.Start()
}
