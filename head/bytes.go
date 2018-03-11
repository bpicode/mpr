package head

func uint16ToBytes(i uint16) []byte {
	return []byte{uint8(i >> 8), uint8(i & 0xff)}
}

func uint32ToBytes(i uint32) []byte {
	return []byte{
		uint8(i >> 24),
		uint8((i >> 16) & 0xff),
		uint8((i >> 8) & 0xff),
		uint8(i & 0xff),
	}
}
