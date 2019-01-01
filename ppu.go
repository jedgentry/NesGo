package main

//PPU Represents the state of the pixel processing unit
type PPU struct {
	// drawing interfaces
	funcPushPixel func(int, int, uint32)
	funcPushFrame func()

	vram          [2048]byte
	oam           [256]byte
	secondaryOam  [32]byte
	palette       [32]byte
	colors        [64]uint32
	warmupTicker  int
	scanlineCount int
	tickCount     int
	frameCount    int
	cycles        uint64

	statusRendering      bool
	vBlank               byte
	sprite0Hit           byte
	spriteOverflow       byte
	ppuDataBuffer        byte
	ppuLatch             byte
	v                    uint16
	t                    uint16
	x                    byte
	w                    byte
	backgroundBitmapData uint64

	// sprite rendering
	spriteEvaluationN         int
	spriteEvaluationM         int
	spriteEvaluationRead      byte
	pendingNumScanlineSprites int
	numScanlineSprites        int
	spriteXPositions          [8]int
	spriteAttributes          [8]byte
	spriteBitmapDataLo        [8]byte
	spriteBitmapDataHi        [8]byte
	spriteZeroAt              int
	spriteZeroAtNext          int

	baseNametable                 byte
	incrementVram                 byte
	spriteTableAddress            byte
	backgroundTableAddress        byte
	spriteSize                    byte
	masterSlave                   byte
	generateNonMaskableInterrupts byte

	grayscale          byte
	showSpritesLeft    byte
	showBackgroundLeft byte
	renderSprites      byte
	renderBackground   byte
	emphasizeRed       byte
	emphasizeGreen     byte
	emphasizeBlue      byte

	oamAddr byte

	ram *Memory
	cpu *CPU
}

func (system *System) resetPPU() {
	system.ppu.resetVRAM()
	system.ppu.resetOAM()
	system.ppu.resetSecondaryOAM()
	system.ppu.resetPallete()
	system.ppu.resetColor()
	system.ppu.resetState()
	system.ppu.resetSpriteRenderingState()
	system.ppu.resetPPUControl()
	system.ppu.resetPPUMask()
	system.ppu.ram = &system.memory
	system.ppu.cpu = &system.cpu
}

func (ppu *PPU) resetVRAM() {
	for i := 0; i < 2048; i++ {
		ppu.vram[i] = 0
	}
}

func (ppu *PPU) resetOAM() {
	for i := 0; i < 256; i++ {
		ppu.oam[i] = 0
	}
}

func (ppu *PPU) resetSecondaryOAM() {
	for i := 0; i < 32; i++ {
		ppu.secondaryOam[i] = 0
	}
}

func (ppu *PPU) resetPallete() {
	for i := 0; i < 32; i++ {
		ppu.palette[i] = 0
	}
}

func (ppu *PPU) resetColor() {
	ppu.colors = [64]uint32{84*256*256 + 84*256 + 84, 0*256*256 + 30*256 + 116, 8*256*256 + 16*256 + 144, 48*256*256 + 0*256 + 136, 68*256*256 + 0*256 + 100, 92*256*256 + 0*256 + 48, 84*256*256 + 4*256 + 0, 60*256*256 + 24*256 + 0, 32*256*256 + 42*256 + 0, 8*256*256 + 58*256 + 0, 0*256*256 + 64*256 + 0, 0*256*256 + 60*256 + 0, 0*256*256 + 50*256 + 60, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 152*256*256 + 150*256 + 152, 8*256*256 + 76*256 + 196, 48*256*256 + 50*256 + 236, 92*256*256 + 30*256 + 228, 136*256*256 + 20*256 + 176, 160*256*256 + 20*256 + 100, 152*256*256 + 34*256 + 32, 120*256*256 + 60*256 + 0, 84*256*256 + 90*256 + 0, 40*256*256 + 114*256 + 0, 8*256*256 + 124*256 + 0, 0*256*256 + 118*256 + 40, 0*256*256 + 102*256 + 120, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 236*256*256 + 238*256 + 236, 76*256*256 + 154*256 + 236, 120*256*256 + 124*256 + 236, 176*256*256 + 98*256 + 236, 228*256*256 + 84*256 + 236, 236*256*256 + 88*256 + 180, 236*256*256 + 106*256 + 100, 212*256*256 + 136*256 + 32, 160*256*256 + 170*256 + 0, 116*256*256 + 196*256 + 0, 76*256*256 + 208*256 + 32, 56*256*256 + 204*256 + 108, 56*256*256 + 180*256 + 204, 60*256*256 + 60*256 + 60, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 236*256*256 + 238*256 + 236, 168*256*256 + 204*256 + 236, 188*256*256 + 188*256 + 236, 212*256*256 + 178*256 + 236, 236*256*256 + 174*256 + 236, 236*256*256 + 174*256 + 212, 236*256*256 + 180*256 + 176, 228*256*256 + 196*256 + 144, 204*256*256 + 210*256 + 120, 180*256*256 + 222*256 + 120, 168*256*256 + 226*256 + 144, 152*256*256 + 226*256 + 180, 160*256*256 + 214*256 + 228, 160*256*256 + 162*256 + 160, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0}
}

