package dopedb

import (
	"bufio"
	"encoding/binary"
	"io"
	"math"
	"time"
)

type Decoder struct {
	r   *bufio.Reader
	buf []byte
	at  int
	end int
	eof bool
}

func NewDecoder(r io.Reader) *Decoder {
	d := &Decoder{
		r:   bufio.NewReader(r),
		buf: make([]byte, bufSize),
		end: 0,
	}
	// for {
	// 	read, err := d.r.Read(d.buf)
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			d.eof = true
	// 			break
	// 		}
	// 		log.Panicf("error reading into buffer: %s\n", err)
	// 	}
	// 	d.end += read
	// }
	return d
}

func (d *Decoder) unread() int {
	return len(d.buf[:d.end])
}

// checkRead checks to see if we have n bytes in the buffer
// we can read from. If not, we read more into the buffer.
func (d *Decoder) checkRead(n int) {
	// // if at eof, panic
	// if d.eof == true {
	// 	// panic("cannot read any more, reached EOF")
	// 	log.Println("cannot read any more, reached EOF")
	// }
	// // buffer is empty, attempt to read data into buffer.
	// if d.end == 0 {
	// 	for {
	// 		read, err := d.r.Read(d.buf)
	// 		if err != nil {
	// 			if err == io.EOF {
	// 				d.eof = true
	// 				break
	// 			}
	// 			log.Panicf("error reading into buffer: %s\n", err)
	// 		}
	// 		d.end += read
	// 	}
	// 	return
	// }
	// buffer does not have n bytes to read from, so read n more bytes
	available := len(d.buf[:d.end])
	// empty := len(d.buf[d.end:])
	if n > available && n+available < cap(d.buf) {
		// log.Printf("[DEBUG]: checkRead(%d) available=%d, empty=%d", n, available, empty)
		read, err := d.r.Read(d.buf[d.end : d.end+n])
		if err != nil {
			if err == io.EOF {
				d.eof = true
			}
		}
		d.end += read
	}
}

func (d *Decoder) readValue() any {
	d.checkRead(1)
	typ := d.buf[d.at]
	var v any = -1
	switch {
	case typ == Nil:
		v = d.readNil()
	case typ == BoolTrue || typ == BoolFalse:
		v = d.readBool()
	case typ == Float32:
		v = d.readFloat32()
	case typ == Float64:
		v = d.readFloat64()
	case typ == Uint8:
		v = d.readUint8()
	case typ == Uint16:
		v = d.readUint16()
	case typ == Uint32:
		v = d.readUint32()
	case typ == Uint64:
		v = d.readUint64()
	case FixInt <= typ && typ <= FixIntMax:
		v = d.readFixInt()
	case typ == Int8:
		v = d.readInt8()
	case typ == Int16:
		v = d.readInt16()
	case typ == Int32:
		v = d.readInt32()
	case typ == Int64:
		v = d.readInt64()
	case FixStr <= typ && typ <= FixStrMax:
		v = d.readFixStr()
	case typ == Str8:
		v = d.readStr8()
	case typ == Str16:
		v = d.readStr16()
	case typ == Str32:
		v = d.readStr32()
	case typ == Bin8:
		v = d.readBin8()
	case typ == Bin16:
		v = d.readBin16()
	case typ == Bin32:
		v = d.readBin32()
	case FixMap <= typ && typ <= FixMapMax:
		v = d.readFixMap()
	case typ == Map16:
		v = d.readMap16()
	case typ == Map32:
		v = d.readMap32()
	case FixArray <= typ && typ <= FixArrayMax:
		v = d.readFixArray()
	case typ == Array16:
		v = d.readArray16()
	case typ == Array32:
		v = d.readArray32()
	}
	return v
}

func (d *Decoder) Decode() (any, error) {
	// Read the contents into the buffer
	for {
		read, err := d.r.Read(d.buf)
		if err != nil {
			if err == io.EOF {
				d.eof = true
				break
			}
			return nil, err
		}
		d.end += read
	}
	return d.readValue(), nil
}

func (d *Decoder) readByte() byte {
	d.checkRead(1)
	v := d.buf[d.at]
	d.at += 1
	return v
}

