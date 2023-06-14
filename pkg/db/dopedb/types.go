package dopedb

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type Type = uint8

const (
	fixTypeStart Type = FixInt

	FixInt      Type = 0x00 // integer type if <= 127 (0x00 - 0x7f)
	FixIntMax   Type = 0x7f
	FixMap      Type = 0x80 // map type containing <= 15 elements (0x80 - 0x8f)
	FixMapMax   Type = 0x8f
	FixArray    Type = 0x90 // array type containing <= 15 elements (0x90 - 0x9f)
	FixArrayMax Type = 0x9f
	FixStr      Type = 0xa0 // string type with <= 31 characters (0xa0 - 0xbf)
	FixStrMax   Type = 0xbf

	fixTypeEnd Type = FixStrMax

	Nil Type = 0xc0

	Unused Type = 0xc1

	BoolFalse Type = 0xc2 // type: store a boolean value set to false
	BoolTrue  Type = 0xc3 // type: store a boolean value set to true

	Bin8  Type = 0xc4 // type + 1 byte: byte array where len <= 255 bytes
	Bin16 Type = 0xc5 // type + 2 bytes: byte array where len <= 65535 bytes
	Bin32 Type = 0xc6 // type + 4 bytes: byte array where len <= 4294967295 bytes

	Ext8  Type = 0xc7 // type + 1 byte integer + 1 byte type: byte array <= 255
	Ext16 Type = 0xc8 // type + 1 byte integer + 1 byte type: byte array <= 65535
	Ext32 Type = 0xc9 // type + 1 byte integer + 1 byte type: byte array <= 4294967295

	Float32 Type = 0xca // type + 4 bytes: float32 value
	Float64 Type = 0xcb // type + 8 bytes: float64 value

	Uint8  Type = 0xcc // type + 1 byte: uint8 value
	Uint16 Type = 0xcd // type + 2 bytes: uint16 value
	Uint32 Type = 0xce // type + 4 bytes: uint32 value
	Uint64 Type = 0xcf // type + 8 bytes: uint64 value

	Int8  Type = 0xd0 // type + 1 byte: int8 value
	Int16 Type = 0xd1 // type + 2 bytes: int16 value
	Int32 Type = 0xd2 // type + 4 bytes: int32 value
	Int64 Type = 0xd3 // type + 8 bytes: int64 value

	FixExt1  Type = 0xd4 // type + 1 byte integer: 1 byte array
	FixExt2  Type = 0xd5 // type + 1 byte integer: 2 byte array
	FixExt4  Type = 0xd6 // type + 1 byte integer: 4 byte array
	FixExt8  Type = 0xd7 // type + 1 byte integer: 8 byte array
	FixExt16 Type = 0xd8 // type + 1 byte integer: 16 byte array

	Str8  Type = 0xd9 // type + 1 byte: string where len <= 255 bytes
	Str16 Type = 0xda // type + 2 bytes: string where len <= 65535 bytes
	Str32 Type = 0xdb // type + 4 bytes: string where len <= 4294967295 bytes

	Array16 Type = 0xdc // type + 2 bytes: array containing <= 65535 items
	Array32 Type = 0xdd // type + 4 bytes: array containing <= 4294967295 items

	Time32 Type = 0xd6
	Time64 Type = 0xd7

	Map16 Type = 0xde // type + 2 bytes: map containing <= 65535 items
	Map32 Type = 0xdf // type + 4 bytes: map containing <= 4294967295 items

	NegFixInt Type = 0xe0 // 111xxxxx	0xe0 - 0xff

	bitFix  = 0x1f
	bit8    = 0xff
	bit16   = 0xffff
	bit32   = 0xffffffff
	intSize = 32 << (^uint(0) >> 63)

	// left bits
	leftBits1 = 0x80
	leftBits2 = 0xc0
	leftBits3 = 0xe0
	leftBits4 = 0xf0
	leftBits5 = 0xf8
	leftBits6 = 0xfc
	leftBits7 = 0xfe

	// right bits
	rightBits1 = 0x01
	rightBits2 = 0x03
	rightBits3 = 0x07
	rightBits4 = 0x0f
	rightBits5 = 0x1f
	rightBits6 = 0x3f
	rightBits7 = 0x7f
)

var typeToString = map[Type]string{
	FixInt:    "fix int",
	FixMap:    "fix map",
	FixArray:  "fix array",
	FixStr:    "fix string",
	Nil:       "nil",
	Unused:    "unused",
	BoolFalse: "bool (false)",
	BoolTrue:  "bool (true)",
	Bin8:      "bin8",
	Bin16:     "bin16",
	Bin32:     "bin32",
	Ext8:      "ext8",
	Ext16:     "ext16",
	Ext32:     "ext32",
	Float32:   "float32",
	Float64:   "float64",
	Uint8:     "uint8",
	Uint16:    "uint16",
	Uint32:    "uint32",
	Uint64:    "uint64",
	Int8:      "int8",
	Int16:     "int16",
	Int32:     "int32",
	Int64:     "int64",
	FixExt1:   "fix ext1",
	FixExt2:   "fix ext2",
	FixExt4:   "fix ext4",
	FixExt8:   "fix ext8",
	FixExt16:  "fix ext16",
	Str8:      "str8",
	Str16:     "str16",
	Str32:     "str32",
	Array16:   "array16",
	Array32:   "array32",
	Map16:     "map16",
	Map32:     "map32",
	NegFixInt: "neg fix int",
}

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