func (ppu *PPU) resetState() {
	ppu.warmupTicker = 0
	ppu.scanlineCount = 0
	ppu.tickCount = 0
	ppu.frameCount = 0
	ppu.cycles = 0
	ppu.statusRendering = false
	ppu.vBlank = 0
	ppu.sprite0Hit = 0
	ppu.spriteOverflow = 0
	ppu.ppuDataBuffer = 0
	ppu.ppuLatch = 0
	ppu.v = 0
	ppu.t = 0
	ppu.x = 0
	ppu.w = 0
	ppu.backgroundBitmapData = 0
}

//ReadPPU Reads a byte from the PPU.
func (memory *Memory) ReadPPU(addr uint16) byte {
	// https://wiki.nesdev.com/w/index.php/PPU_memory_map
	addr = addr & 0x3FFF
	switch {
	case addr <= 0x2FFF:
		return memory.mapper.ReadByte(addr)
	case addr <= 0x3EFF:
		// mirrored from 0x2000
		return memory.mapper.ReadByte(addr - 0x1000)
	case addr <= 0x3FFF:
		// (only bottom 0x1F -- 5 bits)
		index := addr & 0x1F
		return memory.ppu.palette[index]
	}
	panic("Bad read on ppu!")
}

//WritePPU writes a byte to the PPU.
func (memory *Memory) WritePPU(addr uint16, data byte) {
	addr = addr & 0x3FFF
	switch {
	case addr <= 0x2FFF:
		memory.mapper.WriteByte(addr, data)
	case addr <= 0x3EFF:
		// mirrored from 0x2000
		memory.mapper.WriteByte(addr-0x1000, data)
	case addr <= 0x3FFF:
		index := addr & 0x1F
		if index == 0x10 || index == 0x14 || index == 0x18 || index == 0x1C {
			index -= 0x10
		}
		memory.ppu.palette[index] = data
	}
}

func (ppu *PPU) resetSpriteRenderingState() {
	ppu.spriteEvaluationN = 0
	ppu.spriteEvaluationM = 0
	ppu.spriteEvaluationRead = 0
	ppu.pendingNumScanlineSprites = 0
	ppu.numScanlineSprites = 0

	for i := 0; i < 8; i++ {
		ppu.spriteXPositions[i] = 0
		ppu.spriteAttributes[i] = 0
		ppu.spriteBitmapDataLo[i] = 0
		ppu.spriteBitmapDataHi[i] = 0
	}

	ppu.spriteZeroAt = 0
	ppu.spriteZeroAtNext = 0
}

func (ppu *PPU) resetPPUControl() {
	ppu.baseNametable = 0
	ppu.incrementVram = 0
	ppu.spriteTableAddress = 0
	ppu.backgroundTableAddress = 0
	ppu.spriteSize = 0
	ppu.masterSlave = 0
	ppu.generateNonMaskableInterrupts = 0
}

func (ppu *PPU) resetPPUMask() {

}