func (d *Decoder) readBytes(n int) []byte {
	d.checkRead(n)
	v := make([]byte, n)
	size := copy(v, d.buf[d.at:d.at+n])
	d.at += size
	return v
}

func (d *Decoder) slice(n int) []byte {
	d.checkRead(n)
	if d.at+n > d.end {
		panic("LINE 179 IN DECODER, SLICE OUT OF BOUNDS")
	}
	sliceOf := d.buf[d.at : d.at+n]
	d.at += n
	return sliceOf
}

func (d *Decoder) read1fix(t Type) uint8 {
	d.checkRead(1)
	b := d.buf[d.at]
	d.at += 1
	return b &^ t
}

func (d *Decoder) peek1() (Type, uint8) {
	d.checkRead(1)
	b := d.buf[d.at]
	// typ := b & t
	// val := b &^ t
	if fixTypeStart <= b && b <= fixTypeEnd {
		if FixInt <= b && b <= FixIntMax {
			return FixInt, (b &^ FixInt)
		}
		if FixMap <= b && b <= FixMapMax {
			return FixMap, (b &^ FixMap)
		}
		if FixArray <= b && b <= FixArrayMax {
			return FixArray, (b &^ FixArray)
		}
		if FixStr <= b && b <= FixStrMax {
			return FixStr, (b &^ FixStr)
		}
	}
	if b == Nil {
		return Nil, b
	}
	if b == BoolTrue {
		return BoolTrue, b
	}
	if b == BoolFalse {
		return BoolFalse, b
	}
	panic("error reading 1 byte")
}

func (d *Decoder) read1() (Type, uint8) {
	d.checkRead(1)
	t, v := d.peek1()
	d.at += 1
	return t, v
}

func (d *Decoder) read2() (Type, uint8) {
	d.checkRead(2)
	t := d.buf[d.at]
	d.at += 1
	v := d.buf[d.at]
	d.at += 2
	return t, v
}

func (d *Decoder) read3() (Type, uint16) {
	d.checkRead(3)
	t := d.buf[d.at]
	d.at += 1
	v := binary.BigEndian.Uint16(d.buf[d.at : d.at+2])
	d.at += 2
	return t, v
}

func (d *Decoder) read5() (Type, uint32) {
	d.checkRead(5)
	t := d.buf[d.at]
	d.at += 1
	v := binary.BigEndian.Uint32(d.buf[d.at : d.at+4])
	d.at += 4
	return t, v
}

func (d *Decoder) read9() (Type, uint64) {
	d.checkRead(9)
	t := d.buf[d.at]
	d.at += 1
	v := binary.BigEndian.Uint64(d.buf[d.at : d.at+8])
	d.at += 8
	return t, v
}

func (d *Decoder) readNil() any {
	b := d.readByte()
	if b == Nil {
		return nil
	}
	panic("error decoding nil, type does not match expected encoding")
}

func (d *Decoder) readBool() bool {
	b := d.readByte()
	if b == BoolTrue {
		return true
	}
	if b == BoolFalse {
		return false
	}
	panic("error decoding bool, type does not match expected encoding")
}

func (d *Decoder) readFloat32() float32 {
	t, v := d.read5()
	if t != Float32 {
		panic("error decoding float32, type does not match expected encoding")
	}
	return math.Float32frombits(v)
}

func (d *Decoder) readFloat64() float64 {
	t, v := d.read9()
	if t != Float64 {
		panic("error decoding float64, type does not match expected encoding")
	}
	return math.Float64frombits(v)
}

func (d *Decoder) readUint8() uint8 {
	t, v := d.read2()
	if t != Uint8 {
		panic("error decoding uint8, type does not match expected encoding")
	}
	return v
}

func (d *Decoder) readUint16() uint16 {
	t, v := d.read3()
	if t != Uint16 {
		panic("error decoding uint16, type does not match expected encoding")
	}
	return v
}

func (d *Decoder) readUint32() uint32 {
	t, v := d.read5()
	if t != Uint32 {
		panic("error decoding uint32, type does not match expected encoding")
	}
	return v
}

func (d *Decoder) readUint64() uint64 {
	t, v := d.read9()
	if t != Uint64 {
		panic("error decoding uint64, type does not match expected encoding")
	}
	return v
}

