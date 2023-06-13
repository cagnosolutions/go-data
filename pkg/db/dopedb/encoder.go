package dopedb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"math"
	"time"
)

const (
	bufSize = 512
)

type Encoder struct {
	w   *bufio.Writer
	buf []byte // temp write buffer
	off int    // offset
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:   bufio.NewWriter(w),
		buf: make([]byte, bufSize),
	}
}

func (e *Encoder) available() int {
	return len(e.buf[e.off:])
}

func (e *Encoder) reset() {
	e.buf = e.buf[:0]
	e.off = 0
}

// check checks the buffer to see if we can write n
// more bytes. If the buffer does not have the room
// to write n more bytes, it attempts to write and
// flushes the current buffer in order to make room.
// However, if writing the contents of the buffer will
// not create enough space the buffer will be grown.
func (e *Encoder) check(n int) {
	// First check to see if we can fit n bytes in the
	// current buffer
	if n < len(e.buf[e.off:]) {
		// Looks like we can, so we just return
		return
	}
	// If the available bytes in the buffer is smaller
	// than what we want to write, check to see if
	// emptying the buffer will allow us to write n bytes.
	if n < len(e.buf) {
		// We can write n bytes if we empty the buffer.
		_, err := e.w.Write(e.buf[:e.off])
		if err != nil {
			panic("error writing buffer")
		}
		// Reset the buffer after writing.
		e.buf = e.buf[:0]
		e.off = 0
		return
	}
	// Looks like we need to write more data than what we
	// have available in the buffer, we will need to grow
	// the buffer.
	if n > bufSize {
		// Add e.off to account for e.buf[:e.off] being sliced off the front.
		e.buf = growSlice(e.buf[e.off:], e.off+n)
	}
}

// growSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
//
// This code was ripped off of the go source found at the link below:
// https://cs.opensource.google/go/go/+/master:src/bytes/buffer.go;l=229
func growSlice(b []byte, n int) []byte {
	defer func() {
		if recover() != nil {
			panic(bytes.ErrTooLarge)
		}
	}()
	// TODO(http://golang.org/issue/51462): We should rely on the append-make
	// pattern so that the compiler can call runtime.growslice. For example:
	//	return append(b, make([]byte, n)...)
	// This avoids unnecessary zero-ing of the first len(b) bytes of the
	// allocated slice, but this pattern causes b to escape onto the heap.
	//
	// Instead use the append-make pattern with a nil slice to ensure that
	// we allocate buffers rounded up to the closest size class.
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		// The growth rate has historically always been 2x. In the future,
		// we could rely purely on append to determine the growth rate.
		c = 2 * cap(b)
	}
	b2 := append([]byte(nil), make([]byte, c)...)
	copy(b2, b)
	return b2[:len(b)]
}

func (e *Encoder) Encode(v any) error {
	switch v.(type) {
	case nil:
		e.writeNil()
	case bool:
		var ok bool
		if v == true {
			ok = true
		}
		e.writeBool(ok)
	case float32:
		e.writeFloat32(v.(float32))
	case float64:
		e.writeFloat64(v.(float64))
	case uint:
		if intSize == 32 {
			e.writeUint32(v.(uint32))
			break
		}
		e.writeUint64(v.(uint64))
	case uint8:
		e.writeUint8(v.(uint8))
	case uint16:
		e.writeUint16(v.(uint16))
	case uint32:
		e.writeUint32(v.(uint32))
	case uint64:
		e.writeUint64(v.(uint64))
	case int:
		if intSize == 32 {
			e.writeInt32(v.(int32))
			break
		}
		e.writeInt64(v.(int64))
	case int8:
		e.writeInt8(v.(int8))
	case int16:
		e.writeInt16(v.(int16))
	case int32:
		e.writeInt32(v.(int32))
	case int64:
		e.writeInt64(v.(int64))
	case string:
		n := len(v.(string))
		switch {
		case n <= bitFix:
			e.writeFixStr(v.(string))
		case n <= bit8:
			e.writeStr8(v.(string))
		case n <= bit16:
			e.writeStr16(v.(string))
		case n <= bit32:
			e.writeStr32(v.(string))
		}
	case []byte:
		n := len(v.([]byte))
		switch {
		case n <= bit8:
			e.writeBin8(v.([]byte))
		case n <= bit16:
			e.writeBin16(v.([]byte))
		case n <= bit32:
			e.writeBin32(v.([]byte))
		}
		// case map[string]any:
		// 	n := len(v.(map[string]any))
		// 	switch {
		// 	case n <= (bitFix / 2):
		// 		encFixMap(b, v.(map[string]any))
		// 	case n <= bit16:
		// 		encMap16(b, v.(map[string]any))
		// 	case n <= bit32:
		// 		encMap32(b, v.(map[string]any))
		// 	}
		// case []any:
		// 	n := len(v.([]any))
		// 	switch {
		// 	case n <= (bitFix / 2):
		// 		encFixArray(b, v.([]any))
		// 	case n <= bit16:
		// 		encArray16(b, v.([]any))
		// 	case n <= bit32:
		// 		encArray32(b, v.([]any))
		// 	}
	}
	// Write the contents of the buffer
	_, err := e.w.Write(e.buf[:e.off])
	if err != nil {
		return err
	}
	// Flush to disk
	err = e.w.Flush()
	if err != nil {
		return err
	}
	// Reset the buffer
	e.buf = e.buf[:0]
	e.off = 0
	return nil
}

