package ember

import (
	"fmt"
	"strconv"
	"unsafe"
)

type sliceType = byte

var (
	wordSize     uint = uint(unsafe.Sizeof(sliceType(0)))
	log2WordSize uint = log2(wordSize)
)

func log2(i uint) uint {
	var n uint
	for ; i > 0; n++ {
		i >>= 1
	}
	return n - 1
}

func roundTo(value uint, roundTo uint) uint {
	return (value + (roundTo - 1)) &^ (roundTo - 1)
}

func checkResize(bs *[]byte, i uint) {
	if *bs == nil {
		*bs = make([]byte, 8)
		return
	}
	if i > uint(len(*bs)*8) {
		newbs := make([]byte, roundTo(i, 8))
		copy(newbs, *bs)
		*bs = newbs
	}
	return
}

func bitsetHas(bs *[]byte, i uint) bool {
	checkResize(bs, i)
	return (*bs)[i>>3]&(1<<(i&(7))) != 0
}

func bitsetSet(bs *[]byte, i uint) {
	checkResize(bs, i)
	(*bs)[i>>3] |= 1 << (i & (7))
}

func bitsetGet(bs *[]byte, i uint) uint {
	checkResize(bs, i)
	return uint((*bs)[i>>3] & (1 << (i & (7))))
}

func bitsetUnset(bs *[]byte, i uint) {
	checkResize(bs, i)
	(*bs)[i>>3] &^= 1 << (i & (7))
}

func bitsetStringer(bs *[]byte) string {
	// print binary value of bitset
	// var res string = "16" // set this to the "bit resolution" you'd like to see
	var res = strconv.Itoa(len(*bs))
	return fmt.Sprintf("%."+res+"b (%s bits)", bs, res)
}

// const (
// 	_ = iota
// 	typString
// 	typBytes
// )
//
// func encodeHdr(typ, size int) uint64 {
// 	return uint64((size << 32) | typ)
// }
//
// // decode decodes a header and returns the size and type
// func decodeHdr(hdr uint64) (int, int) {
// 	return int(hdr >> 32), int(hdr & 0xff)
// }
//
// type Raw []byte
//
// func String(v string) Raw {
// 	b := make([]byte, 8+len(v))
// 	bin.PutUint64(b[0:8], encodeHdr(typString, len(v)))
// 	copy(b[8:], v)
// 	return b
// }
//
// func Bytes(v []byte) Raw {
// 	b := make([]byte, 8+len(v))
// 	bin.PutUint64(b[0:8], encodeHdr(typBytes, len(v)))
// 	copy(b[8:], v)
// 	return b
// }
//
// func (r Raw) assertString() string {
// 	typ, size := decodeHdr(bin.Uint64(r[0:8]))
// 	if typ != typString {
// 		return fmt.Sprintf("%v", r[8:8+size])
// 	}
// 	return string(r[8 : 8+size])
// }
//
// func (r Raw) assertBytes() []byte {
// 	typ, size := decodeHdr(bin.Uint64(r[0:8]))
// 	if typ != typBytes {
// 		return []byte(fmt.Sprintf("%v", r[8:8+size]))
// 	}
// 	return r[8 : 8+size]
// }
//
// func (r Raw) String() string {
// 	return fmt.Sprintf(">> %#v", r)
//
// 	// typ, size := decodeHdr(bin.Uint64(r[0:8]))
// 	// switch typ {
// 	// case typString:
// 	// 	return r.assertString()
// 	// case typBytes:
// 	// 	return string(r.assertBytes())
// 	// }
// 	// return fmt.Sprintf("%+v", r[8:8+size])
// }
