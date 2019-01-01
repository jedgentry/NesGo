package main

const (
	// NONE No interrupt is currently availible.
	NONE = iota
	// NMI Represents a non maskable interrupt.
	NMI
	// IRQ Represents a interrupt request.
	IRQ
)

const (
	resetVectorAddr = 0xFFFC
	nmiVectorAddr   = 0xFFFA
	brkVectorAddr   = 0xFFFE
	defaultStackPtr = 0xFD
)

//Instructions for CPU
const (
	inclusiveOr = iota
	and
	exclusiveOr
	addWithCarry
	storeAccumulator
	loadAccumulator
	compare
	subtractWithCarry
)

const (
	arithmeticShiftLeft = iota
	rotateLeft
	logicalShiftRight
	rotateRight
	storeXRegister
	loadXRegister
	decrementMemory
	incrementMemory
)

const (
	getImmediateAddress = 0
	getZeroPageAddress  = 1
	getAbsoluteAddress  = 3
	indirectJump        = 5
	getAbsoluteX
)

const (
	breakInstruction           = 0x00
	jumpSubroutine             = 0x20
	returnFromInterrupt        = 0x40
	returnFromSubroutine       = 0x60
	pushProcessorStatus        = 0x08
	pullProcessorStatus        = 0x28
	pushAccumulator            = 0x48
	pullAccumulator            = 0x68
	decrementYRegister         = 0x88
	xferAccumulatorToYRegister = 0xA8
	incrementYRegister         = 0xC8
	incrementXRegister         = 0xE8
	clearCarryFlag             = 0x18
	setCarryFlag               = 0x38
	clearInterruptDisable      = 0x58
	setInterruptDisable        = 0x78
	xferYToAccumulator         = 0x98
	clearOverflowFlag          = 0xB8
	clearDecimalMode           = 0xD8
	setDecimalFlag             = 0xF8
	xferXtoStackPtr            = 0x9A
	xferStackPtrToX            = 0xBA
	decrementY                 = 0xCA
	noop                       = 0xEA
)

// CPU represents the NES CPU.
type CPU struct {
	memory      *Memory
	accumulator byte
	x           byte
	y           byte
	pc          uint16
	sp          byte

	ram *Memory

	carry            bool
	zero             bool
	interruptEnabled bool
	bcdEnabled       bool
	overflow         bool
	negative         bool

	totalCycles      uint64
	pendingInterrupt int
	suspended        int
}

func (system *System) resetCPU() {
	system.cpu = CPU{
		accumulator:      0,
		x:                0,
		y:                0,
		pc:               0,
		sp:               defaultStackPtr,
		ram:              &system.memory,
		carry:            false,
		zero:             false,
		interruptEnabled: false,
		bcdEnabled:       false,
		overflow:         false,
		negative:         false,
		totalCycles:      0,
		pendingInterrupt: 0,
		suspended:        0,
	}
}

func (cpu *CPU) getVectorReset() uint16 {
	return uint16(cpu.ram.ReadUint16(resetVectorAddr))
}

func (cpu *CPU) getVectorNMI() uint16 {
	return uint16(cpu.ram.ReadUint16(nmiVectorAddr))
}

func (cpu *CPU) getVectorBRK() uint16 {
	return uint16(cpu.ram.ReadUint16(brkVectorAddr))
}

func (cpu *CPU) addressImmediate() (uint16, int) {
	return cpu.pc + 1, 2
}

func (cpu *CPU) addressZeroPage() (uint16, int) {
	return uint16(cpu.ram.ReadByte(cpu.pc + 1)), 2
}

func (cpu *CPU) addressZeroPageX() (uint16, int) {
	return uint16((cpu.ram.ReadByte(cpu.pc+1) + cpu.x) & 0xFF), 2
}

func (cpu *CPU) addressZeroPageY() (uint16, int) {
	return uint16((cpu.ram.ReadByte(cpu.pc+1) + cpu.y) & 0xFF), 2
}

func (cpu *CPU) addressRelative() (uint16, int) {
	offset := int8(cpu.ram.ReadByte(cpu.pc + 1))
	return uint16(int32(cpu.pc) + int32(offset)), 2
}

func (cpu *CPU) addressAbsolute() (uint16, int) {
	return uint16(cpu.ram.ReadUint16(cpu.pc + 1)), 3
}

