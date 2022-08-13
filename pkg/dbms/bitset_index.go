package dbms

import (
	"encoding/binary"
	"fmt"
	"io"
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

func (b *BitsetIndex) ReadFile(name string) error {
	// error checking
	if b == nil {
		return io.ErrNoProgress
	}
	// read data from file
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
	// write buffer to file
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
