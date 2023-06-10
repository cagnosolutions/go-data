package dopedb

import (
	"encoding/binary"
)

func encStr(s string) ([]byte, error) {
	n := len(s)
	switch {
	case n <= bitFix:
		return encFixStr(s)
	case n <= bit8:
		return encStr8(s)
	case n <= bit16:
		return encStr16(s)
	case n <= bit32:
		return encStr32(s)
	}
	return nil, encodingError("encStr")
}

func encFixStr(s string) ([]byte, error) {
	n := len(s)
	if len(s) > bitFix {
		return nil, encodingError("makeFixStr")
	}
	t := byte(FixStr | n)
	b := make([]byte, 1+n)
	b[0] = t
	copy(b[1:], s)
	return b, nil
}

func encStr8(s string) ([]byte, error) {
	n := len(s)
	if len(s) > bit8 {
		return nil, encodingError("encStr8")
	}
	t := Str8
	b := make([]byte, 1+1+n)
	b[0] = uint8(t)
	b[1] = uint8(n)
	copy(b[2:], s)
	return b, nil
}

func encStr16(s string) ([]byte, error) {
	n := len(s)
	if len(s) > bit16 {
		return nil, encodingError("encStr16")
	}
	t := Str16
	b := make([]byte, 1+2+n)
	b[0] = uint8(t)
	binary.BigEndian.PutUint16(b[1:3], uint16(n))
	copy(b[3:], s)
	return b, nil
}

func encStr32(s string) ([]byte, error) {
	n := len(s)
	if len(s) > bit32 {
		return nil, encodingError("encStr32")
	}
	t := Str32
	b := make([]byte, 1+4+n)
	b[0] = uint8(t)
	binary.BigEndian.PutUint32(b[1:5], uint32(n))
	copy(b[5:], s)
	return b, nil
}

func decStr(b []byte) (string, error) {
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
	return "", decodingError("decStr")
}

func decFixStr(b []byte) (string, error) {
	n := b[0] &^ FixStr
	return string(b[1 : 1+n]), nil
}

func decStr8(b []byte) (string, error) {
	n := b[1]
	return string(b[2 : 2+n]), nil
}

func decStr16(b []byte) (string, error) {
	n := binary.BigEndian.Uint16(b[1:3])
	return string(b[3 : 3+n]), nil
}

func decStr32(b []byte) (string, error) {
	n := binary.BigEndian.Uint32(b[1:5])
	return string(b[5 : 5+n]), nil
}
