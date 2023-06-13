package dopedb

import (
	"encoding/binary"
	"math"
)

// Bool and Nil types

var Primitive primitiveTypes

type primitiveTypes struct{}

func (e *Encoder) EncNil(p []byte) {
	_ = p[0] // early bounds check to guarantee safety of writes below
	p[0] = Nil
}

func (e primitiveTypes) EncBool(p []byte, v bool) {
	_ = p[0] // early bounds check to guarantee safety of writes below
	if v {
		p[0] = BoolTrue
		return
	}
	p[0] = BoolFalse
}

func (e primitiveTypes) EncFloat32(p []byte, v float32) {
	_ = p[5] // early bounds check to guarantee safety of writes below
	p[0] = Float32
	binary.BigEndian.PutUint32(p[1:5], math.Float32bits(v))
}

func (e primitiveTypes) EncFloat64(p []byte, v float64) {
	_ = p[9] // early bounds check to guarantee safety of writes below
	p[0] = Float64
	binary.BigEndian.PutUint64(p[1:9], math.Float64bits(v))
}

func (e primitiveTypes) EncUint(p []byte, v uint) {
	if intSize == 32 {
		e.EncUint32(p, uint32(v))
		return
	}
	e.EncUint64(p, uint64(v))
}

func (e primitiveTypes) EncUint8(p []byte, v uint8) {
	_ = p[1] // early bounds check to guarantee safety of writes below
	p[0] = Uint8
	p[1] = v
}

func (e primitiveTypes) EncUint16(p []byte, v uint16) {
	_ = p[3] // early bounds check to guarantee safety of writes below
	p[0] = Uint16
	binary.BigEndian.PutUint16(p[1:3], v)
}

func (e primitiveTypes) EncUint32(p []byte, v uint32) {
	_ = p[5] // early bounds check to guarantee safety of writes below
	p[0] = Uint32
	binary.BigEndian.PutUint32(p[1:5], v)
}

func (e primitiveTypes) EncUint64(p []byte, v uint64) {
	_ = p[9] // early bounds check to guarantee safety of writes below
	p[0] = Uint64
	binary.BigEndian.PutUint64(p[1:9], v)
}

func (e primitiveTypes) EncFixInt(p []byte, v int) {
	_ = p[0] // early bounds check to guarantee safety of writes below
	if v > FixIntMax-FixInt {
		panic("value is too large to encoded as a fix int type")
	}
	p[0] = uint8(FixInt | v)
}

func (e primitiveTypes) EncInt(p []byte, v int) {
	_ = p[1] // early bounds check to guarantee safety of writes below
	if intSize == 32 {
		e.EncInt32(p, int32(v))
		return
	}
	e.EncInt64(p, int64(v))
}

func (e primitiveTypes) EncInt8(p []byte, v int8) {
	_ = p[1] // early bounds check to guarantee safety of writes below
	p[0] = Int8
	p[1] = uint8(v)
}

func (e primitiveTypes) EncInt16(p []byte, v int16) {
	_ = p[3] // early bounds check to guarantee safety of writes below
	p[0] = Int16
	binary.BigEndian.PutUint16(p[1:3], uint16(v))
}

func (e primitiveTypes) EncInt32(p []byte, v int32) {
	_ = p[5] // early bounds check to guarantee safety of writes below
	p[0] = Int32
	binary.BigEndian.PutUint32(p[1:5], uint32(v))
}

func (e primitiveTypes) EncInt64(p []byte, v int64) {
	_ = p[9] // early bounds check to guarantee safety of writes below
	p[0] = Int64
	binary.BigEndian.PutUint64(p[1:9], uint64(v))
}

func (e primitiveTypes) DecNil(p []byte) any {
	_ = p[0] // bounds check hint to compiler
	if p[0] != Nil {
		panic("cannot decode, type does not match expected type")
	}
	return nil
}

func (e primitiveTypes) DecBool(p []byte) bool {
	_ = p[0] // bounds check hint to compiler
	if p[0] != BoolTrue && p[0] != BoolFalse {
		panic("cannot decode, type does not match expected type")
	}
	if p[0] == BoolTrue {
		return true
	}
	return false
}

func (e primitiveTypes) DecFloat32(p []byte) float32 {
	_ = p[5] // bounds check hint to compiler
	if p[0] != Float32 {
		panic("cannot decode, type does not match expected type")
	}
	return math.Float32frombits(binary.BigEndian.Uint32(p[1:5]))
}

func (e primitiveTypes) DecFloat64(p []byte) float64 {
	_ = p[9] // bounds check hint to compiler
	if p[0] != Float64 {
		panic("cannot decode, type does not match expected type")
	}
	return math.Float64frombits(binary.BigEndian.Uint64(p[1:9]))
}

