package encoding

import (
	"reflect"
)

const (
	bufSize = 512
	intSize = 32 << (^uint(0) >> 63)

	Unknown = 0x10

	Nil = 0x11

	Bool      = 0x20
	BoolTrue  = 0x21
	BoolFalse = 0x22

	Float32 = 0x30
	Float64 = 0x31

	Int   = 0x40
	Int8  = 0x41
	Int16 = 0x42
	Int32 = 0x43
	Int64 = 0x44

	Uint   = 0x50
	Uint8  = 0x51
	Uint16 = 0x52
	Uint32 = 0x53
	Uint64 = 0x54

	String = 0x60
	Bytes  = 0x70

	Array       = 0x80
	ArrayString = 0x81

	Map     = 0x90
	Struct  = 0xa0
	Pointer = 0xb0
)

func ParseStruct(v any, fn func(reflect.StructField, reflect.Value)) {
	val := reflect.Indirect(reflect.ValueOf(v))
	for i := 0; i < val.NumField(); i++ {
		fn(val.Type().Field(i), val.Field(i))
	}
}