func (e *Encoder) writeByte(v byte) {
	e.check(1)
	e.buf[e.off] = v
	e.off += 1
}

func (e *Encoder) writeBytes(v []byte) {
	e.check(len(v))
	n := copy(e.buf[e.off:], v)
	e.off += n
}

func (e *Encoder) write1(t Type, v uint8) {
	e.check(1)
	e.buf[e.off] = byte(t | v)
	e.off += 1
}

func (e *Encoder) write2(t Type, v uint8) {
	e.check(2)
	e.buf[e.off] = t
	e.off += 1
	e.buf[e.off] = v
	e.off += 1
}

func (e *Encoder) write3(t Type, v uint16) {
	e.check(3)
	e.buf[e.off] = t
	e.off += 1
	binary.BigEndian.PutUint16(e.buf[e.off:e.off+2], v)
	e.off += 2
}

func (e *Encoder) write5(t Type, v uint32) {
	e.check(5)
	e.buf[e.off] = t
	e.off += 1
	binary.BigEndian.PutUint32(e.buf[e.off:e.off+4], v)
	e.off += 4
}

func (e *Encoder) write9(t Type, v uint64) {
	e.check(9)
	e.buf[e.off] = t
	e.off += 1
	binary.BigEndian.PutUint64(e.buf[e.off:e.off+8], v)
	e.off += 8
}

func (e *Encoder) writeNil() {
	e.writeByte(Nil)
}

func (e *Encoder) writeBool(v bool) {
	if v {
		e.writeByte(BoolTrue)
		return
	}
	e.writeByte(BoolFalse)
}

func (e *Encoder) writeFloat32(v float32) {
	e.write5(Float32, math.Float32bits(v))
}

func (e *Encoder) writeFloat64(v float64) {
	e.write9(Float64, math.Float64bits(v))
}

func (e *Encoder) writeUint8(v uint8) {
	e.write2(Uint8, v)
}

func (e *Encoder) writeUint16(v uint16) {
	e.write3(Uint16, v)
}

func (e *Encoder) writeUint32(v uint32) {
	e.write5(Uint32, v)
}

func (e *Encoder) writeUint64(v uint64) {
	e.write9(Uint64, v)
}

func (e *Encoder) writeInt8(v int8) {
	e.write2(Int8, uint8(v))
}

func (e *Encoder) writeInt16(v int16) {
	e.write3(Int16, uint16(v))
}

func (e *Encoder) writeInt32(v int32) {
	e.write5(Int32, uint32(v))
}

func (e *Encoder) writeInt64(v int64) {
	e.write9(Int64, uint64(v))
}

func (e *Encoder) writeStr(v string) {
	n := len(v)
	switch {
	case n <= bitFix:
		e.writeFixStr(v)
	case n <= bit8:
		e.writeStr8(v)
	case n <= bit16:
		e.writeStr16(v)
	case n <= bit32:
		e.writeStr32(v)
	}
}

func (e *Encoder) writeFixStr(v string) {
	e.write1(FixStr, uint8(len(v)))
	e.check(len(v))
	n := copy(e.buf[e.off:], v)
	e.off += n
}

func (e *Encoder) writeStr8(v string) {
	e.write2(Str8, uint8(len(v)))
	e.check(len(v))
	n := copy(e.buf[e.off:], v)
	e.off += n
}

func (e *Encoder) writeStr16(v string) {
	e.write3(Str16, uint16(len(v)))
	e.check(len(v))
	n := copy(e.buf[e.off:], v)
	e.off += n
}

func (e *Encoder) writeStr32(v string) {
	e.write5(Str32, uint32(len(v)))
	e.check(len(v))
	n := copy(e.buf[e.off:], v)
	e.off += n
}

func (e *Encoder) writeBin8(v []byte) {
	e.write2(Bin8, uint8(len(v)))
	e.writeBytes(v)
}

