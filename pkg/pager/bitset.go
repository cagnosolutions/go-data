package pager

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

const (
	ws   = 64         // ws is the word size (how many bits can fit into uint64)
	l2ws = 6          // l2ws is the log2 of the word size (6 for 64, 5 for 32, 4 for 16, and 3 for 8)
	max  = ^uint64(0) // max is the maximum integer that can be stored in a word
)

// bitset is a bitset data structure
type bitset struct {
	bits   []uint64 // 24+(len(bits)*8)
	length uint64   // 8
}

// alignedSize aligns a given size to the word size, so it always works well.
func alignedSize(size uint64) uint64 {
	if size > (max - ws + 1) {
		return max >> l2ws
	}
	return (size + (ws - 1)) >> l2ws
}

// newBitset instantiates a new bitset instance. The length parameter
// is the number of bits you wish to store in the set.
func newBitset(length uint) *bitset {
	size := alignedSize(uint64(length))
	return &bitset{
		bits:   make([]uint64, size),
		length: uint64(length),
	}
}

// resize adds additional words to incorporate new bits if necessary.
func (bm *bitset) resize(i uint64) {
	if i < bm.length || i > max {
		return
	}
	nsize := int(alignedSize(i + 1))
	if bm.bits == nil {
		bm.bits = make([]uint64, nsize)
	} else if cap(bm.bits) >= nsize {
		bm.bits = bm.bits[:nsize] // fast resize
	} else if len(bm.bits) < nsize {
		newset := make([]uint64, nsize, 2*nsize) // increase capacity 2x
		copy(newset, bm.bits)
		bm.bits = newset
	}
	bm.length = i + 1
}

// has reads the bit at the provided index and returns a boolean
// value reporting true if the bit is set or on, and false if the
// bit is unset, or off.
func (bm *bitset) has(i uint) bool {
	return ((bm.bits)[i>>l2ws] & (1 << (i & (ws - 1)))) != 0
}

// set flips the bit at the provided index to "on", thus setting
// said bit if it is currently "off", or unset.
func (bm *bitset) set(i uint) {
	bm.resize(uint64(i))
	(bm.bits)[i>>l2ws] |= 1 << (i & (ws - 1))
}

// get reads the bit at the provided index and returns the value.
func (bm *bitset) get(i uint) uint64 {
	return (bm.bits)[i>>l2ws] & (1 << (i & (ws - 1)))
}

// unset flips the bit at the provided index to "off", thus unsetting
// said bit if it is currently "on", or set.
func (bm *bitset) unset(i uint) {
	(bm.bits)[i>>l2ws] &^= 1 << (i & (ws - 1))
}

// free returns the index of the first "free" (unset) bit it can find
func (bm *bitset) free() int {
	for i := 0; i < int(bm.length); i++ {
		if !bm.has(uint(i)) {
			return i
		}
	}
	return -1
}

// roundTo rounds n to binary powers of pow word size, kind of.
func roundTo(n int, pow int) int {
	mask := pow - 1
	return (n + (mask)) &^ (mask)
}

// read attempts to read the contents of p and fill out
// the bitset to a byte aligned len(p) slice. It returns
// the number of bytes read. Anything that was in the
// bitset previously will be gone, and it will have been
// resized to fit the contents of p.
func (bm *bitset) read(p []byte) int {
	// Ensure we are dealing with an offset that is divisible
	// by eight so that we can work effectively work with it.
	n := roundTo(len(p), 8)
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
	bm.length = uint64(len(bits) * 8)
	// Return the number of bytes read
	return bytesRead
}

// write attempts to write the contents of the bitset to
// a byte aligned len(p) slice. It returns the number of
// bytes written.
func (bm *bitset) write(p []byte) int {
	// Ensure we are dealing with an offset that is divisible
	// by eight so that we can work effectively work with it.
	n := roundTo(len(p), 8)
	if n > len(p) {
		n -= 8
	}
	// Start looping and writing to p in chunks of eight bytes.
	var j, bytesWritten int
	for i := 0; i < n; i += 8 {
		// Convert the bitset entry and write it directly to
		// p's address space.
		binary.LittleEndian.PutUint64(p[i:i+8], bm.bits[j])
		j++
		bytesWritten += 8
	}
	// Return the number of bytes written
	return bytesWritten
}

// sizeof returns the actual size the bitset is occupying in memory.
func (bm *bitset) sizeof() int64 {
	sz := int64((24 + 8) + (len(bm.bits) * 8))
	return sz
}

func (bm *bitset) String() string {
	resstr := strconv.Itoa(ws)
	return fmt.Sprintf("%."+resstr+"b (%d bits)", bm.bits, ws*len(bm.bits))
}

func _has(bits *[]uint64, i uint) bool {
	return ((*bits)[i>>l2ws] & (1 << (i & (ws - 1)))) != 0
}

func _set(bits *[]uint64, i uint) {
	(*bits)[i>>l2ws] |= 1 << (i & (ws - 1))
}

func _get(bits *[]uint64, i uint) uint64 {
	return (*bits)[i>>l2ws] & (1 << (i & (ws - 1)))
}

func _unset(bits *[]uint64, i uint) {
	(*bits)[i>>l2ws] &^= 1 << (i & (ws - 1))
}

func makeFlags(count uint, start uint) []uint {
	flags := make([]uint, count)
	for i := start; i < start+count; i++ {
		flags[i] = (start + 1) << i
	}
	return flags
}

func setFlag(flags *uint16, flag uint16) {
	*flags |= flag
}

func unsetFlag(flags *uint16, flag uint16) {
	*flags &= ^flag
}

func flipFlag(flags *uint16, flag uint16) {
	*flags ^= flag
}

func checkFlag(flags uint16, flag uint16) bool {
	return (flags & flag) > 0
}

func checkFlags(flags uint16, args ...uint16) bool {
	for _, arg := range args {
		if (flags & arg) < 1 {
			return false
		}
	}
	return true
}

var (
	setFlagM = func(n, f uint) uint {
		n |= f
		return n
	}
	clrFlagM = func(n, f uint) uint {
		n &= ^f
		return n
	}
	tglFlagM = func(n, f uint) uint {
		n ^= f
		return n
	}
	chkFlagM = func(n, f uint) bool {
		return (n & f) > 0
	}
)
