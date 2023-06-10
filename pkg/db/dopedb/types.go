package dopedb

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var binaryEncoding = binary.BigEndian

const (
	FixInt   = 0x00 // integer type if <= 127 (0x00 - 0x7f)
	FixMap   = 0x80 // map type containing <= 15 elements (0x80 - 0x8f)
	FixArray = 0x90 // array type containing <= 15 elements (0x90 - 0x9f)
	FixStr   = 0xa0 // string type with <= 31 characters (0xa0 - 0xbf)

	Nil = 0xc0

	Unused = 0xc1

	BoolFalse = 0xc2 // type: store a boolean value set to false
	BoolTrue  = 0xc3 // type: store a boolean value set to true

	Bin8  = 0xc4 // type + 1 byte: byte array where len <= 255 bytes
	Bin16 = 0xc5 // type + 2 bytes: byte array where len <= 65535 bytes
	Bin32 = 0xc6 // type + 4 bytes: byte array where len <= 4294967295 bytes

	Ext8  = 0xc7 // type + 1 byte integer + 1 byte type: byte array <= 255
	Ext16 = 0xc8 // type + 1 byte integer + 1 byte type: byte array <= 65535
	Ext32 = 0xc9 // type + 1 byte integer + 1 byte type: byte array <= 4294967295

	Float32 = 0xca // type + 4 bytes: float32 value
	Float64 = 0xcb // type + 8 bytes: float64 value

	Uint8  = 0xcc // type + 1 byte: uint8 value
	Uint16 = 0xcd // type + 2 bytes: uint16 value
	Uint32 = 0xce // type + 4 bytes: uint32 value
	Uint64 = 0xcf // type + 8 bytes: uint64 value

	Int8  = 0xd0 // type + 1 byte: int8 value
	Int16 = 0xd1 // type + 2 bytes: int16 value
	Int32 = 0xd2 // type + 4 bytes: int32 value
	Int64 = 0xd3 // type + 8 bytes: int64 value

	FixExt1  = 0xd4 // type + 1 byte integer: 1 byte array
	FixExt2  = 0xd5 // type + 1 byte integer: 2 byte array
	FixExt4  = 0xd6 // type + 1 byte integer: 4 byte array
	FixExt8  = 0xd7 // type + 1 byte integer: 8 byte array
	FixExt16 = 0xd8 // type + 1 byte integer: 16 byte array

	Str8  = 0xd9 // type + 1 byte: string where len <= 255 bytes
	Str16 = 0xda // type + 2 bytes: string where len <= 65535 bytes
	Str32 = 0xdb // type + 4 bytes: string where len <= 4294967295 bytes

	Array16 = 0xdc // type + 2 bytes: array containing <= 65535 items
	Array32 = 0xdd // type + 4 bytes: array containing <= 4294967295 items

	Map16 = 0xde // type + 2 bytes: map containing <= 65535 items
	Map32 = 0xdf // type + 4 bytes: map containing <= 4294967295 items

	NegFixInt = 0xe0 // 111xxxxx	0xe0 - 0xff

	bitFix = 0x1f
	bit8   = 0xff
	bit16  = 0xffff
	bit32  = 0xffffffff
)

var (
	ErrWritingBuffer = errors.New("error: no more room to write to buffer")
	ErrReadingBuffer = errors.New("error: nothing left in the buffer to read")
)

func encodingError(s string) error {
	return fmt.Errorf("encoding: there was an issue while encoding [%q]", s)
}

func decodingError(s string) error {
	return fmt.Errorf("decoding: there was an issue while decoding [%q]", s)
}

func hasRoom(p []byte, size int) bool {
	return len(p) >= size
}