func (cpu *CPU) addressAbsoluteX() (uint16, int, bool) {
	addr := uint16(cpu.ram.ReadUint16(cpu.pc+1) + uint16(cpu.x))

	pageCrossed := false
	if uint16(addr&0xFF)+uint16(cpu.x) > 255 {
		cpu.ram.ReadByte(addr)
		pageCrossed = true
	}
	return addr, 3, pageCrossed
}

func (cpu *CPU) addressAbsoluteY() (uint16, int, bool) {
	addr := uint16(cpu.ram.ReadUint16(cpu.pc+1) + uint16(cpu.y))

	pageCrossed := false
	if uint16(addr&0xFF)+uint16(cpu.x) > 255 {
		cpu.ram.ReadByte(addr)
		pageCrossed = true
	}
	return addr, 3, pageCrossed
}

func (cpu *CPU) readUint16Bugged(addr uint16) uint16 {
	low := cpu.ram.ReadByte(addr)
	high := cpu.ram.ReadByte((addr & 0xFF00) | uint16(byte(addr)+1))
	return uint16(high)<<8 | uint16(low)
}

func (cpu *CPU) addressIndirectX() (uint16, int) {
	addr := uint16((cpu.ram.ReadByte(cpu.pc+1) + cpu.x) & 0xFF)
	return uint16(cpu.readUint16Bugged(addr)), 2
}

func (cpu *CPU) addressIndirectY() (uint16, int, bool) {
	addr := uint16(cpu.ram.ReadByte(cpu.pc + 1))
	addr = uint16(cpu.readUint16Bugged(addr) + uint16(cpu.y))

	pageCrossed := false
	if uint16(addr&0xFF)+uint16(cpu.y) > 255 {
		cpu.ram.ReadByte(addr) // ? is this right?
		pageCrossed = true
	}
	return addr, 2, pageCrossed
}

func (cpu *CPU) stackPush(data byte) {
	cpu.ram.WriteByte(uint16(0x0100+uint16(cpu.sp)), data)
	cpu.sp--
}

func (cpu *CPU) stackPull() byte {
	cpu.sp++
	return cpu.ram.ReadByte(uint16(0x0100 + uint16(cpu.sp)))
}

func (cpu *CPU) statusPack(bFlag bool) (data byte) {
	if cpu.carry {
		data |= 1 << 0
	}
	if cpu.zero {
		data |= 1 << 1
	}
	if cpu.interruptEnabled {
		data |= 1 << 2
	}
	if cpu.bcdEnabled {
		data |= 1 << 3
	}
	if bFlag {
		data |= 1 << 4
	}
	data |= 1 << 5
	if cpu.overflow {
		data |= 1 << 6
	}
	if cpu.negative {
		data |= 1 << 7
	}
	return
}

func (cpu *CPU) statusUnpack(data byte) {
	cpu.carry = data&(1<<0) > 0
	cpu.zero = data&(1<<1) > 0
	cpu.interruptEnabled = data&(1<<2) > 0
	cpu.bcdEnabled = data&(1<<3) > 0
	cpu.overflow = data&(1<<6) > 0
	cpu.negative = data&(1<<7) > 0
}

func (cpu *CPU) handleInterrupt(addr uint16) {
	cpu.stackPush(byte((cpu.pc >> 8) & 0xFF))
	cpu.stackPush(byte(cpu.pc & 0xFF))
	cpu.stackPush(cpu.statusPack(false))
	cpu.interruptEnabled = true
	cpu.pc = addr
	cpu.pendingInterrupt = NONE
}

func (cpu *CPU) triggerInterruptNMI() {
	cpu.pendingInterrupt = NMI
}

func (cpu *CPU) triggerInterruptIRQ() {
	if !cpu.interruptEnabled {
		cpu.pendingInterrupt = IRQ
	}
}

func (cpu *CPU) handleInterrupts() {
	switch cpu.pendingInterrupt {
	case NMI:
		cpu.handleInterrupt(cpu.getVectorNMI())
	case IRQ:
		cpu.handleInterrupt(cpu.getVectorBRK())
	}
}

