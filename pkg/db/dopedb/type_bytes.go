package dopedb

import (
	"encoding/binary"
)

func encBin(p []byte) ([]byte, error) {
	n := len(p)
	switch {
	case n <= bit8:
		return encBin8(p)
	case n <= bit16:
		return encBin16(p)
	case n <= bit32:
		return encBin32(p)
	}
	return nil, encodingError("encBin")
}

func encBin8(p []byte) ([]byte, error) {
	n := len(p)
	if len(p) > bit8 {
		return nil, encodingError("encBin8")
	}
	t := Bin8
	b := make([]byte, 1+1+n)
	b[0] = uint8(t)
	b[1] = uint8(n)
	copy(b[2:], p)
	return b, nil
}

func encBin16(p []byte) ([]byte, error) {
	n := len(p)
	if len(p) > bit16 {
		return nil, encodingError("encBin16")
	}
	t := Bin16
	b := make([]byte, 1+2+n)
	b[0] = uint8(t)
	binary.BigEndian.PutUint16(b[1:3], uint16(n))
	copy(b[3:], p)
	return b, nil
}

func encBin32(p []byte) ([]byte, error) {
	n := len(p)
	if len(p) > bit32 {
		return nil, encodingError("encBin32")
	}
	t := Bin32
	b := make([]byte, 1+4+n)
	b[0] = uint8(t)
	binary.BigEndian.PutUint32(b[1:5], uint32(n))
	copy(b[5:], p)
	return b, nil
}

func decBin(b []byte) ([]byte, error) {
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
	return nil, decodingError("decBin")
}

func decBin8(b []byte) ([]byte, error) {
	n := b[1]
	p := make([]byte, n)
	copy(p, b[2:2+n])
	return p, nil
}

func decBin16(b []byte) ([]byte, error) {
	n := binary.BigEndian.Uint16(b[1:3])
	p := make([]byte, n)
	copy(p, b[3:3+n])
	return p, nil
}

func decBin32(b []byte) ([]byte, error) {
	n := binary.BigEndian.Uint32(b[1:5])
	p := make([]byte, n)
	copy(p, b[5:5+n])
	return p, nil
}