func (d *Decoder) readFixInt() int {
	t, v := d.read1()
	if t != FixInt {
		panic("error decoding fix int, type does not match expected encoding")
	}
	return int(v)
}

func (d *Decoder) readInt8() int8 {
	t, v := d.read2()
	if t != Int8 {
		panic("error decoding int8, type does not match expected encoding")
	}
	return int8(v)
}

func (d *Decoder) readInt16() int16 {
	t, v := d.read3()
	if t != Int16 {
		panic("error decoding int16, type does not match expected encoding")
	}
	return int16(v)
}

func (d *Decoder) readInt32() int32 {
	t, v := d.read5()
	if t != Int32 {
		panic("error decoding int32, type does not match expected encoding")
	}
	return int32(v)
}

func (d *Decoder) readInt64() int64 {
	t, v := d.read9()
	if t != Int64 {
		panic("error decoding int64, type does not match expected encoding")
	}
	return int64(v)
}

func (d *Decoder) readStr() string {
	d.checkRead(1)
	b := d.buf[d.at]
	// n := b &^ FixStr
	// fmt.Printf("FixStr=0x%.2x, FixStrMax=0x%.2x, got=0x%.2x, isFixStr=%v\n",
	//	FixStr, FixStrMax, b, FixStr <= b && b <= FixStrMax)
	switch {
	case FixStr <= b && b <= FixStrMax:
		return d.readFixStr()
	case b == Str8:
		return d.readStr8()
	case b == Str16:
		return d.readStr16()
	case b == Str32:
		return d.readStr32()
	}
	panic("error decoding string, type does not match expected encoding")
}

func (d *Decoder) readFixStr() string {
	t, v := d.read1()
	if t != FixStr {
		panic("error decoding fix str, type does not match expected encoding")
	}
	n := int(v)
	d.checkRead(n)
	s := readString(d.slice(n))
	return s
}

func (d *Decoder) readStr8() string {
	t, v := d.read2()
	if t != Str8 {
		panic("error decoding str8, type does not match expected encoding")
	}
	n := int(v)
	d.checkRead(n)
	s := readString(d.slice(n))
	return s
}

func (d *Decoder) readStr16() string {
	t, v := d.read3()
	if t != Str16 {
		panic("error decoding str16, type does not match expected encoding")
	}
	n := int(v)
	d.checkRead(n)
	s := readString(d.slice(n))
	return s
}

func (d *Decoder) readStr32() string {
	t, v := d.read5()
	if t != Str32 {
		panic("error decoding str32, type does not match expected encoding")
	}
	n := int(v)
	d.checkRead(n)
	s := readString(d.slice(n))
	return s
}

func (d *Decoder) readBin8() []byte {
	t, v := d.read2()
	if t != Bin8 {
		panic("error decoding bin8, type does not match expected encoding")
	}
	n := int(v)
	d.checkRead(n)
	b := readBytes(d.slice(n))
	return b
}

func (d *Decoder) readBin16() []byte {
	t, v := d.read3()
	if t != Bin16 {
		panic("error decoding bin16, type does not match expected encoding")
	}
	n := int(v)
	d.checkRead(n)
	b := readBytes(d.slice(n))
	return b
}

func (d *Decoder) readBin32() []byte {
	t, v := d.read5()
	if t != Bin32 {
		panic("error decoding bin32, type does not match expected encoding")
	}
	n := int(v)
	d.checkRead(n)
	b := readBytes(d.slice(n))
	return b
}

func (d *Decoder) readFixArray() []any {
	t, v := d.read1()
	if t != FixArray {
		panic("error decoding fix array, type does not match expected encoding")
	}
	n := int(v)
	arr := make([]any, n)
	for i := 0; i < n; i++ {
		arr[i] = d.readValue()
	}
	return arr
}

func (d *Decoder) readArray16() []any {
	t, v := d.read3()
	if t != Array16 {
		panic("error decoding array 16, type does not match expected encoding")
	}
	n := int(v)
	arr := make([]any, n)
	for i := 0; i < n; i++ {
		arr[i] = d.readValue()
	}
	return arr
}

func (d *Decoder) readArray32() []any {
	t, v := d.read5()
	if t != Array32 {
		panic("error decoding array 32, type does not match expected encoding")
	}
	n := int(v)
	arr := make([]any, n)
	for i := 0; i < n; i++ {
		arr[i] = d.readValue()
	}
	return arr
}

