package main

//System represents the entire NES system.
type System struct {
	memory     Memory
	cpu        CPU
	ppu        PPU
	apu        APU
	controller [2]Controller
}

//ResetSystem resets the system struct. This is equivalent to pressing reset.
func (system *System) ResetSystem(nesFilename string) {
	system.resetCPU()
	system.resetAPU()
	system.resetMemory()
	system.cpu.memory = &system.memory
	system.resetPPU()
	system.resetControllers()
	system.resetCartridge(nesFilename)
}

//NewSystem returns a new system.
func NewSystem() *System {
	return &System{}
}

//Emulate starts the system clock.
func (system *System) Emulate() int {
	cpuCycles := system.cpu.Emulate(1)
	ppuClocks := cpuCycles * 3
	for i := 0; i < ppuClocks; i++ {
		system.ppu.Emulate(1)
		system.memory.mapper.Emulate()
	}

	for i := 0; i < cpuCycles; i++ {
		system.apu.Step()
	}
	return cpuCycles
}

//EmulateFrame emulates one frame of the ssytem.
func (system *System) EmulateFrame() int {
	cycles := 0
	startFrame := system.ppu.frameCount
	for startFrame == system.ppu.frameCount {
		cycles += system.Emulate()
	}
	return cycles
}
