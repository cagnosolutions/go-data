package pager

import (
	"encoding/binary"
)

// getU16 returns an uint16 field value at the provided offset
// reading it from the provided byte slice.
func getU16(d []byte, off int) uint16 {
	return binary.LittleEndian.Uint16(d[off : off+2])
}

// putU16 takes an uint16 field value at the provided offset
// writing it back to the provided byte slice.
func putU16(d []byte, off int, val uint16) {
	binary.LittleEndian.PutUint16(d[off:off+2], val)
}

// incrU16 first reads the uint16 value at the provided offset
// then increments it by the amount indicated with the provided
// by declaration and then writes the incremented uint16 value
// back to the underlying byte slice.
func incrU16(d []byte, off int, by uint16) {
	var val uint16
	val = binary.LittleEndian.Uint16(d[off : off+2])
	val += by
	binary.LittleEndian.PutUint16(d[off:off+2], val)
}

// decrU16 first reads the uint16 value at the provided offset
// then decrements it by the amount indicated with the provided
// by declaration and then writes the decremented uint16 value
// back to the underlying byte slice.
func decrU16(d []byte, off int, by uint16) {
	var val uint16
	val = binary.LittleEndian.Uint16(d[off : off+2])
	val -= by
	binary.LittleEndian.PutUint16(d[off:off+2], val)
}

// getU32 returns an uint32 field value at the provided offset
// reading it from the provided byte slice.
func getU32(d []byte, off int) uint32 {
	return binary.LittleEndian.Uint32(d[off : off+4])
}

// putU32 takes an uint32 field value at the provided offset
// writing it back to the provided byte slice.
func putU32(d []byte, off int, val uint32) {
	binary.LittleEndian.PutUint32(d[off:off+4], val)
}

// incrU32 first reads the uint32 value at the provided offset
// then increments it by the amount indicated with the provided
// by declaration and then writes the incremented uint16 value
// back to the underlying byte slice.
func incrU32(d []byte, off int, by uint32) {
	var val uint32
	val = binary.LittleEndian.Uint32(d[off : off+4])
	val += by
	binary.LittleEndian.PutUint32(d[off:off+4], val)
}

// decrU32 first reads the uint32 value at the provided offset
// then decrements it by the amount indicated with the provided
// by declaration and then writes the decremented uint16 value
// back to the underlying byte slice.
func decrU32(d []byte, off int, by uint32) {
	var val uint32
	val = binary.LittleEndian.Uint32(d[off : off+4])
	val -= by
	binary.LittleEndian.PutUint32(d[off:off+4], val)
}

// diffU16 first reads the uint16 values at the provides offsets
// then subtracts the value of off2 from the value of off1 and
// returns the difference.
func diffU16(d []byte, off1, off2 int) uint16 {
	var val1, val2 uint16
	val1 = binary.LittleEndian.Uint16(d[off1 : off1+2])
	val2 = binary.LittleEndian.Uint16(d[off2 : off2+2])
	return val1 - val2
}
