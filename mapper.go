package main

import "strconv"

//TODO: this is also defined in cartridge, maybe reuse?
const (
	MirrorHorizontal = iota
	MirrorVertical
	MirrorSingleA
	MirrorSingleB
	MirrorFour
)

//MirrorLookup TODO: Figure out why this is needed.
var MirrorLookup = [][4]int{
	{0, 0, 1, 1},
	{0, 1, 0, 1},
	{0, 0, 0, 0},
	{1, 1, 1, 1},
	{0, 1, 2, 3},
}

//Mapper interface defines the functionality of a memory mapper.
type Mapper interface {
	ReadByte(address uint16) byte
	WriteByte(address uint16, value byte)
	Emulate()
}

//ResetMapper gets the current mapper representing the cartridge.
func (system *System) ResetMapper() Mapper {
	print("Using mapper: " + strconv.Itoa(int(system.memory.cartridge.header.MapperNumber)))
	switch system.memory.cartridge.header.MapperNumber {
	case 0:
		return system.memory.resetMapper0()
	case 1:
		return system.memory.resetMapperMMC1()
	case 3:
		return system.memory.resetMapper3()
	case 4:
		return system.memory.resetMapperMMC3()
	default:
		print("Bad mapper: " + strconv.Itoa(int(system.memory.cartridge.header.MapperNumber)))
		panic("bad mapper!")
	}
}

//TranslateVRamAddress takes a physical address to a VRam one.
func TranslateVRamAddress(address uint16, mirrorMode int) int {
	address -= 0x2000
	bank := MirrorLookup[mirrorMode][address/(0x400)]
	return (bank * 0x400) + int(address%0x400)
}