func (e primitiveTypes) DecUint(p []byte) uint {
	if intSize == 32 {
		return uint(e.DecUint32(p))
	}
	return uint(e.DecUint64(p))
}

func (e primitiveTypes) DecUint8(p []byte) uint8 {
	_ = p[1] // bounds check hint to compiler
	if p[0] != Uint8 {
		panic("cannot decode, type does not match expected type")
	}
	return p[1]
}

func (e primitiveTypes) DecUint16(p []byte) uint16 {
	_ = p[3] // bounds check hint to compiler
	if p[0] != Uint16 {
		panic("cannot decode, type does not match expected type")
	}
	return binary.BigEndian.Uint16(p[1:3])
}

func (e primitiveTypes) DecUint32(p []byte) uint32 {
	_ = p[5] // bounds check hint to compiler
	if p[0] != Uint32 {
		panic("cannot decode, type does not match expected type")
	}
	return binary.BigEndian.Uint32(p[1:5])
}

func (e primitiveTypes) DecUint64(p []byte) uint64 {
	_ = p[9] // bounds check hint to compiler
	if p[0] != Uint64 {
		panic("cannot decode, type does not match expected type")
	}
	return binary.BigEndian.Uint64(p[1:9])
}

func (e primitiveTypes) DecFixInt(p []byte) int {
	_ = p[0] // bounds check hint to compiler
	if p[0]&FixInt != FixInt {
		panic("cannot decode, type does not match expected type")
	}
	return int(p[0] &^ FixInt)
}

func (e primitiveTypes) DecInt(p []byte) int {
	if intSize == 32 {
		return int(e.DecInt32(p))
	}
	return int(e.DecInt64(p))
}

func (e primitiveTypes) DecInt8(p []byte) int8 {
	_ = p[1] // bounds check hint to compiler
	if p[0] != Int8 {
		panic("cannot decode, type does not match expected type")
	}
	return int8(p[1])
}

func (e primitiveTypes) DecInt16(p []byte) int16 {
	_ = p[3] // bounds check hint to compiler
	if p[0] != Int16 {
		panic("cannot decode, type does not match expected type")
	}
	return int16(binary.BigEndian.Uint16(p[1:3]))
}

func (e primitiveTypes) DecInt32(p []byte) int32 {
	_ = p[5] // bounds check hint to compiler
	if p[0] != Int32 {
		panic("cannot decode, type does not match expected type")
	}
	return int32(binary.BigEndian.Uint32(p[1:5]))
}

func (e primitiveTypes) DecInt64(p []byte) int64 {
	_ = p[9] // bounds check hint to compiler
	if p[0] != Int64 {
		panic("cannot decode, type does not match expected type")
	}
	return int64(binary.BigEndian.Uint64(p[1:9]))
}

