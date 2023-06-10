package dopedb

import (
	"encoding/binary"
)

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
