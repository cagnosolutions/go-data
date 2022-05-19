package pager

// This package provides mapping functions to work with slices of uint
// types. It allows the getting, setting and transforming of various
// uint types such as uint, uint16, uint32 and uint64. Right now, this
// set of functions uses the LittleEndian encoding format, but future
// versions of this will have support for both the LittleEndian and
// BitEndian byte order.

// BinMapU16Fn is the mapping function signature for an uint16
type BinMapU16Fn = func(n uint16) *uint16

// IncrU16 returns a mapping function that increments the underlying
// uint16 using the amount provided by `by`, and returns the mapping
// indicating that the mapping changes should be encoded to the
// underlying slice (because it does not return a nil value.)
func IncrU16(by uint16) BinMapU16Fn {
	return func(n uint16) *uint16 {
		n += by
		return &n
	}
}

// DecrU16 returns a mapping function that decrements the underlying
// uint16 using the amount provided by `by`, and returns the mapping
// indicating that the mapping changes should be encoded to the
// underlying slice (because it does not return a nil value.)
func DecrU16(by uint16) BinMapU16Fn {
	return func(n uint16) *uint16 {
		n += by
		return &n
	}
}

// SetU16 returns a mapping function that sets the underlying
// uint16 using the amount provided in `to`, and returns the mapping
// indicating that the mapping changes should be encoded to the
// underlying slice (because it does not return a nil value.)
func SetU16(to uint16) BinMapU16Fn {
	return func(n uint16) *uint16 {
		n = to
		return &n
	}
}

// GetU16 returns a mapping function that returns the underlying
// uint16 into the return value provided in `ret`. It returns a nil
// value to in the mapping function indicating that there should be
// no changes persisted to the underlying slice.
func GetU16(ret *uint16) BinMapU16Fn {
	return func(n uint16) *uint16 {
		ret = &n
		return nil
	}
}

// BinMapU16 is a binary mapping function that transforms the underlying
// uint16 span of slice b according to the mapping function mapping. If
// the mapping function returns nil, the mapping is not written to the
// underlying slice b. If b is too small this function will panic.
func BinMapU16(b []byte, mapping BinMapU16Fn) {
	_ = b[1] // early bounds check
	n := uint16(b[0]) | uint16(b[1])<<8
	if nn := mapping(n); nn != nil {
		b[0] = byte(*nn)
		b[1] = byte(*nn >> 8)
	}
}

// BinMapU32Fn is the mapping function signature for an uint32
type BinMapU32Fn = func(n uint32) *uint32

// SetU32 returns a mapping function that sets the underlying
// uint16 using the amount provided in `to`, and returns the mapping
// indicating that the mapping changes should be encoded to the
// underlying slice (because it does not return a nil value.)
func SetU32(to uint32) BinMapU32Fn {
	return func(n uint32) *uint32 {
		n = to
		return &n
	}
}

// GetU32 returns a mapping function that returns the underlying
// uint16 into the return value provided in `ret`. It returns a nil
// value to in the mapping function indicating that there should be
// no changes persisted to the underlying slice.
func GetU32(ret *uint32) BinMapU32Fn {
	return func(n uint32) *uint32 {
		ret = &n
		return nil
	}
}

// BinMapU32 is a binary mapping function that transforms the underlying
// uint32 span of slice b according to the mapping function mapping. If
// the mapping function returns nil, the mapping is not written to the
// underlying slice b. If b is too small this function will panic.
func BinMapU32(b []byte, fn BinMapU32Fn) {
	_ = b[3] // early bounds check
	var n uint32
	n = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	if nn := fn(n); nn != nil {
		b[0] = byte(*nn)
		b[1] = byte(*nn >> 8)
		b[2] = byte(*nn >> 16)
		b[3] = byte(*nn >> 24)
	}
}

// BinMapU64Fn is the mapping function signature for an uint64
type BinMapU64Fn = func(n uint64) *uint64

// BinMapU64 is a binary mapping function that transforms the underlying
// uint64 span of slice b according to the mapping function mapping. If
// the mapping function returns nil, the mapping is not written to the
// underlying slice b. If b is too small this function will panic.
func BinMapU64(b []byte, fn BinMapU64Fn) {
	_ = b[7] // early bounds check
	var n uint64
	n = uint64(b[0])
	n |= uint64(b[1]) << 8
	n |= uint64(b[2]) << 16
	n |= uint64(b[3]) << 24
	n |= uint64(b[4]) << 32
	n |= uint64(b[5]) << 40
	n |= uint64(b[6]) << 48
	n |= uint64(b[7]) << 56
	if nn := fn(n); nn != nil {
		b[0] = byte(*nn)
		b[1] = byte(*nn >> 8)
		b[2] = byte(*nn >> 16)
		b[3] = byte(*nn >> 24)
		b[4] = byte(*nn >> 32)
		b[5] = byte(*nn >> 40)
		b[6] = byte(*nn >> 48)
		b[7] = byte(*nn >> 56)
	}
}
