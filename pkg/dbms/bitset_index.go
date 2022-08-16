package dbms

import (
	"encoding/binary"
	"fmt"
	"io"
	mathbits "math/bits"
	"os"
	"strconv"
)

const (
	bitsetWS   = 64
	bitsetL2   = 6
	bitsetSize = 16
)

type BitsetIndex [bitsetSize]uint64

func NewBitsetIndex() *BitsetIndex {
	return new(BitsetIndex)
}

// (n >> i) & 1
// (1 << (i & (bitsetWS - 1)))

func (b *BitsetIndex) HasBit(i uint) bool {
	return ((*b)[i>>bitsetL2] & (1 << (i & (bitsetWS - 1)))) != 0
}

func (b *BitsetIndex) SetBit(i uint) {
	(*b)[i>>bitsetL2] |= 1 << (i & (bitsetWS - 1))
}

func (b *BitsetIndex) GetBit(i uint) uint64 {
	return (*b)[i>>bitsetL2] & (1 << (i & (bitsetWS - 1)))
}

func (b *BitsetIndex) UnsetBit(i uint) {
	(*b)[i>>bitsetL2] &^= 1 << (i & (bitsetWS - 1))
}

func (b *BitsetIndex) GetFree() int {
	for j, n := range b {
		if n < ^uint64(0) {
			for bit := uint(j * bitsetWS); bit < uint((j*bitsetWS)+bitsetWS); bit++ {
				if !b.HasBit(bit) {
					return int(bit)
				}
			}
		}
	}
	return -1
}

func (b *BitsetIndex) GetFree2() int {
	for j, n := range b {
		if n < ^uint64(0) {
			for i := 0; i < bitsetWS; i++ {
				if ((n >> i) & 1) == 0 {
					// below is shorthand for: (j * bitsetWS)+i
					return (j << bitsetL2) ^ i
				}
			}
		}
	}
	return -1
}

// BitsetIndexInfo contains information about the index, such as the number of
// pages that are currently in use, the number of trailing pages not in use,
// the next unused page offset, and the percent full number (between 0 and 100)
type BitsetIndexInfo struct {
	PagesInUse            int
	TrailingPagesNotInUse int
	NextUnusedPageOffset  int
	PercentFull           int
}

func (bi *BitsetIndexInfo) String() string {
	return fmt.Sprintf(
		"PagesInUse=%d\nTrailingPagesNotInUse=%d\nNextUnusedPageOffset=%d\nPercentFull=%d\n",
		bi.PagesInUse, bi.TrailingPagesNotInUse, bi.NextUnusedPageOffset, bi.PercentFull,
	)
}

// Info returns a new BitsetIndexInfo struct containing the number of pages that
// are currently in use, the number of trailing pages not in use, the next unused
// page offset, and the percent full number.
func (b *BitsetIndex) Info() *BitsetIndexInfo {
	// First, we will create a new bitset index info struct that we can populate.
	bi := new(BitsetIndexInfo)
	// First, we must get the population count.
	var pop int
	var con []int
	for i := range b {
		pop = mathbits.OnesCount64(b[i])
		// Check the population count on each iteration, so we can find our next
		// page offset location.
		if pop > 0 {
			if bi.PagesInUse == 0 {
				// This is the first section of set bits we have encountered. We
				// should set our next unused page offset here.
				if b[i] < ^uint64(0) {
					for j := 0; j < bitsetWS; j++ {
						if ((b[i] >> j) & 1) == 0 {
							// below is shorthand for: (i * bitsetWS)+j
							bi.NextUnusedPageOffset = (i << bitsetL2) ^ j
							break
						}
					}
				}
			}
			// Otherwise, we simply add to our pages in use number, because we
			// have obviously found more bits that are being used.
			bi.PagesInUse += pop
		}
		if b[i] == 0 {
			// Check if we should update our contiguous, unused page mappings
			if len(con) > 0 && con[len(con)-1] == i-1 {
				bi.TrailingPagesNotInUse += bitsetWS
				con = append(con, i)
				continue
			}
			bi.TrailingPagesNotInUse += bitsetWS
			con = append(con, i)
		}
	}
	bi.PercentFull = bi.PagesInUse / 10
	return bi
}

func (b *BitsetIndex) ReadFile(name string) error {
	// error checking
	if b == nil {
		return io.ErrNoProgress
	}
	// read data from current
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	for i, j := 0, 0; i < len(data) && j < bitsetSize; i, j = i+8, j+1 {
		// decode all the bytes back into the uint64 bitset
		(*b)[j] = binary.LittleEndian.Uint64(data[i : i+8])
	}
	// empty the buffer
	data = nil
	// return nil
	return err
}

func (b *BitsetIndex) WriteFile(name string) error {
	// error checking
	if b == nil {
		return io.ErrNoProgress
	}
	// make new buffer
	data := make([]byte, (bitsetSize*bitsetWS)/8, (bitsetSize*bitsetWS)/8)
	for i, j := 0, 0; i < len(data) && j < bitsetSize; i, j = i+8, j+1 {
		// encode each uint64 into the buffer
		binary.LittleEndian.PutUint64(data[i:i+8], (*b)[j])
	}
	// write buffer to current
	err := os.WriteFile(name, data, 0644)
	if err != nil {
		return err
	}
	// empty the buffer
	data = nil
	// return nil
	return nil
}

// Clear clears all the bits
func (b *BitsetIndex) Clear() {
	for i := range b {
		(*b)[i] = 0
	}
}

// Bits returns the number of bits the bitset index can hold
func (b *BitsetIndex) Bits() int {
	return bitsetSize * bitsetWS
}

func (b *BitsetIndex) String() string {
	resstr := strconv.Itoa(64)
	return fmt.Sprintf("%."+resstr+"b (%d bits)", *b, 64*len(*b))
}

// PageOffsetAfter takes id and returns the next page offset after the page
// offset denoted by id
func (b *BitsetIndex) PageOffsetAfter(id int) int {
	for j, n := range b {
		if n < ^uint64(0) {
			for i := 0; i < bitsetWS; i++ {
				if ((j << bitsetL2) ^ i) < id+1 {
					continue
				}
				if ((n >> i) & 1) == 0 {
					// below is shorthand for: (j * bitsetWS)+i
					return (j << bitsetL2) ^ i
				}
			}
		}
	}
	return -1
}
