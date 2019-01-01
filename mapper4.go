package main

//MapperMMC3 represents mapper id 4.
type MapperMMC3 struct {
	memory *Memory

	irqEnabled bool
	irqLatch   byte
	irqReload  byte

	mirrorMode int // 0: vertical, 1: horizontal

	bankRegisters      [8]int
	bankSelectRegister int
	bankPRGMode        int
	bankCHRMode        int

	prgRAM [8192]byte

	counter byte
}

func (memory *Memory) resetMapperMMC3() *MapperMMC3 {
	return &MapperMMC3{
		memory:     memory,
		irqEnabled: true,
		irqReload:  0,
	}
}

//ReadByte acts like mapper mmc 3 reading a byte.
func (mapper *MapperMMC3) ReadByte(addr uint16) byte {
	switch {
	case addr <= 0x1FFF:
		return mapper.memory.cartridge.chr[mapper.resolvePpuRomAddr(addr)]
	case addr <= 0x2FFF:
		return mapper.memory.ppu.vram[TranslateVRamAddress(addr, 1-mapper.mirrorMode)]
	case addr < 0x6000:
		// ?????
	case addr <= 0x7FFF:
		// internal ram
		return mapper.prgRAM[addr-0x6000]
	case addr <= 0xFFFF:
		return mapper.memory.cartridge.prg[mapper.resolveCPURomAddr(addr)]
	}
	return 0
}

//WriteByte acts like mapper mmc 3 writing a byte.
func (mapper *MapperMMC3) WriteByte(addr uint16, data byte) {
	switch {
	case addr <= 0x1FFF:
		mapper.memory.cartridge.chr[mapper.resolvePpuRomAddr(addr)] = data
	case addr <= 0x2FFF:
		mapper.memory.ppu.vram[TranslateVRamAddress(addr, 1-mapper.mirrorMode)] = data
	case addr < 0x6000:
		// ?????
	case addr <= 0x7FFF:
		// write to prg ram
		mapper.prgRAM[addr-0x6000] = data
	case addr <= 0x9FFF && (addr&0x1 == 0):
		// bank select register
		mapper.bankSelectRegister = int(data & 0x7)
		mapper.bankPRGMode = int(data&0x40) >> 6
		mapper.bankCHRMode = int(data&0x80) >> 7
	case addr <= 0x9FFF && (addr&0x1 == 1):
		// TODO bank data register
		if mapper.bankSelectRegister == 6 || mapper.bankSelectRegister == 7 {
			data &= 0x3F
		} else if mapper.bankSelectRegister == 0 || mapper.bankSelectRegister == 1 {
			data &= 0xFE
		}
		mapper.bankRegisters[mapper.bankSelectRegister] = int(data)
	case addr <= 0xBFFF && (addr&0x1 == 0):
		// mirroring register
		mapper.mirrorMode = int(data & 0x1)
	case addr <= 0xBFFF && (addr&0x1 == 1):
		// TODO PRG RAM protect register
	case addr <= 0xDFFF && (addr&0x1 == 0):
		// IRQ latch register
		mapper.irqLatch = data
	case addr <= 0xDFFF && (addr&0x1 == 1):
		// IRQ reload register
		mapper.irqReload = data
	case addr <= 0xFFFF && (addr&0x1 == 0):
		// IRQ disable register
		// TODO acknowledge pending interrupts (??)
		mapper.irqEnabled = false
	case addr <= 0xFFFF && (addr&0x1 == 1):
		// IRQ enable register
		mapper.irqEnabled = true
	}
}

func (mapper *MapperMMC3) resolvePpuRomAddr(addr uint16) int {
	bankAddr := addr & 0x3FF
	bankIndex := int(addr&0x1C00) >> 10

	if mapper.bankCHRMode != 0 {
		if bankIndex >= 4 {
			bankIndex -= 4
		} else {
			bankIndex += 4
		}
	}

	var bank int
	if bankIndex < 4 {
		bank = mapper.bankRegisters[bankIndex/2] | (bankIndex & 0x1)
	} else {
		bank = mapper.bankRegisters[bankIndex-2]
	}

	return bank*1024 + int(bankAddr)
}

func (mapper *MapperMMC3) resolveCPURomAddr(addr uint16) int {
	// maps a raw address for the CPU into the ROM (0x8000 to 0xFFFF)
	if mapper.bankPRGMode == 0 {
		switch {
		case addr <= 0x9FFF:
			return (8192 * mapper.bankRegisters[6]) + int(addr-0x8000)
		case addr <= 0xBFFF:
			return (8192 * mapper.bankRegisters[7]) + int(addr-0xA000)
		case addr <= 0xDFFF:
			return (8192*-2 + len(system.memory.cartridge.prg)) + int(addr-0xC000)
		case addr <= 0xFFFF:
			return (8192*-1 + len(system.memory.cartridge.prg)) + int(addr-0xE000)
		default:
			panic("PRG0 Out of bounds!")
		}
	} else {
		switch {
		case addr <= 0x9FFF:
			return (8192*-2 + len(system.memory.cartridge.prg)) + int(addr-0x8000)
		case addr <= 0xBFFF:
			return (8192 * mapper.bankRegisters[7]) + int(addr-0xA000)
		case addr <= 0xDFFF:
			return (8192 * mapper.bankRegisters[6]) + int(addr-0xC000)
		case addr <= 0xFFFF:
			return (8192*-1 + len(system.memory.cartridge.prg)) + int(addr-0xE000)
		default:
			panic("PRG1 Out of bounds!")
		}
	}
}

func (mapper *MapperMMC3) Emulate() {
	if mapper.memory.ppu.scanlineCount != 280 { // TODO: this *should* be 260
		return
	}
	if mapper.memory.ppu.scanlineCount > 239 && mapper.memory.ppu.scanlineCount < 261 {
		return
	}
	if mapper.memory.ppu.showBackgroundLeft == 0 && mapper.memory.ppu.showSpritesLeft == 0 {
		return
	}
	mapper.handleScanLine()
}

func (mapper *MapperMMC3) handleScanLine() {
	if mapper.counter == 0 {
		mapper.irqReload = mapper.irqLatch
	} else {
		mapper.irqReload--
		if mapper.irqReload == 0 && mapper.irqEnabled {
			mapper.memory.cpu.triggerInterruptIRQ()
		}
	}
}