/*
type Str string

func (s *Str) Decode(b []byte) error {
	*s = Str(b)
	return nil
}

func (s *Str) Encode() ([]byte, error) {
	return []byte(*s), nil
}

const (
	tFloat = 1 << iota
	tInt
	tUint
	b16
	b32
	b64
)

type Num string

func (n *Num) Decode(b []byte) error {

	if len(b) < 1 {
		goto ret
	}

	// Handle float cases
	if b[0] == tFloat|b32 || b[0] == tFloat|b64 {
		num, err := decodeFloat(b)
		if err != nil {
			return err
		}
		*n = Num(num)
		return nil
	}

	// Handle integer cases
	if b[0] == tInt|b16 || b[0] == tInt|b32 || b[0] == tInt|b64 {
		num, err := decodeInt(b)
		if err != nil {
			return err
		}
		*n = Num(num)
		return nil
	}

	// Handle uint cases
	if b[0] == tUint|b16 || b[0] == tUint|b32 || b[0] == tUint|b64 {
		num, err := decodeUint(b)
		if err != nil {
			return err
		}
		*n = Num(num)
		return nil
	}
ret:
	return fmt.Errorf("%s cannot be decoded as a number", b)
}

// func (n *Num) Encode() ([]byte, error) {
// 	buf := new(bytes.Buffer)
// 	err := binary.Write(buf, binary.BigEndian, n)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

func (n *Num) Encode() ([]byte, error) {
	// Handle float cases
	if strings.IndexByte(string(*n), '.') != -1 {
		return encodeFloat(string(*n))
	}
	// Handle other negative integer cases
	if strings.IndexByte(string(*n), '-') != -1 {
		return encodeInt(string(*n))
	}
	// Otherwise, it is a unsigned integer
	//
	return encodeUint(string(*n))
}

func encodeFloat(s string) ([]byte, error) {
	bitSize := 32
	if len(s) > 46 {
		bitSize = 64
	}
	f, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		return nil, err
	}
	// float32
	if bitSize == 32 {
		buf := make([]byte, 5)
		buf[0] = tFloat | b32
		binary.BigEndian.PutUint32(buf[1:], math.Float32bits(float32(f)))
		return buf, nil
	}
	// float64
	buf := make([]byte, 9)
	buf[0] = tFloat | b64
	binary.BigEndian.PutUint64(buf[1:], math.Float64bits(f))
	return buf, nil
}

func decodeFloat(b []byte) (string, error) {
	if len(b) < 5 {
		return "", fmt.Errorf("%s cannot be decoded as a float\n", b)
	}
	if b[0] == tFloat|b32 {
		f := math.Float32frombits(binary.BigEndian.Uint32(b[1:]))
		return strconv.FormatFloat(float64(f), 'E', -1, 32), nil
	}
	if b[0] == tFloat|b64 {
		f := math.Float64frombits(binary.BigEndian.Uint64(b[1:]))
		return strconv.FormatFloat(f, 'E', -1, 64), nil
	}
	return "", fmt.Errorf("%s cannot be decoded as a float\n", b)
}

func encodeInt(s string) ([]byte, error) {
	// int16
	if "-32768" <= s && s <= "32767" {
		i, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 3)
		buf[0] = tInt | b16
		binary.BigEndian.PutUint16(buf[1:], uint16(i))
		return buf, nil
	}
	// int32
	if "-2147483648" <= s && s <= "2147483647" {
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 5)
		buf[0] = tInt | b32
		binary.BigEndian.PutUint32(buf[1:], uint32(i))
		return buf, nil
	}
	// int64
	if "-9223372036854775808" <= s && s <= "9223372036854775807" {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 9)
		buf[0] = tInt | b64
		binary.BigEndian.PutUint64(buf[1:], uint64(i))
		return buf, nil
	}
	return nil, fmt.Errorf("%s cannot be encoded as an int\n", s)
}

func decodeInt(b []byte) (string, error) {
	if len(b) < 3 {
		return "", fmt.Errorf("%s cannot be decoded as a int\n", b)
	}
	if b[0] == tInt|b16 {
		i := binary.BigEndian.Uint16(b[1:])
		return strconv.FormatInt(int64(i), 10), nil
	}
	if b[0] == tInt|b32 {
		i := binary.BigEndian.Uint32(b[1:])
		return strconv.FormatInt(int64(i), 10), nil
	}
	if b[0] == tInt|b64 {
		i := binary.BigEndian.Uint64(b[1:])
		return strconv.FormatInt(int64(i), 10), nil
	}
	return "", fmt.Errorf("%s cannot be decoded as a int\n", b)
}

func encodeUint(s string) ([]byte, error) {
	// uint16
	if "0" <= s && s <= "65535" {
		u, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 3)
		buf[0] = tUint | b16
		binary.BigEndian.PutUint16(buf[1:], uint16(u))
		return buf, nil
	}
	// uint32
	if "0" <= s && s <= "4294967295" {
		u, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 5)
		buf[0] = tUint | b32
		binary.BigEndian.PutUint32(buf[1:], uint32(u))
		return buf, nil
	}
	// uint64
	if "0" <= s && s <= "18446744073709551615" {
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 9)
		buf[0] = tUint | b64
		binary.BigEndian.PutUint64(buf[1:], u)
		return buf, nil
	}
	return nil, fmt.Errorf("%s cannot be encoded as a uint\n", s)
}

func decodeUint(b []byte) (string, error) {
	if len(b) < 1 {
		return "", fmt.Errorf("%s cannot be decoded as a uint\n", b)
	}
	if b[0] == tUint|b16 {
		u := binary.BigEndian.Uint16(b[1:])
		return strconv.FormatUint(uint64(u), 10), nil
	}
	if b[0] == tUint|b32 {
		u := binary.BigEndian.Uint16(b[1:])
		return strconv.FormatUint(uint64(u), 10), nil
	}
	if b[0] == tUint|b64 {
		u := binary.BigEndian.Uint16(b[1:])
		return strconv.FormatUint(uint64(u), 10), nil
	}
	return "", fmt.Errorf("%s cannot be decoded as a uint\n", b)
}

*/
