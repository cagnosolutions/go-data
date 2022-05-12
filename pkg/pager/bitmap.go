package pager

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

const (
	ws   = 64
	l2ws = 6
	max  = ^uint64(0)
)

type bitmap struct {
	bits   []uint64
	length uint
}

func newBitmap(length uint) *bitmap {
	size := alignedSize(uint64(length))
	return &bitmap{
		bits:   make([]uint64, size),
		length: length,
	}
}

// read attempts to read the contents of p and fill out
// the bitmap to a byte aligned len(p) slice. It returns
// the number of bytes read. Anything that was in the
// bitmap previously will be gone, and it will have been
// resized to fit the contents of p.
func (bm *bitmap) read(p []byte) int {
	// Ensure we are dealing with an offset that is divisible
	// by eight so that we can work effectively work with it.
	n := align(len(p), 8)
	if n > len(p) {
		n -= 8
	}
	// Create a fresh slice of uint64's
	var bits []uint64
	// Start looping and read from p in chunks of eight bytes.
	var v uint64
	var bytesRead int
	for i := 0; i < n; i += 8 {
		// Read p in sections of eight bytes and convert
		// it directly into uint64.
		v = binary.LittleEndian.Uint64(p[i : i+8])
		bits = append(bits, v)
		bytesRead += 8
	}
	// Finally, overwrite the bitmaps bitset
	bm.bits = nil
	bm.bits = bits
	bm.length = uint(len(bits) * 8)
	// Return the number of bytes read
	return bytesRead
}

// write attempts to write the contents of the bitmap to
// a byte aligned len(p) slice. It returns the number of
// bytes written.
func (bm *bitmap) write(p []byte) int {
	// Ensure we are dealing with an offset that is divisible
	// by eight so that we can work effectively work with it.
	n := align(len(p), 8)
	if n > len(p) {
		n -= 8
	}
	fmt.Println(">>>>>> n:", n, "len(p):", len(p))
	// Start looping and writing to p in chunks of eight bytes.
	var j, bytesWritten int
	for i := 0; i < n; i += 8 {
		// Convert the bitmap entry and write it directly to
		// p's address space.
		binary.LittleEndian.PutUint64(p[i:i+8], bm.bits[j])
		j++
		bytesWritten += 8
	}
	// Return the number of bytes written
	return bytesWritten
}

func (bm *bitmap) has(i uint) bool {
	return _has(&bm.bits, i)
}

func (bm *bitmap) set(i uint) {
	_set(&bm.bits, i)
}

func (bm *bitmap) get(i uint) uint64 {
	return _get(&bm.bits, i)
}

func (bm *bitmap) unset(i uint) {
	_unset(&bm.bits, i)
}

func (bm *bitmap) first() int {
	for i := 0; i < int(bm.length); i++ {
		if !_has(&bm.bits, uint(i)) {
			return i
		}
	}
	return -1
}

func (bm *bitmap) String() string {
	resstr := strconv.Itoa(ws)
	return fmt.Sprintf("%."+resstr+"b (%d bits)", bm.bits, ws*len(bm.bits))

}

func _has(bits *[]uint64, i uint) bool {
	// checkResize(bits, i)
	return ((*bits)[i>>l2ws] & (1 << (i & (ws - 1)))) != 0
}

func _set(bits *[]uint64, i uint) {
	// checkResize(bits, i)
	(*bits)[i>>l2ws] |= 1 << (i & (ws - 1))
}

func _get(bits *[]uint64, i uint) uint64 {
	// checkResize(bits, i)
	return (*bits)[i>>l2ws] & (1 << (i & (ws - 1)))
}

func _unset(bits *[]uint64, i uint) {
	// checkResize(bits, i)
	(*bits)[i>>l2ws] &^= 1 << (i & (ws - 1))
}

func align(n int, size int) int {
	mask := size - 1
	return (n + mask) &^ mask
}

func alignedSize(size uint64) uint64 {
	if size > (max - ws + 1) {
		return max >> l2ws
	}
	return (size + (ws - 1)) >> l2ws
}

func roundTo(value uint, roundTo uint) uint {
	return (value + (roundTo - 1)) &^ (roundTo - 1)
}

func checkResize(bits *[]uint64, i uint) {
	if *bits == nil {
		*bits = make([]uint64, 8)
		return
	}
	if i > uint(len(*bits)*8) {
		newbs := make([]uint64, roundTo(i, 8))
		copy(newbs, *bits)
		*bits = newbs
	}
	return
}
