package dopedb

import (
	"encoding/binary"
	"strings"
)

var Strings stringTypes

type stringTypes struct{}

func (s stringTypes) EncStr(p []byte, v string) int {
	if len(v) <= bitFix {
		s.EncFixStr(p, v)
		return FixStr
	}
	if len(v) <= bit8 {
		s.EncStr8(p, v)
		return Str8
	}
	if len(v) <= bit16 {
		s.EncStr16(p, v)
		return Str16
	}
	if len(v) <= bit32 {
		s.EncStr32(p, v)
		return Str32
	}
	panic("cannot encode, type does not match expected encoding")
}

func (s stringTypes) EncFixStr(p []byte, v string) {
	_ = p[len(v)+1]      // early bounds check to guarantee safety of writes below
	if len(v) > bitFix { // bitFix = 31
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = byte(FixStr | len(v))
	copy(p[1:], v)
}

func (s stringTypes) EncStr8(p []byte, v string) {
	_ = p[len(v)+2]    // early bounds check to guarantee safety of writes below
	if len(v) > bit8 { // bit8 = 255
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = Str8
	p[1] = byte(len(v))
	copy(p[2:], v)
}

func (s stringTypes) EncStr16(p []byte, v string) {
	_ = p[len(v)+3]     // early bounds check to guarantee safety of writes below
	if len(v) > bit16 { // bit16 = 65535
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = Str16
	binary.BigEndian.PutUint16(p[1:3], uint16(len(v)))
	copy(p[3:], v)
}

func (s stringTypes) EncStr32(p []byte, v string) {
	_ = p[len(v)+5]     // early bounds check to guarantee safety of writes below
	if len(v) > bit32 { // bit32 = 4294967295
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = Str32
	binary.BigEndian.PutUint32(p[1:5], uint32(len(v)))
	copy(p[5:], v)
}

func (s stringTypes) EncBin8(p []byte, v []byte) {
	_ = p[len(v)+2]    // early bounds check to guarantee safety of writes below
	if len(v) > bit8 { // bit8 = 255
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = Bin8
	p[1] = byte(len(v))
	copy(p[2:], v)
}

func (s stringTypes) EncBin16(p []byte, v []byte) {
	_ = p[len(v)+3]     // early bounds check to guarantee safety of writes below
	if len(v) > bit16 { // bit16 = 65535
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = Bin16
	binary.BigEndian.PutUint16(p[1:3], uint16(len(v)))
	copy(p[3:], v)
}

func (s stringTypes) EncBin32(p []byte, v []byte) {
	_ = p[len(v)+5]     // early bounds check to guarantee safety of writes below
	if len(v) > bit32 { // bit32 = 4294967295
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = Bin32
	binary.BigEndian.PutUint32(p[1:5], uint32(len(v)))
	copy(p[5:], v)
}

func (s stringTypes) DecStr(p []byte) string {
	_ = p[0] // bounds check hint to compiler
	if p[0]&FixStr == FixStr {
		return s.DecFixStr(p)
	}
	if p[0] == Str8 {
		return s.DecStr8(p)
	}
	if p[0] == Str16 {
		return s.DecStr16(p)
	}
	if p[0] == Str32 {
		return s.DecStr32(p)
	}
	panic("cannot decode, type does not match expected type")
}

func (s stringTypes) DecFixStr(p []byte) string {
	_ = p[0] // bounds check hint to compiler
	n := int(p[0] &^ FixStr)
	_ = p[n+1] // bounds check hint to compiler
	var sb strings.Builder
	sb.Grow(n)
	sb.Write(p[1 : 1+n])
	return sb.String()
}

func (s stringTypes) DecStr8(p []byte) string {
	_ = p[2] // bounds check hint to compiler
	if p[0] != Str8 {
		panic("cannot decode, type does not match expected type")
	}
	n := int(p[1])
	_ = p[n] // bounds check hint to compiler
	var sb strings.Builder
	sb.Grow(n)
	sb.Write(p[2 : 2+n])
	return sb.String()
}

func (s stringTypes) DecStr16(p []byte) string {
	_ = p[3] // bounds check hint to compiler
	if p[0] != Str16 {
		panic("cannot decode, type does not match expected type")
	}
	n := int(binary.BigEndian.Uint16(p[1:3]))
	_ = p[n] // bounds check hint to compiler
	var sb strings.Builder
	sb.Grow(n)
	sb.Write(p[3 : 3+n])
	return sb.String()
}

func (s stringTypes) DecStr32(p []byte) string {
	_ = p[5] // bounds check hint to compiler
	if p[0] != Str32 {
		panic("cannot decode, type does not match expected type")
	}
	n := int(binary.BigEndian.Uint32(p[1:5]))
	_ = p[n] // bounds check hint to compiler
	var sb strings.Builder
	sb.Grow(n)
	sb.Write(p[5 : 5+n])
	return sb.String()
}

func (s stringTypes) DecBin8(p []byte) []byte {
	_ = p[2] // bounds check hint to compiler
	if p[0] != Bin8 {
		panic("cannot decode, type does not match expected type")
	}
	n := int(p[1])
	_ = p[n] // bounds check hint to compiler
	b := make([]byte, n)
	copy(b, p[2:2+n])
	return b
}

func (s stringTypes) DecBin16(p []byte) []byte {
	_ = p[3] // bounds check hint to compiler
	if p[0] != Bin16 {
		panic("cannot decode, type does not match expected type")
	}
	n := int(binary.BigEndian.Uint16(p[1:3]))
	_ = p[n] // bounds check hint to compiler
	b := make([]byte, n)
	copy(b, p[3:3+n])
	return b
}

func (s stringTypes) DecBin32(p []byte) []byte {
	_ = p[5] // bounds check hint to compiler
	if p[0] != Bin32 {
		panic("cannot decode, type does not match expected type")
	}
	n := int(binary.BigEndian.Uint32(p[1:5]))
	_ = p[n] // bounds check hint to compiler
	b := make([]byte, n)
	copy(b, p[5:5+n])
	return b
}

func encStr(p []byte, s string) {
	n := len(s)
	switch {
	case n <= bitFix:
		encFixStr(p, s)
		return
	case n <= bit8:
		encStr8(p, s)
		return
	case n <= bit16:
		encStr16(p, s)
		return
	case n <= bit32:
		encStr32(p, s)
		return
	}
	panic(encodingError("encStr"))
}

func encFixStr(p []byte, s string) {
	n := len(s)
	if len(s) > bitFix {
		panic(encodingError("makeFixStr"))
	}
	if !hasRoom(p, n+1) {
		panic(ErrWritingBuffer)
	}
	p[0] = byte(FixStr | n)
	copy(p[1:], s)
}

func encStr8(p []byte, s string) {
	n := len(s)
	if len(s) > bit8 {
		panic(encodingError("encStr8"))
	}
	if !hasRoom(p, n+2) {
		panic(ErrWritingBuffer)
	}
	p[0] = Str8
	p[1] = uint8(n)
	copy(p[2:], s)
}

func encStr16(p []byte, s string) {
	n := len(s)
	if len(s) > bit16 {
		panic(encodingError("encStr16"))
	}
	if !hasRoom(p, n+3) {
		panic(ErrWritingBuffer)
	}
	p[0] = Str16
	binary.BigEndian.PutUint16(p[1:3], uint16(n))
	copy(p[3:], s)
}

func encStr32(p []byte, s string) {
	n := len(s)
	if len(s) > bit32 {
		panic(encodingError("encStr32"))
	}
	if !hasRoom(p, n+5) {
		panic(ErrWritingBuffer)
	}
	p[0] = Str32
	binary.BigEndian.PutUint32(p[1:5], uint32(n))
	copy(p[5:], s)
}

func decStr(b []byte) string {
	// static check
	_ = b[1]
	b1 := b[0]
	switch {
	case b1&FixStr == FixStr:
		return decFixStr(b)
	case b1 == Str8:
		return decStr8(b)
	case b1 == Str16:
		return decStr16(b)
	case b1 == Str32:
		return decStr32(b)
	}
	panic(decodingError("decStr"))
}

func decFixStr(p []byte) string {
	// static check
	_ = p[0]

	n := p[0] &^ FixStr
	if !hasRoom(p, int(n)+1) {
		panic(ErrReadingBuffer)
	}
	return string(p[1 : 1+n])
}

func decStr8(p []byte) string {
	// static check
	_ = p[1]

	n := p[1]
	if !hasRoom(p, int(n+2)) {
		panic(ErrReadingBuffer)
	}
	return string(p[2 : 2+n])
}

func decStr16(p []byte) string {
	// static check
	_ = p[3]

	n := binary.BigEndian.Uint16(p[1:3])
	if !hasRoom(p, int(n+3)) {
		panic(ErrReadingBuffer)
	}
	return string(p[3 : 3+n])
}

func decStr32(p []byte) string {
	// static check
	_ = p[5]

	n := binary.BigEndian.Uint32(p[1:5])
	if !hasRoom(p, int(n+5)) {
		panic(ErrReadingBuffer)
	}
	return string(p[5 : 5+n])
}
