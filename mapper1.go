package main

//MapperMMC1 presents the MMC1 Mapper.
type MapperMMC1 struct {
	memory *Memory

	shiftRegister byte
	shiftNumber   int

	mirrorMode int

	registerControl byte
	registerCHR0    byte
	registerCHR1    byte
	registerPRG     byte

	prgRAM [8192]byte
}

func (memory *Memory) resetMapperMMC1() *MapperMMC1 {
	return &MapperMMC1{
		memory:          memory,
		registerControl: 0x0f,
	}
}

//ReadByte reads a byte according to the MMC1 mapper.
func (mapper *MapperMMC1) ReadByte(addr uint16) byte {
	switch {
	case addr <= 0x0FFF:
		// CHR bank 1
		return mapper.memory.cartridge.chr[mapper.getCHR1Index(addr)]
	case addr <= 0x1FFF:
		// CHR bank 2
		return mapper.memory.cartridge.chr[mapper.getCHR2Index(addr)]
	case addr <= 0x2FFF:
		// mirroring
		return mapper.memory.ppu.vram[TranslateVRamAddress(addr, mapper.mirrorMode)]
	case addr < 0x6000:
		panic("Unknown mapper address.")
	case addr >= 0x6000 && addr <= 0x7FFF:
		// internal ram
		return mapper.prgRAM[addr-0x6000]
	case addr <= 0xBFFF:
		// PRG bank 1
		switch (mapper.registerControl & 0xC) >> 2 {
		case 0:
			fallthrough
		case 1:
			// switch 32 KB at $8000, ignoring low bit of bank number
			return mapper.memory.cartridge.prg[16384*int(mapper.registerPRG&0xFE)+int(addr-0x8000)]
		case 2:
			// fix first bank at $8000
			return mapper.memory.cartridge.prg[int(addr-0x8000)]
		case 3:
			// switch 16 KB bank at $8000
			return mapper.memory.cartridge.prg[16384*int(mapper.registerPRG)+int(addr-0x8000)]
		}
	case addr <= 0xFFFF:
		// PRG bank 2
		switch (mapper.registerControl & 0xC) >> 2 {
		case 0:
			fallthrough
		case 1:
			return mapper.memory.cartridge.prg[16384*int(mapper.registerPRG|0x1)+int(addr-0xC000)]
		case 2:
			// switch 16 KB bank at $C000
			return mapper.memory.cartridge.prg[16384*int(mapper.registerPRG)+int(addr-0xC000)]
		case 3:
			// fix last bank at $C000
			return mapper.memory.cartridge.prg[len(mapper.memory.cartridge.prg)-16384+int(addr-0xC000)]
		}
	}
	return 0
}

func (mapper *MapperMMC1) getCHR1Index(addr uint16) int {
	bank := int(mapper.registerCHR0)
	if mapper.registerControl&0x10 == 0 {
		// 8KB mode
		bank &= 0xFE
	}
	bank %= len(mapper.memory.cartridge.chr) / 4096 // XXX is this correct behavior?
	return int(addr-0x0000) + (bank * 4096)
}

func (mapper *MapperMMC1) getCHR2Index(addr uint16) int {
	var bank int
	if mapper.registerControl&0x10 == 0 {
		// 8 KB mode
		bank = int(mapper.registerCHR0) | 0x1
	} else {
		bank = int(mapper.registerCHR1)
	}
	bank %= len(mapper.memory.cartridge.chr) / 4096 // XXX is this correct behavior?
	return int(addr-0x1000) + (bank * 4096)
}

//WriteByte writes a byte according to the mmc1 mapper.
func (mapper *MapperMMC1) WriteByte(addr uint16, data byte) {
	if addr < 0x6000 {
		if addr <= 0x0FFF {
			mapper.memory.cartridge.chr[mapper.getCHR1Index(addr)] = data
		} else if addr <= 0x1FFF {
			mapper.memory.cartridge.chr[mapper.getCHR2Index(addr)] = data
		} else if addr <= 0x2FFF {
			mapper.memory.ppu.vram[TranslateVRamAddress(addr, mapper.mirrorMode)] = data
		}
	} else if addr <= 0x7FFF {
		if mapper.registerPRG&0x10 == 0 {
			mapper.prgRAM[addr-0x6000] = data
		}
	} else {
		if data&0x80 > 0 {
			// clear shift register
			mapper.shiftNumber = 0
			mapper.shiftRegister = 0
		} else {
			// add to shift register
			mapper.shiftRegister = mapper.shiftRegister | ((data & 0x1) << uint(mapper.shiftNumber))
			mapper.shiftNumber++
		}

		if mapper.shiftNumber == 5 {
			switch (addr >> 13) & 0x3 {
			case 0:
				mapper.registerControl = mapper.shiftRegister

				switch mapper.registerControl & 0x3 {
				case 0:
					mapper.mirrorMode = MirrorSingleA
				case 1:
					mapper.mirrorMode = MirrorSingleB
				case 2:
					mapper.mirrorMode = MirrorVertical
				case 3:
					mapper.mirrorMode = MirrorHorizontal
				}
			case 1:
				mapper.registerCHR0 = mapper.shiftRegister
			case 2:
				mapper.registerCHR1 = mapper.shiftRegister
			case 3:
				mapper.registerPRG = mapper.shiftRegister
			}
			mapper.shiftNumber = 0
			mapper.shiftRegister = 0
		}
	}
}

func (mapper *MapperMMC1) Emulate() {
}
