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
		log.Printf("DEBUG: check(%d), we have room in buffer\n", n)
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
		log.Printf("DEBUG: check(%d), just called write, resetting buffer\n", n)
		// Reset the buffer after writing.
		e.buf = e.buf[:0]
		e.off = 0
		return
	}
	// Looks like we need to write more data than what we
	// have available in the buffer, we will need to grow
	// the buffer.
	if n > bufSize {
		log.Printf("DEBUG: check(%d), growing buffer\n", n)
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

func (e *Encoder) encodeValue(v any) error {
	switch t := v.(type) {
	case nil:
		e.writeNil()
	case bool:
		var ok bool
		if t == true {
			ok = true
		}
		e.writeBool(ok)
	case float32:
		e.writeFloat32(t)
	case float64:
		e.writeFloat64(t)
	case uint:
		if intSize == 32 {
			e.writeUint32(v.(uint32))
			break
		}
		e.writeUint64(v.(uint64))
	case uint8:
		e.writeUint8(t)
	case uint16:
		e.writeUint16(t)
	case uint32:
		e.writeUint32(t)
	case uint64:
		e.writeUint64(t)
	case int:
		if t < 128 {
			e.writeFixInt(t)
			break
		}
		if intSize == 32 {
			e.writeInt32(v.(int32))
			break
		}
		e.writeInt64(v.(int64))
	case int8:
		e.writeInt8(t)
	case int16:
		e.writeInt16(t)
	case int32:
		e.writeInt32(t)
	case int64:
		e.writeInt64(t)
	case string:
		n := len(t)
		switch {
		case n <= bitFix:
			e.writeFixStr(t)
		case n <= bit8:
			e.writeStr8(t)
		case n <= bit16:
			e.writeStr16(t)
		case n <= bit32:
			e.writeStr32(t)
		}
	case []byte:
		n := len(t)
		switch {
		case n <= bit8:
			e.writeBin8(t)
		case n <= bit16:
			e.writeBin16(t)
		case n <= bit32:
			e.writeBin32(t)
		}
	case []any:
		n := len(t)
		switch {
		case n <= (bitFix / 2):
			e.writeFixArray(t)
		case n <= bit16:
			e.writeArray16(t)
		case n <= bit32:
			e.writeArray32(t)
			break
		}
	case map[string]any:
		n := len(t)
		switch {
		case n <= (bitFix / 2):
			e.writeFixMap(t)
		case n <= bit16:
			e.writeMap16(t)
		case n <= bit32:
			e.writeMap32(t)
		}
	}
	return nil
}

func (e *Encoder) Reset() {
	e.reset()
	e.w.Reset(e.w)
}

func (e *Encoder) Encode(v any) error {
	err := e.encodeValue(v)
	if err != nil {
		return err
	}
	// Write the contents of the buffer
	_, err = e.w.Write(e.buf[:e.off])
	if err != nil {
		return err
	}
	log.Printf("DEBUG: encoded value (%T=%#v) and called write\n", v, v)
	// Flush to disk
	err = e.w.Flush()
	if err != nil {
		return err
	}
	log.Printf("DEBUG: called flush on underlying writer\n")
	// Reset the buffer
	// e.buf = e.buf[:0]
	// e.off = 0
	// log.Printf("DEBUG: reset buffer after successful encoding\n")
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
	log.Printf("DEBUG: len(e.buf)=%d, e.off=%d\n", len(e.buf), e.off)
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

func (e *Encoder) writeFixInt(v int) {
	e.write1(FixInt, uint8(v))
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

func (e *Encoder) writeFixMap(m map[string]any) {
	if len(m) > bitFix/2 { // 15
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write1(FixMap, uint8(len(m)))
	for k, v := range m {
		e.writeStr(k)
		err := e.encodeValue(v)
		if err != nil {
			log.Panicf("error encoding fix map value [%T]: %s\n", v, err)
		}
	}
}

func (e *Encoder) writeMap16(m map[string]any) {
	if len(m) > bit16 {
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write3(Map16, uint16(len(m)))
	for k, v := range m {
		e.writeStr(k)
		err := e.encodeValue(v)
		if err != nil {
			log.Panicf("error encoding map 16 value [%T]: %s\n", v, err)
		}
	}
}

func (e *Encoder) writeMap32(m map[string]any) {
	if len(m) > bit32 {
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write5(Map32, uint32(len(m)))
	for k, v := range m {
		e.writeStr(k)
		err := e.encodeValue(v)
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
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write2(FixExt2, t)
	e.writeBytes(d)
}

func (e *Encoder) writeFixExt4(t uint8, d []byte) {
	if len(d) > 4 {
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write2(FixExt4, t)
	e.writeBytes(d)
}

func (e *Encoder) writeFixExt8(t uint8, d []byte) {
	if len(d) > 8 {
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write2(FixExt8, t)
	e.writeBytes(d)
}

func (e *Encoder) writeExt8(t uint8, d []byte) {
	if len(d) > bit8 {
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write2(Ext8, uint8(len(d)))
	e.writeByte(t)
	e.writeBytes(d)
}

func (e *Encoder) writeExt16(t uint8, d []byte) {
	if len(d) > bit16 {
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write3(Ext16, uint16(len(d)))
	e.writeByte(t)
	e.writeBytes(d)
}

func (e *Encoder) writeExt32(t uint8, d []byte) {
	if len(d) > bit32 {
		panic("cannot encodeValue, type does not match expected encoding")
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
	e.write2(Time32, 0)
	binary.BigEndian.PutUint32(e.buf[e.off:e.off+4], uint32(t.Unix()))
	e.off += 4
}

func (e *Encoder) writeTime64(t time.Time) {
	e.write2(Time64, 0)
	binary.BigEndian.PutUint64(e.buf[e.off:e.off+4], uint64(t.Unix()))
	e.off += 8
}
