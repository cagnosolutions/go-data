package dio

import (
	"log"
	"unsafe"
)

// alignTo returns an integer representing a byte offset for a region that
// is aligned on a boundary that is consistent with the provided size.
func alignTo(block []byte, size int) int {
	return int(uintptr(unsafe.Pointer(&block[0])) & uintptr(AlignSize-1))
}

// isAligned returns a boolean indicating true if the provided slice is
// correctly aligned on a AlignSize boundary.
func isAligned(block []byte) bool {
	return alignTo(block, AlignSize) == 0
}

// AlignedBlock allocates and returns a slice of []byte that has the length
// and capacity provided by size.
func AlignedBlock(BlockSize int) []byte {
	block := make([]byte, BlockSize+AlignSize)
	if AlignSize == 0 {
		return block
	}
	a := alignTo(block, AlignSize)
	offset := 0
	if a != 0 {
		offset = AlignSize - a
	}
	block = block[offset : offset+BlockSize]
	// Can't check alignment of a zero sized block
	if BlockSize != 0 {
		if !isAligned(block) {
			log.Fatal("Failed to align block")
		}
	}
	return block
}
