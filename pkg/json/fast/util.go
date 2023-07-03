package fast

// bsearch16 returns the smallest i such that a[i] >= x.
// If there is no such i, bsearch16 returns len(a).
func bsearch16(a []uint16, x uint16) int {
	i, j := 0, len(a)
	for i < j {
		h := i + (j-i)>>1
		if a[h] < x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}

// bsearch32 returns the smallest i such that a[i] >= x.
// If there is no such i, bsearch32 returns len(a).
func bsearch32(a []uint32, x uint32) int {
	i, j := 0, len(a)
	for i < j {
		h := i + (j-i)>>1
		if a[h] < x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}

const supplementalPlanesOffset = 0x10000
const highSurrogateOffset = 0xD800
const lowSurrogateOffset = 0xDC00

const basicMultilingualPlaneReservedOffset = 0xDFFF
const basicMultilingualPlaneOffset = 0xFFFF

func combineUTF16Surrogates(high, low rune) rune {
	return supplementalPlanesOffset + (high-highSurrogateOffset)<<10 + (low - lowSurrogateOffset)
}