func (cpu *CPU) handleAccumulator(opcode byte, cyclesLeft *int) {
	var size, cycles int
	var addr uint16
	addressType, instructionType := (opcode>>2)&0x7, (opcode>>5)&0x7
	if addressType != 4 && addressType != 6 {
		switch addressType {
		case 0:
			addr, size = cpu.addressImmediate()
			cycles = 2
		case 1:
			addr, size = cpu.addressZeroPage()
			if instructionType != 4 && instructionType != 5 {
				cycles = 5
			} else {
				cycles = 3
			}
		case 2:
			// ACCUMULATOR!!!
			cycles = 2
			size = 1
		case 3:
			addr, size = cpu.addressAbsolute()
			if instructionType != 4 && instructionType != 5 {
				cycles = 6
			} else {
				cycles = 4
			}
		case 5:
			if instructionType != 4 && instructionType != 5 {
				addr, size = cpu.addressZeroPageX()
				cycles = 6
			} else {
				addr, size = cpu.addressZeroPageY()
				cycles = 4
			}
		case 7:
			var pageCrossed bool
			if instructionType != 4 && instructionType != 5 {
				addr, size, pageCrossed = cpu.addressAbsoluteX()
				cycles = 7
			} else {
				addr, size, pageCrossed = cpu.addressAbsoluteY()
				cycles = 4
				if pageCrossed {
					cycles++
				}
			}
		}

		var data byte
		if addressType != 2 {
			data = cpu.ram.ReadByte(addr)
		} else {
			data = cpu.accumulator
		}

		switch instructionType {
		case arithmeticShiftLeft:
			cpu.carry = (data & 0x80) > 0
			data = data << 1
		case rotateLeft:
			oldCarry := cpu.carry
			cpu.carry = (data & 0x80) > 0
			data = data << 1
			if oldCarry {
				data |= 1
			}
		case logicalShiftRight:
			cpu.carry = (data & 0x1) > 0
			data = data >> 1
		case rotateRight:
			oldCarry := cpu.carry
			cpu.carry = (data & 0x1) > 0
			data = data >> 1
			if oldCarry {
				data |= 0x80
			}
		case storeXRegister:
			// (also TXA when mode = accumulator)
			data = cpu.x
			if addressType == 2 {
				size = 1
			}
		case loadXRegister:
			// also TAX when mode = accumulator
			cpu.x = data
			if addressType == 2 {
				size = 1
			}
		case decrementMemory:
			data--
		case incrementMemory:
			data++
		}

		if addressType != 2 {
			if instructionType != 5 {
				cpu.ram.WriteByte(addr, data)
			}
		} else {
			cpu.accumulator = data
		}

		if instructionType != 4 || addressType == 2 {
			cpu.zero = data == 0
			cpu.negative = (data & 0x80) > 0
		}

		cpu.pc += uint16(size)
		*cyclesLeft = *cyclesLeft - cycles
	}
}

func (cpu *CPU) handleMemoryAccess(opcode byte, cyclesLeft *int) {
	var size, cycles int
	var addr uint16
	addressType, instructionType := (opcode>>2)&0x7, (opcode>>5)&0x7
	if instructionType != 0 && addressType != 2 && addressType != 4 && addressType != 6 {
		switch addressType {
		case 0:
			addr, size = cpu.addressImmediate()
			cycles = 2
		case 1:
			addr, size = cpu.addressZeroPage()
			cycles = 3
		case 3:
			if instructionType != 3 {
				addr, size = cpu.addressAbsolute()
				cycles = 4
				if instructionType == 2 {
					cycles = 3
				}
			} else {
				// jump indirect
				addr, size = uint16(cpu.ram.ReadUint16(cpu.pc+1)), 3
				addr = uint16(cpu.readUint16Bugged(addr))
				cycles = 5
			}
		case 5:
			addr, size = cpu.addressZeroPageX()
			cycles = 4
		case 7:
			var pageCrossed bool
			addr, size, pageCrossed = cpu.addressAbsoluteX()
			cycles = 4
			if pageCrossed {
				cycles++
			}
		}

		switch instructionType {
		case 1:
			// BIT
			data := cpu.ram.ReadByte(addr)
			cpu.zero = (cpu.accumulator & data) == 0
			cpu.overflow = (data & 0x40) > 0
			cpu.negative = (data & 0x80) > 0
		case 2:
			// JMP
			cpu.pc = addr
		case 3:
			// JMP
			cpu.pc = addr
		case 4:
			// STY
			cpu.ram.WriteByte(addr, cpu.y)
		case 5:
			// LDY
			cpu.y = cpu.ram.ReadByte(addr)
			cpu.zero = cpu.y == 0
			cpu.negative = (cpu.y & 0x80) > 0
		case 6:
			// CPY
			data := cpu.ram.ReadByte(addr)
			cpu.carry = cpu.y >= data
			cpu.zero = cpu.y == data
			cpu.negative = (cpu.y-data)&0x80 > 0
		case 7:
			// CPX
			data := cpu.ram.ReadByte(addr)
			cpu.carry = cpu.x >= data
			cpu.zero = cpu.x == data
			cpu.negative = (cpu.x-data)&0x80 > 0
		}

		if instructionType != 2 && instructionType != 3 {
			cpu.pc += uint16(size)
		}
		*cyclesLeft = *cyclesLeft - cycles
	}
}

