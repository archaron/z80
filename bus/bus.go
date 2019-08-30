package bus

import (
	"fmt"
	"os"
)

type Bus struct {
	ram     [64 * 1024]byte
	rom     []byte
	romSize uint16
}

func NewBus(romFile string) *Bus {
	b := &Bus{}

	f, err := os.Open(romFile)
	if err != nil {
		panic(err)
	}

	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}

	b.rom = make([]byte, fi.Size())
	s, err := f.Read(b.rom)
	if err != nil {
		panic(err)
	}
	b.romSize = uint16(s)

	fmt.Printf("ROM size = %d\n", b.romSize)

	return b
}

func (b *Bus) Dump(addr uint16, len uint16) {
	fmt.Println("====================================")
	for i := addr; i < addr+len; i++ {
		if i%16 == 0 {
			fmt.Printf("\n%04x | ", i)
		}
		fmt.Printf("%02x ", b.Read(i))
	}
	fmt.Println("\n====================================")
}

func (b *Bus) Write(addr uint16, data uint8) {
	if addr < b.romSize {
		return
	}
	b.ram[addr] = data
}

func (b *Bus) Read(addr uint16) uint8 {
	if addr < b.romSize {
		return b.rom[addr]
	}

	return b.ram[addr]
}
