package main

// TODO: PPURegisters are mirrored in here and in the PPU class; there can only be one.

const (
	ramLength              = 0x0800
	ppuRegisterLength      = 0x0008
	apuRegisterLength      = 0x0018
	disabledRegisterLength = 0x0008
)

//Memory represents all memory on the snes system.
type Memory struct {
	//Ram represents the RAM in Snes.
	RAM [ramLength]byte
	//PPURegisters for the pixel processing unit.
	PPURegisters [ppuRegisterLength]byte
	//APURegisters for the audio unit.
	APURegisters [apuRegisterLength]byte
	//DisabledRegisters that are by default disabled on the SNES system.
	DisabledRegsiters [disabledRegisterLength]byte
	//Cartridge holds the nes cartridge in memory.
	cartridge *Cartridge
	//The mapper we are using under the hood for ram.
	mapper Mapper
	//Controllers
	controller *[2]Controller
	//PPU for ppu memory access.
	ppu *PPU
	//CPU for cpu memory access.
	cpu *CPU
}

func (system *System) resetMemory() {
	system.memory = Memory{}
	system.memory.ppu = &system.ppu
	system.memory.cpu = &system.cpu
	system.memory.controller = &system.controller
}

//WriteByte Writes a byte to the given address.
func (memory *Memory) WriteByte(address uint16, value byte) {
	switch {
	case address <= 0x1FFF:
		memory.RAM[address%0x0800] = value
	case address < 0x4000:
		memory.ppu.WriteRegister(int(address&0x7), value)
	case address == 0x4014:
		// OAMDMA
		memory.ppu.WriteRegister(0x4014, value)
	case address == 0x4016:
		memory.controller[0].Write(value)
		memory.controller[1].Write(value)
	case address >= 0x4020:
		memory.mapper.WriteByte(address, value)
	}
	// TODO do the APU and I/O
}

//ReadByte Reads a byte from the ram.
func (memory *Memory) ReadByte(address uint16) byte {
	// see https://wiki.nesdev.com/w/index.php/CPU_memory_map
	switch {
	case address <= 0x1FFF:
		return memory.RAM[address&0x07FF]
	case address <= 0x3FFF:
		return memory.ppu.ReadRegister(int(address & 0x7))
	case address == 0x4016:
		return memory.controller[0].Read()
	case address == 0x4017:
		return memory.controller[1].Read()
	case address <= 0x4017:
		// APU & IO Registers
		return 0
	case address <= 0x401F:
		// CPU test mode
		return 0
	case address >= 0x4020:
		return memory.mapper.ReadByte(address)
	}
	return 0
}

//ReadUint16 reads 2 bytes from the given address and returns it in a unsigned 16 byte int.
func (memory *Memory) ReadUint16(address uint16) uint16 {
	return uint16(memory.ReadByte(address)) | (uint16(memory.ReadByte(address+1)) << 8)
}

//WriteUint16 writes 2 bytes to the given address.
func (memory *Memory) WriteUint16(address uint16, value uint16) {
	//TODO: Verify endianness of processor.
	memory.WriteByte(address, byte(value&0xFF00))
	//Move address forward and store low byte.
	memory.WriteByte(address+1, byte(value&0x00FF))
}