/*
func WriteBool(w io.Writer, ok bool) (int, error) {
	if ok {
		return w.Write([]byte{BoolTrue})
	}
	return w.Write([]byte{BoolFalse})
}

func ReadBool(r io.Reader) (bool, error) {
	p := make([]byte, 1)
	_, err := r.Read(p)
	if err != nil {
		return false, err
	}
	if p[0] != BoolTrue && p[0] != BoolFalse {
		return false, ErrInvalidType
	}
	if p[0] == BoolTrue {
		return true, nil
	}
	if p[0] == BoolFalse {
		return false, nil
	}
	return false, err
}

func WriteNil(w io.Writer) (int, error) {
	return w.Write([]byte{Nil})
}

func ReadNil(r io.Reader) (any, error) {
	if !hasRoom(p, 1) {
		panic(ErrReadingBuffer)
	}
	p := make([]byte, 1)
	_, err := r.Read(p)
	if p[0] == Nil {
		return nil
	}
	panic("byte is not a nil byte")
}

// Uint8, Uint16, Uint32 and Uint64 types

func encUint8(p []byte, n uint8) {
	if !hasRoom(p, 2) {
		panic(ErrWritingBuffer)
	}
	p[0] = Uint8
	p[1] = n
}

func encUint16(p []byte, n uint16) {
	if !hasRoom(p, 3) {
		panic(ErrWritingBuffer)
	}
	p[0] = Uint16
	binary.BigEndian.PutUint16(p[1:3], n)
}

func encUint32(p []byte, n uint32) {
	if !hasRoom(p, 5) {
		panic(ErrWritingBuffer)
	}
	p[0] = Uint32
	binary.BigEndian.PutUint32(p[1:5], n)
}

func encUint64(p []byte, n uint64) {
	if !hasRoom(p, 9) {
		panic(ErrWritingBuffer)
	}
	p[0] = Uint64
	binary.BigEndian.PutUint64(p[1:9], n)
}

func decUint8(p []byte) uint8 {
	if !hasRoom(p, 2) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Uint8 {
		panic("not a uint8 type")
	}
	return p[1]
}

func decUint16(p []byte) uint16 {
	if !hasRoom(p, 3) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Uint16 {
		panic("not a uint16 type")
	}
	return binary.BigEndian.Uint16(p[1:3])
}

func decUint32(p []byte) uint32 {
	if !hasRoom(p, 5) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Uint32 {
		panic("not a uint32 type")
	}
	return binary.BigEndian.Uint32(p[1:5])
}

func decUint64(p []byte) uint64 {
	if !hasRoom(p, 9) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Uint64 {
		panic("not a uint64 type")
	}
	return binary.BigEndian.Uint64(p[1:9])
}

// Int8, Int16, Int32 and Int64 types

func encFixInt(p []byte, n int) {
	if n > FixIntMax-FixInt {
		panic(encodingError("encFixInt"))
	}
	if !hasRoom(p, 1) {
		panic(ErrWritingBuffer)
	}
	p[0] = byte(FixInt | n)
}

func encInt8(p []byte, n int8) {
	if !hasRoom(p, 2) {
		panic(ErrWritingBuffer)
	}
	p[0] = Int8
	p[1] = byte(n)
}

func encInt16(p []byte, n int16) {
	if !hasRoom(p, 3) {
		panic(ErrWritingBuffer)
	}
	p[0] = Int16
	binary.BigEndian.PutUint16(p[1:3], uint16(n))
}

func encInt32(p []byte, n int32) {
	if !hasRoom(p, 5) {
		panic(ErrWritingBuffer)
	}
	p[0] = Int32
	binary.BigEndian.PutUint32(p[1:5], uint32(n))
}

func encInt64(p []byte, n int64) {
	if !hasRoom(p, 9) {
		panic(ErrWritingBuffer)
	}
	p[0] = Int64
	binary.BigEndian.PutUint64(p[1:9], uint64(n))
}

func decFixInt(p []byte) int {
	if !hasRoom(p, 1) {
		panic(ErrReadingBuffer)
	}
	if p[0]&FixInt != FixInt {
		panic("not a fixint type")
	}
	return int(p[0] &^ FixInt)
}

func decInt8(p []byte) int8 {
	if !hasRoom(p, 2) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Int8 {
		panic("not a int8 type")
	}
	return int8(p[1])
}

func decInt16(p []byte) int16 {
	if !hasRoom(p, 3) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Int16 {
		panic("not a int16 type")
	}
	return int16(binary.BigEndian.Uint16(p[1:3]))
}

func decInt32(p []byte) int32 {
	if !hasRoom(p, 5) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Int32 {
		panic("not a int32 type")
	}
	return int32(binary.BigEndian.Uint32(p[1:5]))
}

func decInt64(p []byte) int64 {
	if !hasRoom(p, 9) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Int64 {
		panic("not a int64 type")
	}
	return int64(binary.BigEndian.Uint64(p[1:9]))
}

// Float32 and Float64 types

func WriteFloat32(w io.Writer, n float32) (int, error) {
	// if !hasRoom(p, 5) {
	// 	panic(ErrWritingBuffer)
	// }
	p := make([]byte, 5)
	p[0] = Float32
	binary.BigEndian.PutUint32(p[1:5], math.Float32bits(n))
	return w.Write(p)
}

func WriteFloat64(w io.Writer, n float64) (int, error) {
	// if !hasRoom(p, 9) {
	// 	panic(ErrWritingBuffer)
	// }
	p := make([]byte, 9)
	p[0] = Float64
	binary.BigEndian.PutUint64(p[1:9], math.Float64bits(n))
	return w.Write(p)
}

var ErrInvalidType = errors.New("the encoding type did not match the expected type")

func ReadFloat32(r io.Reader, v *float32) (int, error) {
	// if !hasRoom(p, 5) {
	// 	panic(ErrReadingBuffer)
	// }
	p := make([]byte, 5)
	n, err := r.Read(p)
	if err != nil {
		return 0, err
	}
	if p[0] != Float32 {
		return 0, ErrInvalidType
	}
	*v = math.Float32frombits(binary.BigEndian.Uint32(p[1:]))
	return n, err
}

func decFloat64(p []byte) float64 {
	if !hasRoom(p, 9) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Float64 {
		panic("not a float64 type")
	}
	return math.Float64frombits(binary.BigEndian.Uint64(p[1:]))
}
*/
