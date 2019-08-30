package cpu

type Registers struct {
	A uint8
	B uint8
	D uint8
	H uint8
	C uint8
	E uint8
	L uint8
}

type Z80Flag uint8

type SpecialRegisters struct {
	F Z80Flag

	IFF1 bool
	IFF2 bool

	I uint8
	R uint8

	IX uint16
	IY uint16

	SP uint16
	PC uint16
}