func (e *Encoder) writeBin16(v []byte) {
	e.write3(Bin16, uint16(len(v)))
	e.writeBytes(v)
}

func (e *Encoder) writeBin32(v []byte) {
	e.write5(Bin32, uint32(len(v)))
	e.writeBytes(v)
}

func (e *Encoder) writeFixArray(v []any) {
	if len(v) > bitFix/2 { // 15
		panic("cannot encode, type does not match expected encoding")
	}
	e.write1(FixArray, uint8(len(v)))
	for i := range v {
		err := e.Encode(v[i])
		if err != nil {
			log.Panicf("error encoding fix array element [%T]: %s\n", v[i], err)
		}
	}
}

func (e *Encoder) writeArray16(v []any) {
	if len(v) > bit16 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write3(Array16, uint16(len(v)))
	for i := range v {
		err := e.Encode(v[i])
		if err != nil {
			log.Panicf("error encoding array 16 element [%T]: %s\n", v[i], err)
		}
	}
}

func (e *Encoder) writeArray32(v []any) {
	if len(v) > bit32 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write5(Array32, uint32(len(v)))
	for i := range v {
		err := e.Encode(v[i])
		if err != nil {
			log.Panicf("error encoding array 32 element [%T]: %s\n", v[i], err)
		}
	}
}

func (e *Encoder) writeFixMap(m map[string]any) {
	if len(m) > bitFix/2 { // 15
		panic("cannot encode, type does not match expected encoding")
	}
	e.write1(FixMap, uint8(len(m)))
	for k, v := range m {
		e.writeStr(k)
		err := e.Encode(v)
		if err != nil {
			log.Panicf("error encoding fix map value [%T]: %s\n", v, err)
		}
	}
}

func (e *Encoder) writeMap16(m map[string]any) {
	if len(m) > bit16 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write3(Map16, uint16(len(m)))
	for k, v := range m {
		e.writeStr(k)
		err := e.Encode(v)
		if err != nil {
			log.Panicf("error encoding map 16 value [%T]: %s\n", v, err)
		}
	}
}

func (e *Encoder) writeMap32(m map[string]any) {
	if len(m) > bit32 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write5(Map32, uint32(len(m)))
	for k, v := range m {
		e.writeStr(k)
		err := e.Encode(v)
		if err != nil {
			log.Panicf("error encoding map 32 value [%T]: %s\n", v, err)
		}
	}
}

func (e *Encoder) writeFixExt1(t uint8, d byte) {
	e.write2(FixExt1, t)
	e.writeByte(d)
}

func (e *Encoder) writeFixExt2(t uint8, d []byte) {
	if len(d) > 2 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write2(FixExt2, t)
	e.writeBytes(d)
}

func (e *Encoder) writeFixExt4(t uint8, d []byte) {
	if len(d) > 4 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write2(FixExt4, t)
	e.writeBytes(d)
}

func (e *Encoder) writeFixExt8(t uint8, d []byte) {
	if len(d) > 8 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write2(FixExt8, t)
	e.writeBytes(d)
}

func (e *Encoder) writeExt8(t uint8, d []byte) {
	if len(d) > bit8 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write2(Ext8, uint8(len(d)))
	e.writeByte(t)
	e.writeBytes(d)
}

func (e *Encoder) writeExt16(t uint8, d []byte) {
	if len(d) > bit16 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write3(Ext16, uint16(len(d)))
	e.writeByte(t)
	e.writeBytes(d)
}

func (e *Encoder) writeExt32(t uint8, d []byte) {
	if len(d) > bit32 {
		panic("cannot encode, type does not match expected encoding")
	}
	e.write5(Ext32, uint32(len(d)))
	e.writeByte(t)
	e.writeBytes(d)
}

func (e *Encoder) writeExt(t uint8, d []byte) {
	n := len(d)
	switch {
	case n == 1:
		e.writeFixExt1(t, d[0])
	case n == 2:
		e.writeFixExt2(t, d)
	case n == 4:
		e.writeFixExt4(t, d)
	case n == 8:
		e.writeFixExt8(t, d)
	case n <= bit8:
		e.writeExt8(t, d)
	case n <= bit16:
		e.writeExt16(t, d)
	case n <= bit32:
		e.writeExt32(t, d)
	}
}

func (e *Encoder) writeTime32(t time.Time) {
	e.write2(Time32, -1)
	binary.BigEndian.PutUint32(e.buf[e.off:e.off+4], uint32(t.Unix()))
	e.off += 4
}

func (e *Encoder) writeTime64(t time.Time) {
	e.write2(Time64, -1)
	binary.BigEndian.PutUint64(e.buf[e.off:e.off+4], uint64(t.Unix()))
	e.off += 8
}
