package cpu

import (
	"fmt"
	"github.com/archaron/z80/bus"
)

type Z80Cpu struct {
	bus        *bus.Bus
	regs       [2]Registers
	CurrentSet uint8
	cycles     uint8
	SpecialRegisters
}

const (
	C  Z80Flag = 1 << 0 // Carry
	N  Z80Flag = 1 << 1 // Subtract
	PV Z80Flag = 1 << 2 // Parity / Overflow
	F3 Z80Flag = 1 << 3 // Unused
	H  Z80Flag = 1 << 4 // Half Carry
	F5 Z80Flag = 1 << 5 // Unused
	Z  Z80Flag = 1 << 6 // Zero
	S  Z80Flag = 1 << 7 // Sign
)

func NewZ80CPU(b *bus.Bus) *Z80Cpu {
	z80 := &Z80Cpu{}
	z80.ConnectBus(b)
	z80.Reset()
	return z80
}

func (c *Z80Cpu) ConnectBus(b *bus.Bus) {
	c.bus = b
}

func (c *Z80Cpu) write(addr uint16, data uint8) {
	c.bus.Write(addr, data)
}

func (c *Z80Cpu) read(addr uint16) uint8 {
	return c.bus.Read(addr)
}

func (c *Z80Cpu) read8_next() uint8 {
	data := c.bus.Read(c.PC)
	c.PC++
	return data
}

func (c *Z80Cpu) read16_next() uint16 {
	op1 := c.bus.Read(c.PC)
	c.PC++
	op2 := c.bus.Read(c.PC)
	c.PC++
	return uint16(op1) + uint16(op2)<<8
}

func (c *Z80Cpu) GetFlag(flag Z80Flag) bool {
	return c.F&flag == flag
}

func (c *Z80Cpu) SetFlag(flag Z80Flag, value bool) {
	if !value {
		c.F &^= flag
	} else {
		c.F = c.F | flag
	}
}

func (c *Z80Cpu) DumpRegs() {
	regNames := [8]string{
		"C",
		"N",
		"PV",
		"F3",
		"H",
		"F5",
		"Z",
		"S",
	}

	fmt.Printf("PC: %04x    SP: %04x     F: %02x    R: %04x\n", c.PC, c.SP, c.F, c.R)
	fmt.Println("________________________________________________________")
	var i uint
	for i = 0; i < 8; i++ {
		fmt.Printf("%s\t", regNames[i])
	}
	fmt.Println()
	for i = 0; i < 8; i++ {
		if c.GetFlag(1 << i) {
			fmt.Printf("1\t")
		} else {
			fmt.Printf("0\t")
		}
	}
	fmt.Println()
	fmt.Println("________________________________________________________")

	fmt.Printf("A: %02x   A`: %02x\n", c.regs[0].A, c.regs[1].A)
	fmt.Printf("B: %02x   B`: %02x\n", c.regs[0].B, c.regs[1].B)
	fmt.Printf("C: %02x   C`: %02x\n", c.regs[0].C, c.regs[1].C)
	fmt.Printf("D: %02x   D`: %02x\n", c.regs[0].D, c.regs[1].D)
	fmt.Printf("H: %02x   H`: %02x\n", c.regs[0].H, c.regs[1].H)
	fmt.Printf("L: %02x   L`: %02x\n", c.regs[0].L, c.regs[1].L)
	fmt.Println()
	fmt.Printf("BC: %04x   BC`: %04x\n", uint16(c.regs[0].C)|uint16(c.regs[0].B)<<8, uint16(c.regs[1].C)|uint16(c.regs[1].B)<<8)
	fmt.Printf("DE: %04x   DE`: %04x\n", uint16(c.regs[0].E)|uint16(c.regs[0].D)<<8, uint16(c.regs[1].E)|uint16(c.regs[1].D)<<8)
	fmt.Printf("HL: %04x   HL`: %04x\n", uint16(c.regs[0].L)|uint16(c.regs[0].H)<<8, uint16(c.regs[1].L)|uint16(c.regs[1].H)<<8)
	fmt.Println()
	fmt.Printf("IX %04x    IY %04x\n", c.IX, c.IY)
	fmt.Println("________________________________________________________")
}

func (c *Z80Cpu) Reset() {
	c.PC = 0
	c.I = 0
	c.R = 0

}

func (c *Z80Cpu) Clock() {
	if c.cycles == 0 {
		var (
			opcode       uint8
			prefix       uint16
			displacement int8
		)

		opcode = c.read(c.PC)
		c.PC++

		if opcode == 0xDD || opcode == 0xED || opcode == 0xFD || opcode == 0xCB {
			o := c.read8_next()

			if (opcode == 0xDD || opcode == 0xFD) && o == 0xCB {
				// DD CB dd oo
				prefix = uint16(opcode<<8) + 0xCB

				displacement = int8(c.read8_next())
				c.PC++

			} else {
				// CB oo   |  ED oo
				// DD oo   |  FD oo
				prefix = uint16(opcode)
				opcode = o
				displacement = 0
			}

			opcode = c.read8_next()

		}

		c.decode(prefix, displacement, opcode)
	}
}

func (c *Z80Cpu) decode(prefix uint16, displacement int8, opcode uint8) (instruction string) {
	x := opcode & 0xC0 >> 6 // 11000000
	y := opcode & 0x38 >> 3 // 00111000
	z := opcode & 0x7       // 00000111
	p := opcode & 0x30 >> 4 // 00110000
	q := opcode & 0x8 >> 3  // 00110000

	fmt.Printf("PREFIX [%04X]  OPCODE [%02X] DISPLACEMENT [%d]\n", prefix, opcode, displacement)
	fmt.Printf("x=%d, y=%d, z=%d, p=%d, q=%d\n", x, y, z, p, q)

	r := [...]string{
		"B",
		"C",
		"D",
		"E",
		"H",
		"L",
		"(HL)",
		"A",
	}

	alu := [...]string{
		"ADD A,",
		"ADC A,",
		"SUB",
		"SBC A,",
		"AND",
		"XOR",
		"OR",
		"CP",
	}

	rp := [...]string{
		"BC",
		"DE",
		"HL",
		"SP",
	}

	//rp2:=[...]string{
	//	"BC",
	//	"DE",
	//	"HL",
	//	"AF",
	//}

	if prefix == 0 {
		switch x {
		case 0:
			switch z {
			case 0:

			case 1:
				if q == 0 {
					nn := c.read16_next()

					instruction = "LD " + rp[p] + ", " + fmt.Sprintf("0x%04X", nn)
				} else {
					instruction = "ADD HL, " + rp[p]
				}
			case 2:
			case 3:

			case 4:
			case 5:
			case 6:
			case 7:
			}
		case 1:

		case 2:
			instruction = alu[y] + " " + r[z]
		case 3:
			switch z {
			case 0:
			case 1:
			case 2:
			case 3:
				switch y {
				case 0:
					nn := c.read16_next()
					instruction = "JP " + fmt.Sprintf("0x%04X", nn)
				case 1:
				case 2:
				case 3:

				case 4:
					instruction = "EX (SP). HL"
				case 5:
					instruction = "EX DE,HL"
				case 6:
					instruction = "DI"
				case 7:
					instruction = "EI"
				}
			case 4:
			case 5:
			case 6:
			case 7:
			}

		}
	}

	fmt.Printf("%s\n", instruction)

	return

}