func (d *Decoder) readFixMap() map[string]any {
	t, v := d.read1()
	if t != FixMap {
		panic("error decoding fix map, type does not match expected encoding")
	}
	n := int(v)
	m := make(map[string]any, n)
	for i := 0; i < n; i++ {
		key := d.readStr()
		val := d.readValue()
		m[key] = val
	}
	return m
}

func (d *Decoder) readMap16() map[string]any {
	t, v := d.read3()
	if t != Map16 {
		panic("error decoding map 16, type does not match expected encoding")
	}
	n := int(v)
	m := make(map[string]any, n)
	for i := 0; i < n; i++ {
		m[d.readStr()] = d.readValue()
	}
	return m
}

func (d *Decoder) readMap32() map[string]any {
	t, v := d.read5()
	if t != Map32 {
		panic("error decoding map 32, type does not match expected encoding")
	}
	n := int(v)
	m := make(map[string]any, n)
	for i := 0; i < n; i++ {
		m[d.readStr()] = d.readValue()
	}
	return m
}

func (d *Decoder) readFixExt1() (uint8, []byte) {
	t, v := d.read2()
	if t != FixExt1 {
		panic("error decoding fix ext 1, type does not match expected encoding")
	}
	return v, []byte{d.readByte()}
}

func (d *Decoder) readFixExt2() (uint8, []byte) {
	t, v := d.read2()
	if t != FixExt2 {
		panic("error decoding fix ext 2, type does not match expected encoding")
	}
	return v, d.readBytes(2)
}

func (d *Decoder) readFixExt4() (uint8, []byte) {
	t, v := d.read2()
	if t != FixExt4 {
		panic("error decoding fix ext 4, type does not match expected encoding")
	}
	return v, d.readBytes(4)
}

func (d *Decoder) readFixExt8() (uint8, []byte) {
	t, v := d.read2()
	if t != FixExt8 {
		panic("error decoding fix ext 8, type does not match expected encoding")
	}
	return v, d.readBytes(8)
}

func (d *Decoder) readExt8() (uint8, []byte) {
	t, v := d.read2()
	if t != Ext8 {
		panic("error decoding ext 8, type does not match expected encoding")
	}
	n := int(v)
	if n > bit8 {
		panic("cannot decode ext 8, size is to large")
	}
	return d.readByte(), d.readBytes(n)
}

func (d *Decoder) readExt16() (uint8, []byte) {
	t, v := d.read2()
	if t != Ext16 {
		panic("error decoding ext 16, type does not match expected encoding")
	}
	n := int(v)
	if n > bit16 {
		panic("cannot decode ext 16, size is to large")
	}
	return d.readByte(), d.readBytes(n)
}

func (d *Decoder) readExt32() (uint8, []byte) {
	t, v := d.read2()
	if t != Ext32 {
		panic("error decoding ext 32, type does not match expected encoding")
	}
	n := int(v)
	if n > bit32 {
		panic("cannot decode ext 32, size is to large")
	}
	return d.readByte(), d.readBytes(n)
}

func (d *Decoder) readExt() (uint8, []byte) {
	d.checkRead(1)
	typ := d.buf[d.end]
	switch typ {
	case FixExt1:
		return d.readFixExt1()
	case FixExt2:
		return d.readFixExt2()
	case FixExt4:
		return d.readFixExt4()
	case FixExt8:
		return d.readFixExt8()
	case Ext8:
		return d.readExt8()
	case Ext16:
		return d.readExt16()
	case Ext32:
		return d.readExt32()
	}
	panic("error decoding ext, type does not match expected encoding")
}

func (d *Decoder) readTime32() time.Time {
	t, _ := d.read2()
	if t != Time32 {
		panic("error decoding time32, type does not match expected encoding")
	}
	v := binary.BigEndian.Uint32(d.readBytes(4))
	return time.Unix(int64(v), 0)
}

func (d *Decoder) readTime64() time.Time {
	t, _ := d.read2()
	if t != Time64 {
		panic("error decoding time64, type does not match expected encoding")
	}
	v := binary.BigEndian.Uint64(d.readBytes(8))
	return time.Unix(int64(v), 0)
}