//ReadRegister reads a register from the PPU and returns the value.
func (ppu *PPU) ReadRegister(register int) byte {
	switch register {
	case 2:
		// PPUSTATUS
		status := byte(ppu.ppuLatch & 0x1F)
		status |= ppu.spriteOverflow << 5
		status |= ppu.sprite0Hit << 6
		status |= ppu.vBlank << 7

		ppu.vBlank = 0
		ppu.ppuLatch = status
		ppu.w = 0
		return status
	case 4:
		// OAMDATA
		// TODO if visible scanline and cycle between 1-64, return 0xFF
		return ppu.oam[ppu.oamAddr]
		// XXX increment after read during rendering?
	case 7:
		// PPUDATA
		var data byte
		if ppu.v <= 0x3EFF {
			// buffer this read
			data = ppu.ram.ReadPPU(uint16(ppu.v))
			ppu.ppuDataBuffer, data = data, ppu.ppuDataBuffer
		} else {
			ppu.ppuDataBuffer = ppu.ram.ReadPPU(uint16(ppu.v - 0x1000))
		}

		if ppu.incrementVram == 0 {
			ppu.v++
		} else {
			ppu.v += 32
		}
		return data
	default:
		return ppu.ppuLatch
	}
}

//WriteRegister writes to a PPU register.
func (ppu *PPU) WriteRegister(register int, data byte) {
	ppu.ppuLatch = data
	switch register {
	case 0:
		// PPUCTRL
		if ppu.cycles > 29658*3 {
			ppu.baseNametable = data & 0x3
			ppu.incrementVram = data & 0x4 >> 2
			ppu.spriteTableAddress = data & 0x8 >> 3
			ppu.backgroundTableAddress = data & 0x10 >> 4
			ppu.spriteSize = data & 0x20 >> 5
			ppu.masterSlave = data & 0x40 >> 6
			ppu.generateNonMaskableInterrupts = data & 0x80 >> 7
			ppu.t = (ppu.t & 0xF3FF) | ((uint16(data) & 0x03) << 10)
		}
	case 1:
		// PPUMASK
		ppu.grayscale = data & 0x1 >> 0
		ppu.showBackgroundLeft = data & 0x2 >> 1
		ppu.showSpritesLeft = data & 0x4 >> 2
		ppu.renderBackground = data & 0x8 >> 3
		ppu.renderSprites = data & 0x10 >> 4
		ppu.emphasizeRed = data & 0x20 >> 5
		ppu.emphasizeGreen = data & 0x40 >> 6
		ppu.emphasizeBlue = data & 0x80 >> 7
	case 3:
		// OAMADDR
		ppu.oamAddr = data
	case 4:
		// OAMDATA
		if !ppu.statusRendering {
			ppu.oam[ppu.oamAddr] = data
			ppu.oamAddr++
		}
	case 5:
		// PPUSCROLL
		if ppu.w == 0 {
			ppu.t = (ppu.t & 0xFFE0) | (uint16(data) >> 3)
			ppu.x = data & 0x7
			ppu.w = 1
		} else {
			ppu.t = (ppu.t & 0x8C1F) | ((uint16(data) & 0xF8) << 2) | ((uint16(data) & 0x7) << 12)
			ppu.w = 0
		}
	case 6:
		// PPUADDR
		if ppu.w == 0 {
			ppu.t = (ppu.t & 0x80FF) | ((uint16(data) & 0x3F) << 8)
			ppu.w = 1
		} else {
			ppu.t = (ppu.t & 0xFF00) | uint16(data)
			ppu.v = ppu.t
			ppu.w = 0
		}
	case 7:
		// PPUDATA
		ppu.ram.WritePPU(uint16(ppu.v), data)
		if ppu.incrementVram == 0 {
			ppu.v++
		} else {
			ppu.v += 32
		}
	case 0x4014:
		// OAMDMA
		ppu.cpu.suspended = 513
		//TODO: This can be sped up.
		if ppu.cpu.totalCycles%2 == 1 {
			ppu.cpu.suspended++
		}

		addr := uint16(data) << 8
		for i := 0; i < 256; i++ {
			addr2 := addr + uint16(i)
			data := ppu.cpu.memory.ReadByte(addr2)
			ppu.oam[(ppu.oamAddr+byte(i))&0xFF] = data
		}
	default:
		panic("Bad ppu register")
	}
}

