package encoding

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"reflect"
	"strings"
)

type Decoder struct {
	r   *bufio.Reader
	buf []byte
	off int
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:   bufio.NewReader(r),
		buf: make([]byte, bufSize),
	}
}

func (d *Decoder) checkRead(n int) {
	_, err := d.r.Read(d.buf[d.off : d.off+n])
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Printf("error reading %d bytes: %s\n", n, err)
	}
}

func (d *Decoder) Decode() (any, error) {
	// Read data into buffer
	_, err := d.r.Read(d.buf)
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
	}
	// Decode value
	return d.readValue(), nil
}

func (d *Decoder) readValue() any {
	d.checkRead(1)
	typ := d.buf[d.off]
	var v any
	switch typ {
	case Nil:
		v = d.read1()
	case Bool:
		_, v = d.read2()
	case Float32:
		_, v = d.read5()
	case Float64:
		_, v = d.read9()
	case Int8:
		_, v = d.read2()
	case Int16:
		_, v = d.read3()
	case Int32:
		_, v = d.read5()
	case Int64:
		_, v = d.read9()
	case Uint8:
		_, v = d.read2()
	case Uint16:
		_, v = d.read3()
	case Uint32:
		_, v = d.read5()
	case Uint64:
		_, v = d.read9()
	case String:
		v = d.readString()
	case Bytes:
		v = d.readBytes()
	case Array:
		v = d.readArray()
	case ArrayString:
		v = d.readArrayString()
	case Map:
		v = d.readMap()
	default:
		v = d.readStruct()
	}
	return v
}

func (d *Decoder) read1() (v uint8) {
	d.checkRead(1)
	v = d.buf[d.off]
	d.off += 1
	return v
}

func (d *Decoder) read2() (t uint8, v uint8) {
	d.checkRead(2)
	t = d.buf[d.off]
	d.off += 1
	v = d.buf[d.off]
	d.off += 1
	return t, v
}

func (d *Decoder) read3() (t uint8, v uint16) {
	d.checkRead(3)
	t = d.buf[d.off]
	d.off += 1
	v = binary.BigEndian.Uint16(d.buf[d.off : d.off+2])
	d.off += 2
	return t, v
}

func (d *Decoder) read5() (t uint8, v uint32) {
	d.checkRead(5)
	t = d.buf[d.off]
	d.off += 1
	v = binary.BigEndian.Uint32(d.buf[d.off : d.off+4])
	d.off += 4
	return t, v
}

func (d *Decoder) read9() (t uint8, v uint64) {
	d.checkRead(9)
	t = d.buf[d.off]
	d.off += 1
	v = binary.BigEndian.Uint64(d.buf[d.off : d.off+8])
	d.off += 8
	return t, v
}

func (d *Decoder) readString() (v string) {
	_, sz := d.read3()
	n := int(sz)
	d.checkRead(n)
	var sb strings.Builder
	sb.Grow(n)
	sb.Write(d.buf[d.off : d.off+n])
	d.off += n
	return sb.String()
}

func (d *Decoder) readBytes() (v []byte) {
	_, sz := d.read5()
	n := int(sz)
	d.checkRead(n)
	v = make([]byte, n)
	copy(v, d.buf[d.off:d.off+n])
	d.off += n
	return v
}

func (d *Decoder) readArray() (v []any) {
	// max elements in array = 4,294,967,295
	_, sz := d.read5()
	n := int(sz)
	v = make([]any, n)
	for i := 0; i < n; i++ {
		v[i] = d.readValue()
	}
	return v
}

func (d *Decoder) readArrayString() (v []string) {
	// max elements in array = 4,294,967,295
	_, sz := d.read5()
	n := int(sz)
	v = make([]string, n)
	for i := 0; i < n; i++ {
		v[i] = d.readString()
	}
	return v
}

func (d *Decoder) readMap() (v map[string]any) {
	// max elements in map = 4,294,967,295
	_, sz := d.read5()
	n := int(sz)
	v = make(map[string]any, n)
	for i := 0; i < n; i++ {
		v[d.readString()] = d.readValue()
	}
	return v
}

func decAlloc(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func (d *Decoder) readStruct() (v any) {
	// struct layout
	// [type][numFields][fieldName][fieldValue]

	// get type and num fields
	typ, num := d.read3()
	if typ != Struct {
		return nil
	}

	// create slice of fields
	fields := make([]reflect.StructField, num)
	values := make([]reflect.Value, num)
	for i := 0; i < int(num); i++ {
		// read field name
		fn := d.readString()

		// read field value
		fv := d.readValue()

		// fill out struct field
		fields[i] = reflect.StructField{
			Name: fn,
			Type: reflect.TypeOf(fv),
		}
		// add to value
		values[i] = reflect.ValueOf(fv)
	}

	// create new struct type
	sct := reflect.New(reflect.StructOf(fields)).Elem()

	// set the values for each field
	for i := 0; i < int(num); i++ {
		sct.Field(i).Set(decAlloc(values[i]))
	}

	// return new struct
	return sct.Addr().Interface()
}