//Handle misc processor instructions (jumps, interrupts, etc.) General bookeeping.
func (cpu *CPU) handleMiscInstructions(opcode byte, cyclesLeft *int) {
	switch opcode {
	case breakInstruction:
		returnAddr := cpu.pc + 2
		cpu.stackPush(byte((returnAddr >> 8) & 0xFF))
		cpu.stackPush(byte(returnAddr & 0xFF))
		cpu.stackPush(cpu.statusPack(true))
		cpu.interruptEnabled = true // disable interrupts
		*cyclesLeft -= 7
		cpu.pc = cpu.getVectorBRK()
	case jumpSubroutine:
		addr := uint16(cpu.ram.ReadUint16(cpu.pc + 1))
		returnAddr := cpu.pc + 3 - 1 // returnAddr minus one
		cpu.stackPush(byte((returnAddr >> 8) & 0xFF))
		cpu.stackPush(byte(returnAddr & 0xFF))
		*cyclesLeft -= 6
		cpu.pc = addr
	case returnFromInterrupt:
		cpu.statusUnpack(cpu.stackPull())
		cpu.pc = uint16(cpu.stackPull()) + uint16(cpu.stackPull())*256
		*cyclesLeft -= 6
	case returnFromSubroutine:
		cpu.pc = (uint16(cpu.stackPull()) + uint16(cpu.stackPull())*256) + 1
		*cyclesLeft -= 6
	case pushProcessorStatus:
		cpu.stackPush(cpu.statusPack(true))
		cpu.pc++
		*cyclesLeft -= 3
	case pullProcessorStatus:
		cpu.statusUnpack(cpu.stackPull())
		cpu.pc++
		*cyclesLeft -= 4
	case pushAccumulator:
		cpu.stackPush(cpu.accumulator)
		cpu.pc++
		*cyclesLeft -= 3
	case pullAccumulator:
		cpu.accumulator = cpu.stackPull()
		cpu.zero = cpu.accumulator == 0
		cpu.negative = (cpu.accumulator & 0x80) > 0
		cpu.pc++
		*cyclesLeft -= 4
	case decrementYRegister:
		cpu.y--
		cpu.zero = cpu.y == 0
		cpu.negative = (cpu.y & 0x80) > 0
		cpu.pc++
		*cyclesLeft -= 2
	case xferAccumulatorToYRegister:
		cpu.y = cpu.accumulator
		cpu.zero = cpu.y == 0
		cpu.negative = (cpu.y & 0x80) > 0
		cpu.pc++
		*cyclesLeft -= 2
	case incrementYRegister:
		cpu.y++
		cpu.zero = cpu.y == 0
		cpu.negative = (cpu.y & 0x80) > 0
		cpu.pc++
		*cyclesLeft -= 2
	case incrementXRegister:
		cpu.x++
		cpu.zero = cpu.x == 0
		cpu.negative = (cpu.x & 0x80) > 0
		cpu.pc++
		*cyclesLeft -= 2
	case clearCarryFlag:
		cpu.carry = false
		cpu.pc++
		*cyclesLeft -= 2
	case setCarryFlag:
		cpu.carry = true
		cpu.pc++
		*cyclesLeft -= 2
	case clearInterruptDisable:
		cpu.interruptEnabled = false
		cpu.pc++
		*cyclesLeft -= 2
	case setInterruptDisable:
		cpu.interruptEnabled = true
		cpu.pc++
		*cyclesLeft -= 2
	case xferYToAccumulator:
		cpu.accumulator = cpu.y
		cpu.zero = cpu.accumulator == 0
		cpu.negative = (cpu.accumulator & 0x80) > 0
		cpu.pc++
		*cyclesLeft -= 2
	case clearOverflowFlag:
		cpu.overflow = false
		cpu.pc++
		*cyclesLeft -= 2
	case clearDecimalMode:
		cpu.bcdEnabled = false
		cpu.pc++
		*cyclesLeft -= 2
	case setDecimalFlag:
		cpu.bcdEnabled = true
		cpu.pc++
		*cyclesLeft -= 2
	case xferXtoStackPtr:
		cpu.sp = cpu.x
		cpu.pc++
		*cyclesLeft -= 2
	case xferStackPtrToX:
		cpu.x = cpu.sp
		cpu.zero = cpu.x == 0
		cpu.negative = (cpu.x & 0x80) > 0
		cpu.pc++
		*cyclesLeft -= 2
	case decrementY:
		cpu.x--
		cpu.zero = cpu.x == 0
		cpu.negative = (cpu.x & 0x80) > 0
		cpu.pc++
		*cyclesLeft -= 2
	case noop:
		cpu.pc++
		*cyclesLeft -= 2
	case 0x30:
		var address uint16
		//TODO: May need to handle third argument since it's page crossed.
		address, _, _ = cpu.getAddressMode(opcode, cyclesLeft)
		if cpu.negative {
			cpu.pc = address
		} else {
			cpu.pc++
		}
		*cyclesLeft--
	case 0x04:
		fallthrough
	case 0x44:
		fallthrough
	case 0x64:
		//NOP
		cpu.pc += 2
		*cyclesLeft -= 3
	case 0x0C:
		fallthrough
	case 0x1C:
		fallthrough
	case 0x3C:
		fallthrough
	case 0x5C:
		fallthrough
	case 0x7C:
		fallthrough
	case 0xDC:
		fallthrough
	case 0xFC:
		//NOP
		cpu.pc += 3
		*cyclesLeft -= 4
	case 0x14:
		fallthrough
	case 0x34:
		fallthrough
	case 0x54:
		fallthrough
	case 0x74:
		fallthrough
	case 0xD4:
		fallthrough
	case 0xF4:
		//NOP
		cpu.pc += 2
		*cyclesLeft -= 4
	case 0x1A:
		fallthrough
	case 0x3A:
		fallthrough
	case 0x5A:
		fallthrough
	case 0x7A:
		fallthrough
	case 0xDA:
		fallthrough
	case 0xFA:
		//NOP
		cpu.pc++
		*cyclesLeft -= 2
	case 0xFB:
		return
	default:
		if opcode&0x3 == 2 {
			cpu.handleAccumulator(opcode, cyclesLeft)
		} else if opcode&0x2 == 0 {
			cpu.handleMemoryAccess(opcode, cyclesLeft)
		} else if opcode&0x1F == 0x10 {
			cpu.handleBranchOperations(opcode, cyclesLeft)
		} else {
			panic("Unhandled opcode!")
		}
	}
}

