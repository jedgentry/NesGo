package main

import "testing"

func test(t *testing.T) {
	startWithRom("test-roms/nestest.nes", 0x0C00)
	print("Finished CPU test!")
}