type IEncoder interface {
	check(n int)
	encodeValue(v any) error
	Encode(v any) error
	writeByte(v byte)
	writeBytes(v []byte)
	write1(t Type, v uint8)
	write2(t Type, v uint8)
	write3(t Type, v uint16)
	write5(t Type, v uint32)
	write9(t Type, v uint64)
	writeNil()
	writeBool(v bool)
	writeFloat32(v float32)
	writeFloat64(v float64)
	writeUint8(v uint8)
	writeUint16(v uint16)
	writeUint32(v uint32)
	writeUint64(v uint64)
	writeFixInt(v int)
	writeInt8(v int8)
	writeInt16(v int16)
	writeInt32(v int32)
	writeInt64(v int64)
	writeStr(v string)
	writeFixStr(v string)
	writeStr8(v string)
	writeStr16(v string)
	writeStr32(v string)
	writeBin8(v []byte)
	writeBin16(v []byte)
	writeBin32(v []byte)
	writeFixArray(v []any)
	writeArray16(v []any)
	writeArray32(v []any)
	writeFixMap(m map[string]any)
	writeMap16(m map[string]any)
	writeMap32(m map[string]any)
	writeFixExt1(t uint8, d byte)
	writeFixExt2(t uint8, d []byte)
	writeFixExt4(t uint8, d []byte)
	writeFixExt8(t uint8, d []byte)
	writeExt8(t uint8, d []byte)
	writeExt16(t uint8, d []byte)
	writeExt32(t uint8, d []byte)
	writeExt(t uint8, d []byte)
	writeTime32(t time.Time)
	writeTime64(t time.Time)
}

type IDecoder interface {
	checkRead(n int)
	decodeValue(v any) error
	Decode(v any) (any, error)
	readByte() byte
	readBytes() []byte
	read1() (Type, uint8)
	read2() (Type, uint8)
	read3() (Type, uint16)
	read5() (Type, uint32)
	read9() (Type, uint64)
	readNil() any
	readBool() bool
	readFloat32() float32
	readFloat64() float64
	readUint8() uint8
	readUint16() uint16
	readUint32() uint32
	readUint64() uint64
	readFixInt() int
	readInt8() int8
	readInt16() int16
	readInt32() int32
	readInt64() int64
	readStr() string
	readFixStr() string
	readStr8() string
	readStr16() string
	readStr32() string
	readBin8() []byte
	readBin16() []byte
	readBin32() []byte
	readFixArray() []any
	readArray16() []any
	readArray32() []any
	readFixMap() map[string]any
	readMap16() map[string]any
	readMap32() map[string]any
	readFixExt1() (uint8, byte)
	readFixExt2() (uint8, []byte)
	readFixExt4() (uint8, []byte)
	readFixExt8() (uint8, []byte)
	readExt8() (uint8, []byte)
	readExt16() (uint8, []byte)
	readExt32() (uint8, []byte)
	readExt() (uint8, []byte)
	readTime32() time.Time
	readTime64() time.Time
}

func hasRoom(p []byte, size int) bool {
	return len(p) >= size
}

func getType(v any) Type {
	switch v.(type) {
	case bool:
		if v == true {
			return BoolTrue
		}
		return BoolFalse
	case nil:
		return Nil
	case float32:
		return Float32
	case float64:
		return Float64
	case uint:
		if intSize == 32 {
			return Uint32
		}
		return Uint64
	case uint8:
		return Uint8
	case uint16:
		return Uint16
	case uint32:
		return Uint32
	case uint64:
		return Uint64
	case int:
		if intSize == 32 {
			return Int32
		}
		return Int64
	case int8:
		return Int8
	case int16:
		return Int16
	case int32:
		return Int32
	case int64:
		return Int64
	case string:
		n := len(v.(string))
		switch {
		case n <= bitFix:
			return FixStr
		case n <= bit8:
			return Str8
		case n <= bit16:
			return Str16
		case n <= bit32:
			return Str32
		}
	case []byte:
		n := len(v.([]byte))
		switch {
		case n <= bit8:
			return Bin8
		case n <= bit16:
			return Bin16
		case n <= bit32:
			return Bin32
		}
	case map[string]any:
		n := len(v.(map[string]any))
		switch {
		case n <= (bitFix / 2):
			return FixMap
		case n <= bit16:
			return Map16
		case n <= bit32:
			return Map32
		}
	case []any:
		n := len(v.([]any))
		switch {
		case n <= (bitFix / 2):
			return FixArray
		case n <= bit16:
			return Array16
		case n <= bit32:
			return Array32
		}
	default:
		// use reflect as last resort for map and array types
		rv := reflect.ValueOf(v)

		if rv.Kind() == reflect.Map {
			n := rv.Len()
			switch {
			case n <= (bitFix / 2):
				return FixMap
			case n <= bit16:
				return Map16
			case n <= bit32:
				return Map32
			}
		}
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			n := rv.Len()
			switch {
			case n <= (bitFix / 2):
				return FixArray
			case n <= bit16:
				return Array16
			case n <= bit32:
				return Array32
			}
		}
	}
	return Unused
}