func (cpu *CPU) handleBranchOperations(opcode byte, cyclesLeft *int) {
	var flag bool
	switch (opcode & 0xC0) >> 6 {
	case 0:
		flag = cpu.negative
	case 1:
		flag = cpu.overflow
	case 2:
		flag = cpu.carry
	case 3:
		flag = cpu.zero
	default:
		panic("Invalid branch flag.")
	}
	comp := opcode&0x20 > 0

	if comp == flag {
		offset := int8(cpu.ram.ReadByte(cpu.pc + 1))
		destination := uint16(int32(cpu.pc)+int32(offset)) + 2
		if destination&0xFF00 != (cpu.pc+2)&0xFF00 {
			*cyclesLeft -= 4
		} else {
			*cyclesLeft -= 3
		}
		cpu.pc = destination
	} else {
		cpu.pc += 2
		*cyclesLeft -= 2
	}
}

//Get Address mode sets the address mode of the read or write.
func (cpu *CPU) getAddressMode(opcode byte, cycles *int) (uint16, int, byte) {
	addressType, instructionType := (opcode>>2)&0x7, (opcode>>5)&0x7
	addr := uint16(0)
	size := 0
	switch addressType {
	case 0:
		addr, size = cpu.addressIndirectX()
		*cycles = 6
	case 1:
		addr, size = cpu.addressZeroPage()
		*cycles = 3
	case 2:
		addr, size = cpu.addressImmediate()
		*cycles = 2
	case 3:
		addr, size = cpu.addressAbsolute()
		*cycles = 4
	case 4:
		pageCrossed := false
		addr, size, pageCrossed = cpu.addressIndirectY()
		*cycles = 5
		if pageCrossed || instructionType == 4 {
			*cycles++
		}
	case 5:
		addr, size = cpu.addressZeroPageX()
	case 6:
		pageCrossed := false
		addr, size, pageCrossed = cpu.addressAbsoluteY()
		*cycles = 4
		if pageCrossed || instructionType == 4 {
			*cycles++
		}
	case 7:
		pageCrossed := false
		addr, size, pageCrossed = cpu.addressAbsoluteX()
		*cycles = 4
		if pageCrossed || instructionType == 4 {
			*cycles++
		}
	}

	return addr, size, instructionType
}

