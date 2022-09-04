package page

import (
	"fmt"
)

// Record flags
const (
	R_OVERFLOW = 0x0100

	RK_NUM = 0x0001
	RK_STR = 0x0002

	RV_NUM = 0x0010
	RV_STR = 0x0020

	R_NUM_NUM = RK_NUM | RV_NUM
	R_NUM_STR = RK_NUM | RV_STR
	R_STR_NUM = RK_STR | RV_NUM
	R_STR_STR = RK_STR | RV_STR
	R_PK_IDX  = R_NUM_NUM
	R_PK_DAT  = R_NUM_STR
)

// recordHeader is a pageHeader struct for encoding and
// decoding information for a record
type recordHeader struct {
	flags  uint16
	keyLen uint16
	valLen uint16
}

// record is a binary type
type record []byte

// newRecord initiates and returns a new record
func newRecord(flags uint16, key, val []byte) record {
	rsz := recordHeaderSize + len(key) + len(val)
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			flags:  flags,
			keyLen: uint16(len(key)),
			valLen: uint16(len(val)),
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+n:], val)
	return rec
}

func newUintUintRecord(key uint32, val uint32) record {
	rsz := recordHeaderSize + 4 + 4
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			flags:  RK_NUM | RV_NUM,
			keyLen: 4,
			valLen: 4,
		},
	)
	encU32(rec[recordHeaderSize:], key)
	encU32(rec[recordHeaderSize+4:recordHeaderSize+8], val)
	return rec
}

func newUintCharRecord(key uint32, val []byte) record {
	rsz := recordHeaderSize + 4 + len(val)
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			flags:  RK_NUM | RV_STR,
			keyLen: 4,
			valLen: uint16(len(val)),
		},
	)
	encU32(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+4:], val)
	return rec
}

func newCharUintRecord(key []byte, val uint32) record {
	rsz := recordHeaderSize + len(key) + 4
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			flags:  RK_STR | RV_NUM,
			keyLen: uint16(len(key)),
			valLen: 4,
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	encU32(rec[recordHeaderSize+n:], val)
	return rec
}

func newCharCharRecord(key []byte, val []byte) record {
	rsz := recordHeaderSize + len(key) + len(val)
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			flags:  RK_STR | RV_STR,
			keyLen: uint16(len(key)),
			valLen: uint16(len(val)),
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+n:], val)
	return rec
}

func (r *record) encRecordHeader(h *recordHeader) {
	_ = (*r)[recordHeaderSize] // early bounds check
	encU16((*r)[0:2], h.flags)
	encU16((*r)[2:4], h.keyLen)
	encU16((*r)[4:6], h.valLen)
}

func (r *record) decRecordHeader() *recordHeader {
	_ = (*r)[recordHeaderSize] // early bounds check
	return &recordHeader{
		flags:  decU16((*r)[0:2]),
		keyLen: decU16((*r)[2:4]),
		valLen: decU16((*r)[4:6]),
	}
}

func (r record) key() []byte {
	return r[recordHeaderSize : recordHeaderSize+decU16(r[2:4])]
}

func (r record) val() []byte {
	return r[recordHeaderSize+decU16(r[2:4]) : recordHeaderSize+decU16(r[2:4])+decU16(r[4:6])]
}

func (r record) String() string {
	rh := r.decRecordHeader()
	var k, v string
	k = fmt.Sprintf("%d", decU32(r.key()))
	v = fmt.Sprintf("%q", string(r.val()))
	// if rh.flags&t_uint_key == 0 {
	// 	k = fmt.Sprintf("key: %d", decU32(r.key()))
	// }
	// if rh.flags&t_char_key == 0 {
	// 	k = fmt.Sprintf("key: %s", string(r.key()))
	// }
	// if rh.flags&t_uint_val == 0 {
	// 	v = fmt.Sprintf("val: %d", decU32(r.val()))
	// }
	// if rh.flags&t_char_val == 0 {
	// 	v = fmt.Sprintf("val: %s", string(r.val()))
	// }
	return fmt.Sprintf("{ flags: %.4x, key: %s, val: %s }", rh.flags, k, v)
}
