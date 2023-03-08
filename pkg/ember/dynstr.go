package ember

import (
	"fmt"
	"unsafe"
)

const (
	dsKindNil    = 0x0000 // dsKindNil is nil or empty data
	dsKindStr    = 0x0001 // dsKindStr is string or byte data
	dsKindNum    = 0x0002 // dsKindNum is number data
	dsKindNumNeg = 0x0003
	dsKindBin    = 0x0004 // dsKindBin is raw bit data
	dsKindKV     = 0x0008 // dsKindKV is a key and value type
	dsKindList   = 0x0010 // dsKindList is a list type
	dsKindMap    = 0x0020 // dsKindMap is a map type
	_            = 0x0040
	_            = 0x0080
	_            = 0x0100
	_            = 0x0200
	_            = 0x0400
	_            = 0x0800
	dsKindErr    = 0x8000 // dsKindErr is an error

	dsTerm = 0x1e

	_ = 0x0f00
)

var kindMap = map[uint16]string{
	dsKindNil:  "empty",
	dsKindStr:  "string",
	dsKindNum:  "number",
	dsKindBin:  "binary",
	dsKindKV:   "key/value",
	dsKindList: "list",
	dsKindMap:  "map",
	dsKindErr:  "error",
}

func getKindAsString(k uint16) string {
	kind := getKind(k)
	return fmt.Sprintf("0x%.4x (%s)", kind, kindMap[kind])
}

func getKind(k uint16) uint16 {
	switch k & 0xffff {
	case dsKindNil:
		return dsKindNil
	case dsKindStr:
		return dsKindStr
	case dsKindNum:
		return dsKindNum
	case dsKindBin:
		return dsKindBin
	case dsKindKV:
		return dsKindKV
	case dsKindList:
		return dsKindList
	case dsKindMap:
		return dsKindMap
	}
	return dsKindErr
}

const (
	max64 = ^uint64(0)
	max32 = ^uint32(0)
	max16 = ^uint16(0)
	max8  = ^uint8(0)
)

func calcNum(n uint) int {
	switch {
	case n <= uint(max8):
		return 1
	case n <= uint(max16):
		return 2
	case n <= uint(max32):
		return 4
	case n <= uint(max64):
		return 8
	}
	return -1
}

// dynstr is a lot like the sds (simple dynamic strings) library derived
// from the original version used in redis. It is a particularly useful
// way of storing string type data on disk. We are simply replicating the
// way go stores strings in memory, but we do it on disk.
//
// Basic structure:
// +-------------+---------------+-----------+
// |    header   |     data      | separator |
// +-------------+---------------+-----------+
//
// Header structure:
// +------+------+--
// | kind | size | ... data
// +------+------+--
// kind and size are represented as uint16's
//
// Data and ending structure:
//
//	         --+----------------+-----------+
//	header ... |   byte array   |    0x1E   |
//	         --+----------------+-----------+
//
// The data portion is an array of bytes, ending with a
// special ascii record separator byte (0x1E).
type dynstr struct {
	kind uint16
	size uint16
	data []byte
}

func newNil() *dynstr {
	return &dynstr{
		kind: dsKindNil,
		size: 0,
		data: nil,
	}
}

func newStr(v string) *dynstr {
	return &dynstr{
		kind: dsKindStr,
		size: uint16(len(v)),
		data: []byte(v),
	}
}

func newByt(v []byte) *dynstr {
	return &dynstr{
		kind: dsKindStr,
		size: uint16(len(v)),
		data: v,
	}
}

func newNum(v int) *dynstr {
	kind := dsKindNum
	if v < 0 {
		kind = dsKindNumNeg
		v = ^v
	}
	sz := calcNum(uint(v))
	b := make([]byte, sz)
	switch sz {
	case 1:
		b[0] = uint8(v)
	case 2:
		bin.PutUint16(b, uint16(v))
	case 4:
		bin.PutUint32(b, uint32(v))
	case 8:
		bin.PutUint64(b, uint64(v))
	}
	return &dynstr{
		kind: uint16(kind),
		size: uint16(sz),
		data: b,
	}
}

func newBin(v []byte) *dynstr {
	return &dynstr{
		kind: dsKindBin,
		size: uint16(len(v)),
		data: v,
	}
}

func newKV() *dynstr {
	return &dynstr{
		kind: dsKindKV,
		size: 0,
		data: []byte{dsTerm},
	}
}

func newList() *dynstr {
	return &dynstr{
		kind: dsKindList,
		size: 0,
		data: []byte{dsTerm},
	}
}

func newMap() *dynstr {
	return &dynstr{
		kind: dsKindMap,
		size: 0,
		data: []byte{dsTerm},
	}
}

// sizeInMemory returns the number of bytes d would be
// represented with in memory
func (d *dynstr) sizeInMemory() int {
	size := int(unsafe.Sizeof(*d))
	size += len(d.data)
	return size
}

// sizeOnDisk returns the number of bytes d would be
// represented with on disk
func (d *dynstr) sizeOnDisk() int {
	return 2 + 2 + len(d.data) + 1
}
