package edf

func uint32FromBytes(in []byte) uint32 {
	ret := uint32(0)
	ret |= uint32(in[0]) << 24
	ret |= uint32(in[1]) << 16
	ret |= uint32(in[2]) << 8
	ret |= uint32(in[3])
	return ret
}

func uint32ToBytes(in uint32, out []byte) {
	out[0] = byte(in & (0xFF << 24) >> 24)
	out[1] = byte(in & (0xFF << 16) >> 16)
	out[2] = byte(in & (0xFF << 8) >> 8)
	out[3] = byte(in & 0xFF)
}