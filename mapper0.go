package main

//Mapper0 represents the snes Mapper0, simple and direct.
type Mapper0 struct {
	memory *Memory
}

//ResetMapper0 resets the mapper to the current memory.
func (memory *Memory) resetMapper0() *Mapper0 {
	return &Mapper0{
		memory: memory,
	}
}

//ReadByte reads a byte from the given memory location.
func (mapper *Mapper0) ReadByte(addr uint16) byte {
	switch {
	case addr <= 0x1FFF:
		return mapper.memory.cartridge.chr[addr]
	case addr <= 0x2FFF:
		return mapper.memory.ppu.vram[TranslateVRamAddress(addr, mapper.memory.cartridge.mirrorMode)]
	case addr >= 0x8000 && addr <= 0xBFFF:
		return mapper.memory.cartridge.prg[addr-0x8000]
	case addr >= 0xC000 && addr <= 0xFFFF:
		if len(mapper.memory.cartridge.prg) > 0x4000 {
			return mapper.memory.cartridge.prg[addr-0x8000]
		}
		return mapper.memory.cartridge.prg[int(addr)-0xC000]
	default:
		return 0
	}
}

func (mapper *Mapper0) Emulate() {
	return
}

//WriteByte writes a byte to the given memory location.
func (mapper *Mapper0) WriteByte(addr uint16, data byte) {
	switch {
	case addr <= 0x1FFF:
		if int(addr) > len(mapper.memory.cartridge.chr) {
			print("Oh no")
		}
		mapper.memory.cartridge.chr[addr] = data
	case addr <= 0x2FFF:
		mapper.memory.ppu.vram[TranslateVRamAddress(addr, mapper.memory.cartridge.mirrorMode)] = data
	}
}
