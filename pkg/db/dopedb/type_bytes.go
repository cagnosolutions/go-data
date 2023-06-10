package dopedb

import (
	"encoding/binary"
)

func encBin(p []byte, data []byte) {
	n := len(p)
	switch {
	case n <= bit8:
		encBin8(p, data)
		return
	case n <= bit16:
		encBin16(p, data)
		return
	case n <= bit32:
		encBin32(p, data)
		return
	}
	panic(encodingError("encBin"))
}

func encBin8(p []byte, data []byte) {
	n := len(p)
	if len(p) > bit8 {
		panic(encodingError("encBin8"))
	}
	if !hasRoom(p, n+2) {
		panic(ErrWritingBuffer)
	}
	p[0] = uint8(Bin8)
	p[1] = uint8(n)
	copy(p[2:], data)
}

func encBin16(p []byte, data []byte) {
	n := len(p)
	if len(p) > bit16 {
		panic(encodingError("encBin16"))
	}
	if !hasRoom(p, n+3) {
		panic(ErrWritingBuffer)
	}
	p[0] = uint8(Bin16)
	binary.BigEndian.PutUint16(p[1:3], uint16(n))
	copy(p[3:], data)
}

func encBin32(p []byte, data []byte) {
	n := len(p)
	if len(p) > bit32 {
		panic(encodingError("encBin32"))
	}
	if !hasRoom(p, n+5) {
		panic(ErrWritingBuffer)
	}
	p[0] = uint8(Bin32)
	binary.BigEndian.PutUint32(p[1:5], uint32(n))
	copy(p[5:], data)
}

func decBin(b []byte) []byte {
	// static check
	_ = b[1]
	switch b[0] {
	case Bin8:
		return decBin8(b)
	case Bin16:
		return decBin16(b)
	case Bin32:
		return decBin32(b)
	}
	panic(decodingError("decBin"))
}

func decBin8(p []byte) []byte {
	// static check
	_ = p[1]

	n := p[1]
	if !hasRoom(p, int(n+1)) {
		panic(ErrReadingBuffer)
	}
	return p[2 : 2+n]
}

func decBin16(p []byte) []byte {
	// static check
	_ = p[3]

	n := binary.BigEndian.Uint16(p[1:3])
	if !hasRoom(p, int(n+3)) {
		panic(ErrReadingBuffer)
	}
	return p[3 : 3+n]
}

func decBin32(p []byte) []byte {
	// static check
	_ = p[5]

	n := binary.BigEndian.Uint32(p[1:5])
	if !hasRoom(p, int(n+5)) {
		panic(ErrReadingBuffer)
	}
	return p[5 : 5+n]
}
