package page

import (
	"errors"
	"fmt"
)

/*
 * Section containing types and methods for the `recordHeader` and `record`
 */

// https://go.dev/play/p/1CRP9LeYuiC
// --->>> https://go.dev/play/p/XUWtw4viTrF <<<---

const (

	// key and value masks
	R_KEY = 0xf0
	R_VAL = 0x0f

	// types for keys and values
	R_NUM = 0x11
	R_STR = 0x22
	R_PTR = 0x44

	// record flags
	R_NUM_NUM = (R_KEY | R_VAL) & R_NUM
	R_STR_STR = (R_KEY | R_VAL) & R_STR

	R_NUM_STR = (R_KEY & R_NUM) | (R_VAL & R_STR)
	R_STR_NUM = (R_KEY & R_STR) | (R_VAL & R_NUM)
	R_NUM_PTR = (R_KEY & R_NUM) | (R_VAL & R_PTR)
	R_STR_PTR = (R_KEY & R_STR) | (R_VAL & R_PTR)
)

func makeRecordFlags(kt, vt uint8) uint8 {
	mask := kt & R_KEY
	if mask != 0x10 && mask != 0x20 {
		panic("Record: bad key flag")
	}
	mask = vt & R_VAL
	if mask != 0x01 && mask != 0x02 && mask != 0x04 {
		panic("Record: bad val flag")
	}
	return (R_KEY & kt) | (R_VAL & vt)
}

var rFlags = []uint8{
	0x11, // number types for keys and number types for values
	0x12, // number types for keys and string types for values
	0x14, // number types for keys and pointer types for values
	0x21, // string types for keys and number types for values
	0x22, // string types for keys and string types for values
	0x24, // string types for keys and pointer types for values
}

func inSet(f uint8) bool {
	for _, v := range rFlags {
		if f == v {
			return true
		}
	}
	return false
}

var (
	ErrBadRecFlags  = errors.New("bad record flag option")
	ErrBadRecKeyLen = errors.New("bad record key length, max length is 255")
	ErrBadRecValLen = errors.New("bad record value length, max length is 65535")
)

func setHiBits(flag *uint8, t uint8) {
	*flag |= t << 4
}

func setLoBits(flag *uint8, t uint8) {
	*flag |= t
}

// recordHeader is a PageHeader struct for encoding and
// decoding information for a Record
type recordHeader struct {
	Flags  uint8
	KeyLen uint8
	ValLen uint16
}

// newRecordHeader constructs and returns a Record header using the provided flags
// along with the provided key and value data
func newRecordHeader(flags uint8, klen, vlen int) (*recordHeader, error) {
	if uint8(klen) > ^uint8(0) {
		return nil, ErrBadRecKeyLen
	}
	if uint16(vlen) > ^uint16(0) {
		return nil, ErrBadRecValLen
	}
	if !inSet(flags) {
		return nil, ErrBadRecFlags
	}
	return &recordHeader{
		Flags:  flags,
		KeyLen: uint8(klen),
		ValLen: uint16(vlen),
	}, nil
}

// size returns the size of a Record header
func (rh *recordHeader) size() int {
	return recordHeaderSize + int(rh.KeyLen) + int(rh.ValLen)
}

// Record is a binary type
type Record []byte

// NewRecord initiates and returns a new Record using the flags provided
// as the indicators of what types the keys and values should hold.
func NewRecord(kflag, vflag uint8, key, val []byte) Record {
	rh, err := newRecordHeader(makeRecordFlags(kflag, vflag), len(key), len(val))
	if err != nil {
		panic(err)
	}
	rec := make(Record, rh.size(), rh.size())
	rec.encRecordHeader(rh)
	n := copy(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+n:], val)
	return rec
}

// encRecordHeader takes a pointer to a recordHeader and encodes it
// directly into the Record as a []byte slice
func (r *Record) encRecordHeader(h *recordHeader) {
	_ = (*r)[recordHeaderSize] // early bounds check
	(*r)[0] = h.Flags
	(*r)[1] = h.KeyLen
	encU16((*r)[2:4], h.ValLen)
}

// decRecordHeader decodes the header from the Record and fills and
// returns a pointer to the recordHeader
func (r *Record) decRecordHeader() *recordHeader {
	_ = (*r)[recordHeaderSize] // early bounds check
	return &recordHeader{
		Flags:  (*r)[0],
		KeyLen: (*r)[1],
		ValLen: decU16((*r)[2:4]),
	}
}

// Flags returns the underlying uint16 representing the flags set for this record.
func (r *Record) Flags() uint8 {
	return (*r)[0]
}

func (r *Record) hasFlag(flag uint8) bool {
	return r.Flags()&flag != 0
}

// Key returns the underlying slice of bytes representing the record key
func (r *Record) Key() []byte {
	return (*r)[recordHeaderSize : recordHeaderSize+(*r)[1]]
}

// Val returns the underlying slice of bytes representing the record value
func (r *Record) Val() []byte {
	return (*r)[recordHeaderSize+(*r)[1] : uint16(recordHeaderSize+(*r)[1])+decU16((*r)[2:4])]
}

// KeyType returns the underlying type of the record key
func (r *Record) KeyType() uint8 {
	return r.Flags() & R_KEY
}

// ValType returns the underlying type of the record value
func (r *Record) ValType() uint8 {
	return r.Flags() & R_VAL
}

// String is the stringer method for a record
func (r *Record) String() string {
	fl := r.Flags()
	var k, v string
	if (fl & R_KEY) == (R_KEY & R_NUM) {
		k = fmt.Sprintf("key: %d", r.Key())
	}
	if (fl & R_KEY) == (R_KEY & R_STR) {
		k = fmt.Sprintf("key: %q", string(r.Key()))
	}
	if (fl&R_VAL) == (R_VAL&R_NUM) || (fl&R_VAL) == (R_KEY&R_PTR) {
		v = fmt.Sprintf("val: %d", decU32(r.Val()))
	}
	if (fl & R_VAL) == (R_KEY & R_STR) {
		v = fmt.Sprintf("val: %q", string(r.Val()))
	}
	return fmt.Sprintf("{ flags: %.4x, key: %s, val: %s }", fl, k, v)
}

func EncodeRecordID(rid *RecordID) uint64 {
	return encodeRID(rid.PageID, rid.CellID)
}

func DecodeRecordID(rid uint64) *RecordID {
	pid, cid := decodeRID(rid)
	return &RecordID{
		PageID: pid,
		CellID: cid,
	}
}

func encodeRID(pid uint32, cid uint16) uint64 {
	var n uint64
	n |= uint64(pid) << 32
	n |= uint64(cid)
	return n
}

func decodeRID(rid uint64) (uint32, uint16) {
	pid := uint32(rid >> 32)
	cid := uint16(rid)
	return pid, cid
}