// https://wiki.nesdev.com/w/index.php/PPU_sprite_evaluation
func (ppu *PPU) handleSpriteEvaluation() {
	if ppu.tickCount >= 1 && ppu.tickCount <= 64 {
		if ppu.tickCount%2 == 0 {
			ppu.secondaryOam[(ppu.tickCount-1)/2] = 0xFF
		}
	}
	if ppu.tickCount == 65 {
		ppu.spriteEvaluationN = 0
		ppu.spriteEvaluationM = 0
		ppu.pendingNumScanlineSprites = 0
		ppu.spriteZeroAtNext = 0
	}
	if ppu.tickCount >= 65 && ppu.tickCount <= 256 {
		// Sprite Evaluation Stage 2: Loading the Secondary OAM
		spriteHeight := byte(8)
		if ppu.spriteSize != 0 {
			spriteHeight = 16
		}

		if ppu.spriteEvaluationN < 64 && ppu.pendingNumScanlineSprites < 8 {
			if ppu.tickCount%2 == 1 {
				// read from primary
				ppu.spriteEvaluationRead = ppu.oam[4*ppu.spriteEvaluationN+ppu.spriteEvaluationM]
			} else {
				// write to secondary
				ppu.secondaryOam[4*ppu.pendingNumScanlineSprites+ppu.spriteEvaluationM] = ppu.spriteEvaluationRead
				if ppu.spriteEvaluationM == 0 {
					// check to see if it's in range
					if byte(ppu.scanlineCount) >= ppu.spriteEvaluationRead && byte(ppu.scanlineCount) < ppu.spriteEvaluationRead+spriteHeight {
						// it's in range!
					} else {
						// not in range.
						ppu.spriteEvaluationM--
						ppu.spriteEvaluationN++
					}
				}
				if ppu.spriteEvaluationM == 3 {
					if ppu.spriteEvaluationN == 0 {
						ppu.spriteZeroAt = ppu.pendingNumScanlineSprites
					}
					ppu.spriteEvaluationN++
					ppu.spriteEvaluationM = 0
					ppu.pendingNumScanlineSprites++
				} else {
					ppu.spriteEvaluationM++
				}
			}
		}
	}
	if ppu.tickCount >= 257 && ppu.tickCount <= 320 {
		ppu.spriteEvaluationN = (ppu.tickCount - 257) / 8
		ppu.numScanlineSprites = ppu.pendingNumScanlineSprites
		ppu.spriteZeroAt = ppu.spriteZeroAtNext
		if (ppu.tickCount-257)%8 == 0 {
			// fetch x position, attribute into temporary latches and counters
			var ypos, tile, attribute, xpos byte
			if ppu.spriteEvaluationN < ppu.numScanlineSprites {
				ypos = ppu.secondaryOam[ppu.spriteEvaluationN*4+0]
				tile = ppu.secondaryOam[ppu.spriteEvaluationN*4+1]
				attribute = ppu.secondaryOam[ppu.spriteEvaluationN*4+2]
				xpos = ppu.secondaryOam[ppu.spriteEvaluationN*4+3]
			} else {
				ypos, tile, attribute, xpos = 0xFF, 0xFF, 0xFF, 0xFF
			}
			ppu.spriteXPositions[ppu.spriteEvaluationN], ppu.spriteAttributes[ppu.spriteEvaluationN] = int(xpos), attribute

			spriteTable := ppu.spriteTableAddress
			tileRow := ppu.scanlineCount - int(ypos)

			if ppu.spriteSize != 0 {
				// 8x16 sprites
				spriteTable = tile & 0x1
				tile = tile & 0xFE
				if tileRow >= 8 {
					tile |= 1 - (attribute & 0x80 >> 7)
					tileRow += 8
				} else {
					tile |= attribute & 0x80 >> 7
				}
			}

			// fetch bitmap data into shift registers
			if attribute&0x80 > 0 {
				// flip sprite vertically
				tileRow = 7 - tileRow
			}
			patternAddr := uint16(0)
			patternAddr |= uint16(tileRow)
			patternAddr |= uint16(tile) << 4
			patternAddr |= uint16(spriteTable) << 12
			lo, hi := ppu.ram.ReadPPU(patternAddr), ppu.ram.ReadPPU(patternAddr+8)

			if attribute&0x40 > 0 {
				// flip sprite horizontally
				var hi2, lo2 byte
				for i := 0; i < 8; i++ {
					hi2 = (hi2 << 1) | (hi & 1)
					lo2 = (lo2 << 1) | (lo & 1)
					hi >>= 1
					lo >>= 1
				}
				lo, hi = lo2, hi2
			}

			ppu.spriteBitmapDataLo[ppu.spriteEvaluationN] = lo
			ppu.spriteBitmapDataHi[ppu.spriteEvaluationN] = hi
		}
	}
}

