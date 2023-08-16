package encoding

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"reflect"
)

type Encoder struct {
	w   *bufio.Writer
	buf []byte
	off int
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:   bufio.NewWriter(w),
		buf: make([]byte, bufSize),
	}
}

func (e *Encoder) checkWrite(n int) {
	// check to see if we have room in the current buffer
	if n < len(e.buf[e.off:]) {
		// we have room, so just return
		return
	}
	// check to see if we can fit n bytes in our buffer
	if n < len(e.buf) {
		// looks like we can, but we have to clear it
		// before writing more...
		_, err := e.w.Write(e.buf[:e.off])
		if err != nil {
			panic("error writing buffer")
		}
		// now we can reset it, so our write will work
		e.buf = e.buf[:0]
		e.off = 0
		return
	}
	// check to see if we need to grow our buffer
	if n > cap(e.buf) {
		// looks like we do...
		e.buf = growSlice(e.buf[e.off:], e.off+n)
	}
}

// growSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
//
// This code was ripped end of the go source found at the link below:
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

func (e *Encoder) Encode(v any) (err error) {
	e.writeValue(v)
	_, err = e.w.Write(e.buf[:e.off])
	if err != nil {
		return err
	}
	err = e.w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (e *Encoder) writeValue(v any) {
	switch t := v.(type) {
	case nil:
		e.write1(Nil)
	case bool:
		if t == true {
			e.write2(Bool, BoolTrue)
		}
		e.write2(Bool, BoolFalse)
	case float32:
		e.write5(Float32, math.Float32bits(t))
	case float64:
		e.write9(Float64, math.Float64bits(t))
	case int:
		if intSize == 32 {
			e.write5(Int32, uint32(t))
			break
		}
		e.write9(Int64, uint64(t))
	case int8:
		e.write2(Int8, uint8(t))
	case int16:
		e.write3(Int16, uint16(t))
	case int32:
		e.write5(Int32, uint32(t))
	case int64:
		e.write9(Int64, uint64(t))
	case uint:
		if intSize == 32 {
			e.write5(Uint32, uint32(t))
			break
		}
		e.write9(Uint64, uint64(t))
	case uint8:
		e.write2(Uint8, t)
	case uint16:
		e.write3(Uint16, t)
	case uint32:
		e.write5(Uint32, t)
	case uint64:
		e.write9(Uint64, t)
	case string:
		e.writeString(t)
	case []byte:
		e.writeBytes(t)
	case []any:
		e.writeArray(t)
	case []string:
		e.writeArrayString(t)
	case map[string]any:
		e.writeMap(t)
	default:
		val := reflect.Indirect(reflect.ValueOf(v))
		if val.Kind() == reflect.Struct {
			e.writeStruct(val)
		}
	}
}

func (e *Encoder) write1(v uint8) {
	e.checkWrite(1)
	e.buf[e.off] = v
	e.off += 1
}

func (e *Encoder) write2(t uint8, v uint8) {
	e.checkWrite(2)
	e.buf[e.off] = t
	e.off += 1
	e.buf[e.off] = v
	e.off += 1
}

func (e *Encoder) write3(t uint8, v uint16) {
	e.checkWrite(3)
	e.buf[e.off] = t
	e.off += 1
	binary.BigEndian.PutUint16(e.buf[e.off:e.off+2], v)
	e.off += 2
}

func (e *Encoder) write5(t uint8, v uint32) {
	e.checkWrite(5)
	e.buf[e.off] = t
	e.off += 1
	binary.BigEndian.PutUint32(e.buf[e.off:e.off+4], v)
	e.off += 4
}

func (e *Encoder) write9(t uint8, v uint64) {
	e.checkWrite(9)
	e.buf[e.off] = t
	e.off += 1
	binary.BigEndian.PutUint64(e.buf[e.off:e.off+8], v)
	e.off += 8
}

func (e *Encoder) writeString(v string) {
	// max string length = 65,535
	e.write3(String, uint16(len(v)))
	e.checkWrite(len(v))
	n := copy(e.buf[e.off:], v)
	e.off += n
}

func (e *Encoder) writeBytes(v []byte) {
	// max byte slice length = 4,294,967,295
	e.write5(Bytes, uint32(len(v)))
	e.checkWrite(len(v))
	n := copy(e.buf[e.off:], v)
	e.off += n
}

func (e *Encoder) writeArray(v []any) {
	// max elements in array = 4,294,967,295
	e.write5(Array, uint32(len(v)))
	for i := range v {
		e.checkWrite(len(v))
		e.writeValue(v[i])
	}
}

func (e *Encoder) writeArrayString(v []string) {
	// max elements in array = 4,294,967,295
	e.write5(ArrayString, uint32(len(v)))
	for i := range v {
		e.checkWrite(len(v[i]))
		e.writeString(v[i])
	}
}

func (e *Encoder) writeMap(v map[string]any) {
	// max elements in map = 4,294,967,295
	e.write5(Map, uint32(len(v)))
	for key, val := range v {
		e.writeString(key)
		e.writeValue(val)
	}
}

func (e *Encoder) writeStruct(v reflect.Value) {
	// struct layout
	// [type][numFields][fieldName][fieldType][fildValue]

	// write type and num fields
	e.write3(Struct, uint16(v.NumField()))
	for i := 0; i < v.NumField(); i++ {
		sf := v.Type().Field(i)
		sv := v.Field(i)

		// write field name
		e.writeString(sf.Name)

		// write field value
		e.writeValue(sv.Interface())
	}
}
