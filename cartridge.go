package main

//TODO: This will need to be removed when targeting WASM.
import (
	"io"
	"os"
)

const (
	headerSize     = 16
	prgRomBankSize = 16384
	chrRomBankSize = 8192
	trainerSize    = 512
)

//Represents a iNES header.
type iNES struct {
	SizeRomPRG          byte
	SizeRomCHR          byte
	VerticalMirroring   bool
	FourScreenMirroring bool
	BatteryBacked       bool
	TrainerExists       bool
	IgnoreMirroring     bool
	MapperNumber        uint8
	SizeRAMPRG          byte
	ExtraFlags          [7]byte
}

//Cartridge struct defines a Cartridge in memory.
type Cartridge struct {
	header     iNES
	prg        []byte
	chr        []byte
	mapperID   uint8
	mirrorMode int
}

func (system *System) resetCartridge(nesFileName string) {
	system.memory.cartridge = &Cartridge{}
	system.LoadFromFile(nesFileName)
}

//LoadFromFile TODO: This will need to be removed when targeting WASM. Maybe #define it out?
//LoadFromFile loads a nes rom from a file.
func (system *System) LoadFromFile(nesFileName string) {
	file, _ := os.Open(nesFileName)
	defer file.Close()
	fileSize, _ := os.Stat(nesFileName)
	fileData := make([]byte, fileSize.Size())
	_, _ = io.ReadFull(file, fileData)
	system.LoadFromString(fileData)
}

//LoadFromString loads a nes rom from a string.
func (system *System) LoadFromString(nesFile []byte) {
	//Verify that the header is valid.
	headerSlice := nesFile[0:headerSize]
	assertCartridge(headerSlice)
	system.memory.cartridge.getNumPrgRomBanks(headerSlice)
	system.memory.cartridge.getNumChrRomBanks(headerSlice)
	system.memory.cartridge.isRomVerticalMirroring(headerSlice)
	system.memory.cartridge.isFourScreenMirroring(headerSlice)
	system.memory.cartridge.doesTrainerExist(headerSlice)
	system.memory.cartridge.getMapperNumber(headerSlice)
	offset := uint32(headerSize)

	//Trainers are not supported.
	if system.memory.cartridge.header.TrainerExists {
		offset += trainerSize
	}

	//Read in the PRG rom.
	system.memory.cartridge.prg = nesFile[offset : uint32(system.memory.cartridge.header.SizeRomPRG)*prgRomBankSize+offset]
	offset += uint32(system.memory.cartridge.header.SizeRomPRG) * prgRomBankSize
	//Read in the CHR rom
	if system.memory.cartridge.header.SizeRomCHR != 0 {
		system.memory.cartridge.chr = nesFile[offset : uint32(system.memory.cartridge.header.SizeRomCHR)*chrRomBankSize+offset]
	} else {
		system.memory.cartridge.chr = make([]byte, chrRomBankSize)
	}

	//Load the mapper for our system.
	system.memory.mapper = system.ResetMapper()
}

func assertCartridge(nesFile []byte) {
	if len(nesFile) < headerSize {
		panic("Not a valid iNES file!")
	}

	if string(nesFile[:3]) != "NES" || nesFile[3] != byte(0x1A) {
		panic("Not a valid iNES file!")
	}
}

func (cartridge *Cartridge) getNumPrgRomBanks(nesFile []byte) {
	cartridge.header.SizeRomPRG = nesFile[4]
}

func (cartridge *Cartridge) getNumChrRomBanks(nesFile []byte) {
	cartridge.header.SizeRomCHR = nesFile[5]
}

func (cartridge *Cartridge) isRomVerticalMirroring(nesFile []byte) {
	cartridge.header.VerticalMirroring = (nesFile[6] & 0x01) != 0
}

func (cartridge *Cartridge) doesTrainerExist(nesFile []byte) {
	cartridge.header.TrainerExists = (nesFile[6] & 0x04) != 0
}

func (cartridge *Cartridge) isFourScreenMirroring(nesFile []byte) {
	cartridge.header.FourScreenMirroring = (nesFile[6] & 0x08) != 0
}

func (cartridge *Cartridge) getMapperNumber(nesFile []byte) {
	cartridge.header.MapperNumber = uint8((0xF0 & nesFile[7]) | (0xF0&nesFile[6])>>4)
}