func (ppu *PPU) updateScrolling() {
	if ppu.tickCount == 256 {
		ppu.incrementScrollY()
	}
	if ppu.tickCount == 257 {
		// copy horizontal bits from t to v
		ppu.v = (ppu.v & 0xFBE0) | (ppu.t & 0x41F)

		if ppu.scanlineCount < 240 {
			ppu.handleSpriteEvaluation()
		}
	}
	if ((ppu.tickCount >= 321 && ppu.tickCount <= 336) || (ppu.tickCount >= 1 && ppu.tickCount <= 256)) && (ppu.tickCount%8 == 0) {
		ppu.incrementScrollX()
	}
}

func (ppu *PPU) renderVisibleScanlines() {
	visibleLine := ppu.scanlineCount < 240 && ppu.scanlineCount >= 0
	visibleCycle := ppu.tickCount >= 1 && ppu.tickCount <= 256
	if visibleLine {
		ppu.handleSpriteEvaluation()
		if visibleCycle {
			ppu.renderPixel()
		}

		// fetching tile data
		if visibleCycle || (ppu.tickCount >= 321 && ppu.tickCount <= 336) {
			ppu.backgroundBitmapData <<= 4
			if ppu.tickCount%8 == 0 {
				ppu.fetchTileData()
			}
		}

		ppu.updateScrolling()
	}
}

func (ppu *PPU) maybePerformVBlank() {
	if ppu.scanlineCount == 241 && ppu.tickCount == 1 {
		ppu.funcPushFrame()
		if ppu.generateNonMaskableInterrupts == 1 {
			ppu.cpu.triggerInterruptNMI()
		}
		ppu.vBlank = 1
		ppu.spriteOverflow = 0
		ppu.sprite0Hit = 0
		ppu.frameCount++
		ppu.statusRendering = false
	}
}

//Emulate emulates the PPU for a given number of cycles.
func (ppu *PPU) Emulate(cycles int) {
	cyclesLeft := cycles
	for cyclesLeft > 0 {
		ppu.cycles++
		ppu.tickCount++
		if ppu.tickCount == 341 || (ppu.tickCount == 340 && ppu.scanlineCount == -1 && ppu.frameCount%2 == 1) {
			ppu.tickCount = 0
			ppu.scanlineCount++
			if ppu.scanlineCount > 260 {
				ppu.scanlineCount = -1
			}
		}

		ppu.maybePerformVBlank()
		cyclesLeft--

		renderingEnabled := ppu.renderBackground != 0 || ppu.renderSprites != 0

		if ppu.scanlineCount == -1 {
			if ppu.tickCount == 1 {
				// prerender
				ppu.sprite0Hit = 0
				ppu.vBlank = 0
				ppu.spriteOverflow = 0
				ppu.statusRendering = true
			}
			//Copy veritcal to scroll.
			if ppu.tickCount >= 280 && ppu.tickCount <= 304 {
				// v: IHGF.ED CBA..... = t: IHGF.ED CBA.....
				ppu.v = (ppu.v & 0x841F) | (ppu.t & 0x7BE0)
			}
		}

		if renderingEnabled {
			ppu.renderVisibleScanlines()
		}
	}
}

