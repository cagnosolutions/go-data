package bits

func SwapByteOrder(n uint32) uint32 {
	var b0 = (n & 0x000000ff) >> 0
	var b1 = (n & 0x0000ff00) >> 8
	var b2 = (n & 0x00ff0000) >> 16
	var b3 = (n & 0xff000000) >> 24
	return b0<<24 | b1<<16 | b2<<8 | b3<<0
}

func ModBit32Lo(n uint32, pos uint32, on bool) uint32 {
	var b uint32
	if on {
		b = 1
	}
	mask := uint32(1 << pos)
	return (n & ^mask) | (b << pos)
}

func ModBit32Hi(n uint32, pos uint32, on bool) uint32 {
	var b uint32
	if on {
		b = 1
	}
	mask := 1<<31 - pos
	return (n & ^mask) | (b<<31 - pos)
}

func ModBit16Lo(n uint16, pos uint16, on bool) uint16 {
	var b uint16
	if on {
		b = 1
	}
	mask := uint16(1 << pos)
	return (n & ^mask) | (b << pos)
}

func ModBit16Hi(n uint16, pos uint16, on bool) uint16 {
	var b uint16
	if on {
		b = 1
	}
	mask := 1<<15 - pos
	return (n & ^mask) | (b<<15 - pos)
}