func (cpu *CPU) handleAddressInstructionType(address uint16, instructionType byte) {
	switch instructionType {
	case inclusiveOr:
		cpu.accumulator |= cpu.ram.ReadByte(address)
	case and:
		cpu.accumulator &= cpu.ram.ReadByte(address)
	case exclusiveOr:
		cpu.accumulator ^= cpu.ram.ReadByte(address)
	case addWithCarry:
		a := cpu.accumulator
		b := cpu.ram.ReadByte(address)
		c := byte(0)
		if cpu.carry {
			c = 1
		}
		cpu.accumulator = a + b + c
		cpu.carry = int(a)+int(b)+int(c) > 0xFF
		cpu.overflow = (a^b)&0x80 == 0 && (a^cpu.accumulator)&0x80 != 0
	case storeAccumulator:
		cpu.ram.WriteByte(address, cpu.accumulator)
	case loadAccumulator:
		cpu.accumulator = cpu.ram.ReadByte(address)
	case compare:
		data := cpu.ram.ReadByte(address)
		cpu.carry = cpu.accumulator >= data
		cpu.zero = cpu.accumulator == data
		cpu.negative = (cpu.accumulator-data)&0x80 > 0
	case subtractWithCarry:
		a := cpu.accumulator
		b := cpu.ram.ReadByte(address)
		c := byte(0)
		if cpu.carry {
			c = 1
		}

		cpu.accumulator = a - b - (1 - c)
		cpu.carry = int(a)-int(b)-int(1-c) >= 0
		cpu.overflow = (a^b)&0x80 != 0 && (a^cpu.accumulator)&0x80 != 0
	}
}

func (cpu *CPU) handleMemoryOpcode(opcode byte, cyclesLeft *int) {
	if (opcode & 0x3) == 1 {
		cycles := 0
		address, size, instructionType := cpu.getAddressMode(opcode, &cycles)
		cpu.handleAddressInstructionType(address, instructionType)

		if instructionType != 4 && instructionType != 6 {
			cpu.zero = cpu.accumulator == 0
			cpu.negative = (cpu.accumulator & 0x80) > 0
		}

		cpu.pc += uint16(size)
		*cyclesLeft -= cycles
	}
}

//Emulate emulates the CPU for a number of cycles, returning the amount of cycles emulated.
func (cpu *CPU) Emulate(cycles int) int {
	//See how many cycles we have to emulate.
	cyclesLeft := cycles
	//Emulate while cycles left is not zero.
	for cyclesLeft > 0 {
		//Handle suspension case.
		if cpu.suspended > 0 {
			cpu.suspended--
		} else {
			//Handle pending interrupts.
			cpu.handleInterrupts()
			//Read our next opcode.
			opcode := cpu.ram.ReadByte(cpu.pc)
			//Perform our next opcode.
			cpu.handleMemoryOpcode(opcode, &cyclesLeft)
			cpu.handleMiscInstructions(opcode, &cyclesLeft)
		}
	}
	//Return how many cycles we emulated.
	cyclesThisTick := cycles - cyclesLeft
	cpu.totalCycles += uint64(cyclesThisTick)
	return cyclesThisTick
}