func (ppu *PPU) renderPixel() {
	x, y := ppu.tickCount-1, ppu.scanlineCount

	// background pixel
	backgroundPixel := byte(ppu.backgroundBitmapData >> (32 + ((7 - ppu.x) * 4)) & 0xF)

	// sprite pixel
	spritePixel := byte(0)
	var spriteIndex = 0
	for n := 0; n < ppu.numScanlineSprites; n++ {
		offset := x - int(ppu.spriteXPositions[n])
		if offset >= 0 && offset < 8 {
			attributes := ppu.spriteAttributes[n]
			data := ((ppu.spriteBitmapDataHi[n] & 0x80) >> 6) | ((ppu.spriteBitmapDataLo[n] & 0x80) >> 7)
			ppu.spriteBitmapDataHi[n] <<= 1
			ppu.spriteBitmapDataLo[n] <<= 1
			if data != 0 {
				spritePixel = 0x10 + data + 4*(attributes&0x3)
				spriteIndex = n
				break
			}
		}
	}

	// left screen hiding
	if x < 8 {
		if ppu.showBackgroundLeft == 0 {
			backgroundPixel = 0
		}
		if ppu.showSpritesLeft == 0 {
			spritePixel = 0
		}
	}

	output := byte(0)
	bgVisible, spVisible := backgroundPixel%4 != 0, spritePixel%4 != 0
	if !bgVisible && !spVisible {
		output = 0
	} else if !bgVisible && spVisible {
		output = spritePixel | 0x10
	} else if !spVisible && bgVisible {
		output = backgroundPixel
	} else {
		output = ppu.checkSpriteCollision(spriteIndex, spritePixel, backgroundPixel)
	}

	ppu.funcPushPixel(x, y, ppu.FetchColor(output))
}

func (ppu *PPU) checkSpriteCollision(spriteIndex int, spritePixel byte, backgroundPixel byte) byte {
	if spriteIndex == ppu.spriteZeroAt && ppu.tickCount-1 < 255 {
		ppu.sprite0Hit = 1
	}

	spriteHasPriority := ((ppu.spriteAttributes[spriteIndex] >> 5) & 1)
	if spriteHasPriority == 0 {
		return spritePixel | 0x10
	}
	return backgroundPixel
}

func (ppu *PPU) fetchTileData() {
	// run on (_ % 8 == 1) ticks in prerender and render scanlines
	// we need to fetch a tile AND the attribute data, combine them, and
	// shove them onto our queue of uh, stuff
	nametableAddress := 0x2000 | (ppu.v & 0x0FFF)
	nametableData := ppu.ram.ReadPPU(uint16(nametableAddress))
	//Extract attribute byte.
	attributeAddress := 0x23C0 | (ppu.v & 0x0C00) | ((ppu.v >> 4) & 0x38) | ((ppu.v >> 2) & 0x07)
	attributeData := ppu.ram.ReadPPU(uint16(attributeAddress))
	// process attribute data to select correct tile
	shift := ((ppu.v >> 4) & 4) | (ppu.v & 2)
	attributeData = ((attributeData >> shift) & 3) << 2

	patternAddr := uint16(0)
	patternAddr |= uint16((ppu.v >> 12) & 0x7)
	patternAddr |= uint16(nametableData) << 4
	patternAddr |= uint16(ppu.backgroundTableAddress) << 12
	patternLo, patternHi := ppu.ram.ReadPPU(patternAddr), ppu.ram.ReadPPU(patternAddr+8)

	bitmap := uint32(0)
	for i := 0; i < 8; i++ {
		// shift on the data
		pixelData := attributeData | ((patternLo & 0x80) >> 7) | ((patternHi & 0x80) >> 6)
		patternLo <<= 1
		patternHi <<= 1
		bitmap = (bitmap << 4) | uint32(pixelData)
	}

	ppu.backgroundBitmapData |= uint64(bitmap)
}

func (ppu *PPU) incrementScrollY() {
	if ppu.v&0x7000 != 0x7000 {
		ppu.v += 0x1000
	} else {
		ppu.v &= 0x8FFF
		y := (ppu.v & 0x03E0) >> 5
		if y == 29 {
			y = 0
			ppu.v ^= 0x0800
		} else if y == 31 {
			y = 0
		} else {
			y++
		}
		ppu.v = (ppu.v & 0xFC1F) | (y << 5)
	}
}

func (ppu *PPU) incrementScrollX() {
	if ppu.v&0x001F == 31 {
		ppu.v &= 0xFFE0
		ppu.v ^= 0x0400
	} else {
		ppu.v++
	}
}

//FetchColor grabs a color from the given index.
func (ppu *PPU) FetchColor(index byte) uint32 {
	return ppu.colors[ppu.palette[index&0x1F]]
}
