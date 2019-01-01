package main

//Mapper3 stores the state of the mapper 3 struct.
type Mapper3 struct {
	system   *System
	bank     int
	numBanks int
}

func (memory *Memory) resetMapper3() *Mapper3 {
	numBanks := len(system.memory.cartridge.chr) / 8192
	return &Mapper3{
		system:   system,
		numBanks: numBanks,
		bank:     numBanks - 1,
	}
}

//ReadByte reads a byte according to mapper 3.
func (mapper *Mapper3) ReadByte(addr uint16) byte {
	switch {
	case addr <= 0x1FFF:
		return mapper.system.memory.cartridge.chr[uint16(mapper.bank*8192)+addr]
	case addr <= 0x2FFF:
		return mapper.system.ppu.vram[TranslateVRamAddress(addr, mapper.system.memory.cartridge.mirrorMode)]
	case addr >= 0x8000 && addr <= 0xBFFF:
		return mapper.system.memory.cartridge.prg[addr-0x8000]
	case addr >= 0xC000 && addr <= 0xFFFF:
		if len(mapper.system.memory.cartridge.prg) > 0x4000 {
			return mapper.system.memory.cartridge.prg[addr-0x8000]
		}
		return mapper.system.memory.cartridge.prg[addr-0xC000]
	default:
		panic("MMC3 Read out of bounds!")
	}
}

//WriteByte writes a byte according to mapper 3.
func (mapper *Mapper3) WriteByte(addr uint16, data byte) {
	if addr >= 0x8000 && addr <= 0xFFFF {
		mapper.bank = int(data) % mapper.numBanks
	}
}

func (mapper *Mapper3) Emulate() {
	return
}
