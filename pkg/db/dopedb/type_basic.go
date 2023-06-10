package dopedb

import (
	"encoding/binary"
	"math"
)

// Bool and Nil types

func encBool(p []byte, ok bool) {
	if !hasRoom(p, 1) {
		panic(ErrWritingBuffer)
	}
	if ok {
		p[0] = BoolTrue
		return
	}
	p[0] = BoolFalse
}

func decBool(p []byte) bool {
	if !hasRoom(p, 1) {
		panic(ErrReadingBuffer)
	}
	if p[0] == BoolTrue {
		return true
	}
	if p[0] == BoolFalse {
		return false
	}
	panic("byte is not a boolean byte")
}

func encNil(p []byte) {
	if !hasRoom(p, 1) {
		panic(ErrWritingBuffer)
	}
	p[0] = Nil
}

func decNil(p []byte) any {
	if !hasRoom(p, 1) {
		panic(ErrReadingBuffer)
	}
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

func encFloat32(p []byte, n float32) {
	if !hasRoom(p, 5) {
		panic(ErrWritingBuffer)
	}
	p[0] = Float32
	binary.BigEndian.PutUint32(p[1:5], math.Float32bits(n))
}

func encFloat64(p []byte, n float64) {
	if !hasRoom(p, 9) {
		panic(ErrWritingBuffer)
	}
	p[0] = Float64
	binary.BigEndian.PutUint64(p[1:9], math.Float64bits(n))
}

func decFloat32(p []byte) float32 {
	if !hasRoom(p, 5) {
		panic(ErrReadingBuffer)
	}
	if p[0] != Float32 {
		panic("not a float32 type")
	}
	return math.Float32frombits(binary.BigEndian.Uint32(p[1:]))
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
