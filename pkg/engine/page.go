package engine

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

// Binary encoding and decoding helpers
var (
	encU16 = binary.LittleEndian.PutUint16
	encU32 = binary.LittleEndian.PutUint32
	encU64 = binary.LittleEndian.PutUint64
	decU16 = binary.LittleEndian.Uint16
	decU32 = binary.LittleEndian.Uint32
	decU64 = binary.LittleEndian.Uint64
)

type RecordID struct {
	PageID uint32
	CellID uint16
}

func (r *RecordID) String() string {
	return fmt.Sprintf("{%d, %d}", r.PageID, r.CellID)
}

const (
	// P_FREE and P_USED are flags to be used with the page
	P_FREE     uint32 = 0x00000001 // indicates the Page is free to use
	P_USED     uint32 = 0x00000002 // indicates the Page is being used
	P_SORTED   uint32 = 0x00000010 // indicates the Page cells are sorted
	P_UNSORTED uint32 = 0x00000020

	// constants for the page, headers, cellptrs and record sizes
	pageSize         = 16 << 10
	pageHeaderSize   = 24
	pageCellPtrSize  = 8
	recordHeaderSize = 6

	// offsets to be used for decoding and encoding the page header
	offPID      uint32 = 0  // PID=uint32		offs=0-4 	(4 bytes)
	offPrev     uint32 = 4  // prev=uint32 		offs=4-8	(4 bytes)
	offNext     uint32 = 8  // next=uint32 		offs=8-12	(4 bytes)
	offFlags    uint32 = 12 // flags=uint32		offs=12-16	(4 bytes)
	offNumCells uint16 = 16 // numCells=uint16	offs=16-18	(2 bytes)
	offNumFree  uint16 = 18 // numFree=uint16	offs=18-20	(2 bytes)
	offLower    uint16 = 20 // lower=uint16		offs=20-22 	(2 bytes)
	offUpper    uint16 = 22 // upper=uint16		offs=22-24	(2 bytes)
)

type PageHeader struct {
	ID    uint32 // id of Page
	Prev  uint32 // prev Page pointer
	Next  uint32 // next Page pointer
	Flags uint32 // flags and meta data for Page
	Cells uint16 // number of cells allocated
	Free  uint16 // number of cells that are free
	Lower uint16 // lower numFree space bound
	Upper uint16 // upper numFree space bound
}

func (h *PageHeader) Size() int {
	return int(unsafe.Sizeof(*h))
}

func (h *PageHeader) String(p *Page) string {
	ss := fmt.Sprintf("%10v +---------+\n", "")
	ss += fmt.Sprintf("%10v |%2v%v%-2v|\n", "", "", " Page", "")
	ss += fmt.Sprintf("%10v +---------+\n", "")
	hv := reflect.ValueOf(h).Elem()
	for i := 0; i < hv.NumField(); i++ {
		sf := hv.Type().Field(i)
		ss += fmt.Sprintf("%10v |%2v%5d%-2v|\n", strings.ToLower(sf.Name), "", hv.Field(i), "")
		// ss += fmt.Sprintf("%10v |%-8d|\n", strings.ToLower(sf.Name), hv.Field(i))
		// ss += fmt.Sprintf("%16s\t%8s\t (index=%d, size=%d, align=%d, offset=%d)\n",
		//	sf.Name, sf.Type.Name(), sf.Index[0], sf.Type.Size(), sf.Type.Align(), sf.Offset)
	}
	ss += fmt.Sprintf("%10v +---------+\n", "")
	for i := uint16(0); i < h.Cells; i++ {
		c := p.decCell(i)
		var t = "T"
		if !c.hasFlag(C_USED) {
			t = "F"
		}
		ss += fmt.Sprintf("%10v |%2v%v%-2v|\n", "", "", fmt.Sprintf("c%d,u%1s", c.getID(), t), "")
	}
	ss += fmt.Sprintf("%10v +---------+ %v %d\n", "", "(lo)", h.Lower)
	ss += fmt.Sprintf("%10v |%2v%s%3v|\n", "", "", "Free", "")
	ss += fmt.Sprintf("%10v +---------+ %v %d\n", "", "(hi)", h.Upper)
	var cc []cellptr
	for i := uint16(0); i < h.Cells; i++ {
		cc = append(cc, p.decCell(i))
	}
	sort.Slice(
		cc, func(i, j int) bool {
			return cc[i].getID() < cc[j].getID()
		},
	)
	for i := len(cc) - 1; i >= 0; i-- {
		if cc[i].hasFlag(C_USED) {
			r := p.getRecordUsingCell(cc[i])
			ss += fmt.Sprintf("%10v |%2v%s%3v|\n", "", "", r.Key(), "")
		} else {
			ss += fmt.Sprintf("%10v |%2v%s%3v|\n", "", "", "____", "")
		}
	}
	ss += fmt.Sprintf("%10v +---------+ %10v\n", "", "")
	return ss
}

type Page []byte

func NewPage(id uint32) Page {
	p := make(Page, pageSize, pageSize)
	p.setPageHeader(
		&PageHeader{
			ID:    id,
			Prev:  0,
			Next:  0,
			Flags: P_USED,
			Cells: 0,
			Free:  0,
			Lower: pageHeaderSize,
			Upper: pageSize,
		},
	)
	return p
}

func (p *Page) getPageHeader() *PageHeader {
	return &PageHeader{
		ID:    decU32((*p)[offPID : offPID+4]),
		Prev:  decU32((*p)[offPrev : offPrev+4]),
		Next:  decU32((*p)[offNext : offNext+4]),
		Flags: decU32((*p)[offFlags : offFlags+4]),
		Cells: decU16((*p)[offNumCells : offNumCells+2]),
		Free:  decU16((*p)[offNumFree : offNumFree+2]),
		Lower: decU16((*p)[offLower : offLower+2]),
		Upper: decU16((*p)[offUpper : offUpper+2]),
	}
}

func (p *Page) setPageHeader(h *PageHeader) {
	encU32((*p)[offPID:offPID+4], h.ID)
	encU32((*p)[offPrev:offPrev+4], h.Prev)
	encU32((*p)[offNext:offNext+4], h.Next)
	encU32((*p)[offFlags:offFlags+4], h.Flags)
	encU16((*p)[offNumCells:offNumCells+2], h.Cells)
	encU16((*p)[offNumFree:offNumFree+2], h.Free)
	encU16((*p)[offLower:offLower+2], h.Lower)
	encU16((*p)[offUpper:offUpper+2], h.Upper)
}

func (p *Page) getRecordUsingCellPos(pos uint16) Record {
	cp := p.decCell(pos)
	beg, end := cp.getBounds()
	return Record((*p)[beg:end])
}

func (p *Page) getRecordUsingCell(c cellptr) Record {
	beg, end := c.getBounds()
	return Record((*p)[beg:end])
}

func (p *Page) addCell(size uint16) cellptr {
	p.incrNumCells(1)
	p.incrLower(pageCellPtrSize)
	p.decrUpper(size)
	c := newCell(p.getNumCells(), p.getUpper(), size)
	p.encCell(c, p.getNumCells()-1)
	return c
}

func (p *Page) delCell(c cellptr, pos uint16) {
	// mark the cell free
	c.setFlags(C_FREE)
	// get the "window" of all the cells
	cells := (*p)[pageHeaderSize:p.getLower()]
	// slide all the cells in the window down
	copy(cells[pos*8:], cells[(pos+1)*8:])
	// re-encode the "deleted" cell, at the front of the list
	p.encCell(c, p.getNumCells()-1)
	// And increment the free cell count in the page header.
	p.incrNumFree(1)
}

func (p *Page) swapCells(pos1, pos2 uint16) {
	cp1 := p.decCell(pos1)
	cp2 := p.decCell(pos2)
	p.encCell(cp1, pos2)
	p.encCell(cp2, pos1)
}

func (p *Page) encCell(c cellptr, pos uint16) {
	off := pageHeaderSize + (pos * pageCellPtrSize)
	if off < pageHeaderSize || off > p.getLower() {
		panic("error: cell position out of bounds")
	}
	if !c.isValid() {
		panic("error: cell is not a valid cell")
	}
	encU64((*p)[off:off+8], uint64(c))
}

func (p *Page) decCell(pos uint16) cellptr {
	off := pageHeaderSize + (pos * pageCellPtrSize)
	if off < pageHeaderSize || off > p.getLower() {
		panic("error: cell position out of bounds")
	}
	c := cellptr(decU64((*p)[off : off+8]))
	if !c.isValid() {
		panic("error: cell is not a valid cell")
	}
	return c
}

func (p *Page) makeRecordID(c cellptr) *RecordID {
	return &RecordID{
		PageID: p.getPageID(),
		CellID: c.getID(),
	}
}

// recycleCell attempts to reuse a free cellptr for a record, if there is a candidate that
// works well. It returns the used cellptr, and a boolean indicating true if it suceeded in
// recycling the cellptr, and false if it could not recycle one.
func (p *Page) recycleCell(freeCells, numCells uint16, r Record) cellptr {
	// We do, so let's see if we have any candidates for recycling.
	for pos := numCells - 1; pos > numCells-freeCells-1; pos-- {
		// Get the free cell at the first location
		c := p.decCell(pos)
		// Check the cell to see if it's a candidate
		if c.hasFlag(C_FREE) && c.canFit(uint16(len(r))) {
			// We have found ourselves a candidate, now we just need to
			// update the cell, and re-encode it.
			c.setFlags(C_USED)
			c.setLength(uint16(len(r)))
			p.encCell(c, pos)
			// And decrement the free cell count in the page header.
			p.decrNumFree(1)
			// Now, we can return the cell
			return c
		}
	}
	// Otherwise, we could not find one to recycle, so we return 0 because we will be
	// checking this cell to ensure that it reads as a valid one, and 0 will not so
	// it is perfectly safe to return.
	return 0
}

func (p *Page) AddRecord(r Record) *RecordID {
	// Get our free cells, and our total cell count.
	freeCells, numCells := p.getNumFree(), p.getNumCells()
	// Allocate our cell pointer, we will need to use one no matter what.
	var cp cellptr
	// Before continuing, we must check to see if we can re-use any cells.
	if freeCells > 0 {
		// We do, so let's see if we have any candidates for recycling.
		cp = p.recycleCell(freeCells, numCells, r)
		// We will be checking below to see if the cell pointer is valid,
		// and it will only be valid if we had a successful time recycling
		// in here, so no need to do anything else, just proceed.
	}
	// This check basically tells us if we have recycled a cell or not.
	if !cp.isValid() {
		// No valid cell pointer found, which means we did not recycle any,
		// and we are free to allocate a fresh one. So that is what we do.
		cp = p.addCell(uint16(len(r)))
	}
	// We want to ensure we write the record data to the page before we
	// check or try to sort.
	copy((*p)[cp.getOffset():cp.getOffset()+cp.getLength()], r)
	// We will check to see if we need to sort, and if so, we will. Checking
	// to see if we need to sort is always faster (if a sort does not need
	// to be done) than performing the actual sort.
	if !sort.IsSorted(p) {
		// Use sort.Sort, as opposed to sort.Stable when possible, unless you
		// are absolutely requiring a stable sort. The sort.Stable version
		// makes one call to data.Len to determine n, O(n*log(n)) calls to
		// data.Less and O(n*log(n)*log(n)) calls to data.Swap. In contrast,
		// sort.Sort also makes one call to data.Len to determine n, but then
		// only O(n*log(n)) calls to data.Less and data.Swap. If your data set
		// is short, sort.Stable would be very close to as fast, but again,
		// you should not use it unless it is necessary.
		sort.Sort(p)
	}
	// And finally, return our RecordID
	return &RecordID{p.getPageID(), cp.getID()}
}

func (p *Page) GetRecord(id *RecordID) *Record {
	// First, we will attempt to locate the record.
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		c := p.decCell(pos)
		if c.getID() == id.CellID {
			// We have located the record. Let's check to make sure it has
			// not been deleted.
			if c.hasFlag(C_FREE) {
				return nil
			}
			// It has not, so we can fetch the record.
			r := p.getRecordUsingCell(c)
			// We should make a copy of it, so we do not mutate the original.
			rc := make(Record, len(r), len(r))
			copy(rc, r)
			return &rc
		}
	}
	// Otherwise, we did not locate the record
	return nil
}

func (p *Page) GetRecordByKey(key []byte) *Record {
	// First, we will attempt to locate the record.
	pos, _ := p.FindCellPos(key)
	// Otherwise, we have located it. Let's check to make sure it has
	// not been deleted.
	c := p.decCell(uint16(pos))
	if c.hasFlag(C_FREE) {
		return nil
	}
	// It has not, so we will get it.
	r := p.getRecordUsingCell(c)
	// We should make a copy of it, so we do not mutate the original.
	rc := make(Record, len(r), len(r))
	copy(rc, r)
	return &rc
}

func (p *Page) DelRecord(id *RecordID) bool {
	// First, we will attempt to locate the record.
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		// Check the cell at the provided position
		c := p.decCell(pos)
		if c.getID() == id.CellID {
			// We have located the record. First overwrite the record data.
			beg, end := c.getBounds()
			copy((*p)[beg:end], make([]byte, c.getLength()))
			// Then, we must free the cell pointer.
			p.delCell(c, pos)
			return true
		}
	}
	// Otherwise, we did not locate the record
	return false
}

var SkipRecord = errors.New("skip this record")

func (p *Page) RangeRecords(fn func(r *Record) error) error {
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		c := p.decCell(pos)
		if c.hasFlag(C_FREE) {
			continue
		}
		r := p.getRecordUsingCell(c)
		if err := fn(&r); err != nil {
			if err == SkipRecord {
				continue
			}
			return err
		}
	}
	return nil
}

func (p *Page) Len() int {
	return int(p.getNumCells() - p.getNumFree())
}

func (p *Page) Less(i, j int) bool {
	r1 := p.getRecordUsingCellPos(uint16(i))
	r2 := p.getRecordUsingCellPos(uint16(j))
	return bytes.Compare(r1.Key(), r2.Key()) < 0
}

func (p *Page) Swap(i, j int) {
	cp1 := p.decCell(uint16(i))
	cp2 := p.decCell(uint16(j))
	p.encCell(cp1, uint16(j))
	p.encCell(cp2, uint16(i))
}

func (p *Page) FindCellPos(k []byte) (uint16, bool) {
	cmp := func(key []byte, pos uint16) int {
		r := p.getRecordUsingCellPos(pos)
		return bytes.Compare(key, r.Key())
	}
	// The invariants here are similar to the ones in Search.
	// Define cmp(-1) > 0 and cmp(n) <= 0
	// Invariant: cmp(i-1) > 0, cmp(j) <= 0
	var at int
	n := int(p.getNumCells())
	i, j := 0, n
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if at = cmp(k, uint16(h)); at > 0 {
			i = h + 1 // preserves cmp(i-1) > 0
		} else {
			j = h // preserves cmp(j) <= 0
		}
	}
	// i == j, cmp(i-1) > 0 and cmp(j) <= 0
	return uint16(i), i < n && at == 0
}

func (p *Page) String() string {
	return p.getPageHeader().String(p)
}

// getPageID decodes and returns the Page ID directly from the encoded pageHeader.
func (p *Page) getPageID() uint32 {
	return decU32((*p)[offPID : offPID+4])
}

// getPrev decodes and returns the previous pointer directly from the encoded pageHeader.
func (p *Page) getPrev() uint32 {
	return decU32((*p)[offPrev : offPrev+4])
}

// getNext decodes and returns the next pointer directly from the encoded pageHeader.
func (p *Page) getNext() uint32 {
	return decU32((*p)[offNext : offNext+4])
}

// getFlags decodes and returns the flags field directly from the encoded pageHeader.
func (p *Page) getFlags() uint32 {
	return decU32((*p)[offFlags : offFlags+4])
}

// getNumFree decodes and returns the number of free cells directly from the encoded pageHeader.
func (p *Page) getNumFree() uint16 {
	return decU16((*p)[offNumFree : offNumFree+2])
}

// getNumCells decodes and returns the number of allocated cells directly from the encoded pageHeader.
func (p *Page) getNumCells() uint16 {
	return decU16((*p)[offNumCells : offNumCells+2])
}

// getLower decodes and returns the lower bound marker directly from the encoded pageHeader.
func (p *Page) getLower() uint16 {
	return decU16((*p)[offLower : offLower+2])
}

// getUpper decodes and returns the upper bound marker directly from the encoded pageHeader.
func (p *Page) getUpper() uint16 {
	return decU16((*p)[offUpper : offUpper+2])
}

// setPageID encodes the provided value directly into the pageHeader.
func (p *Page) setPageID(n uint32) {
	encU32((*p)[offPID:offPID+4], n)
}

// setPrev encodes the provided value directly into the pageHeader.
func (p *Page) setPrev(n uint32) {
	encU32((*p)[offPrev:offPrev+4], n)
}

// setNext encodes the provided value directly into the pageHeader.
func (p *Page) setNext(n uint32) {
	encU32((*p)[offNext:offNext+4], n)
}

// setFlags encodes the provided value directly into the pageHeader.
func (p *Page) setFlags(n uint32) {
	encU32((*p)[offFlags:offFlags+4], n)
}

// setNumFree encodes the provided value directly into the pageHeader.
func (p *Page) setNumFree(n uint16) {
	encU16((*p)[offNumFree:offNumFree+2], n)
}

func (p *Page) incrNumFree(n uint16) {
	encU16((*p)[offNumFree:offNumFree+2], decU16((*p)[offNumFree:offNumFree+2])+n)
}

func (p *Page) decrNumFree(n uint16) {
	encU16((*p)[offNumFree:offNumFree+2], decU16((*p)[offNumFree:offNumFree+2])-n)
}

// setNumCells encodes the provided value directly into the pageHeader.
func (p *Page) setNumCells(n uint16) {
	encU16((*p)[offNumCells:offNumCells+2], n)
}

func (p *Page) incrNumCells(n uint16) {
	encU16((*p)[offNumCells:offNumCells+2], decU16((*p)[offNumCells:offNumCells+2])+n)
}

func (p *Page) decrNumCells(n uint16) {
	encU16((*p)[offNumCells:offNumCells+2], decU16((*p)[offNumCells:offNumCells+2])-n)
}

// setLower encodes the provided value directly into the pageHeader.
func (p *Page) setLower(n uint16) {
	encU16((*p)[offLower:offLower+2], n)
}

func (p *Page) incrLower(n uint16) {
	encU16((*p)[offLower:offLower+2], decU16((*p)[offLower:offLower+2])+n)
}

func (p *Page) decrLower(n uint16) {
	encU16((*p)[offLower:offLower+2], decU16((*p)[offLower:offLower+2])-n)
}

// setUpper encodes the provided value directly into the pageHeader.
func (p *Page) setUpper(n uint16) {
	encU16((*p)[offUpper:offUpper+2], n)
}

func (p *Page) incrUpper(n uint16) {
	encU16((*p)[offUpper:offUpper+2], decU16((*p)[offUpper:offUpper+2])+n)
}

func (p *Page) decrUpper(n uint16) {
	encU16((*p)[offUpper:offUpper+2], decU16((*p)[offUpper:offUpper+2])-n)
}

const (
	// C_FREE C_USED and C_MAGK are flags for the cellptrs
	C_FREE uint16 = 0x0001
	C_USED uint16 = 0x0002
	C_MAGK uint16 = 0x55a0

	// constants for the cellptrs
	u64mask0to2 = 0x000000000000ffff
	u64mask2to4 = 0x00000000ffff0000
	u64mask4to6 = 0x0000ffff00000000
	u64mask6to8 = 0xffff000000000000
	shift2B     = 16
	shift4B     = 32
	shift6B     = 48
)

// cellptr is a record cell pointer in the slotted page structure
type cellptr uint64

// newCell takes an id, a record offset and a record length and returns
// a new cell pointer
func newCell(id, offset, length uint16) cellptr {
	var c cellptr
	c |= cellptr(id)
	c |= cellptr(C_MAGK|C_USED) << shift2B
	c |= cellptr(offset) << shift4B
	c |= cellptr(length) << shift6B
	return c
}

// canFit takes a record length and returns a boolean indicating true if
// the record could fit in the current offset. It is used when attempting
// to recycle a cell pointer.
func (c *cellptr) canFit(length uint16) bool {
	return uint16(*c>>shift6B) <= length
}

// isValid returns true if the cell pointer is valid (contains the magic flag)
func (c *cellptr) isValid() bool {
	return uint16(*c>>shift2B)&C_MAGK != 0
}

// getBounds returns the beginning and ending bounds of the record based
// on the stored record offset and length. It is considered a helper method.
func (c *cellptr) getBounds() (uint16, uint16) {
	offset := uint16(*c >> shift4B)
	length := uint16(*c >> shift6B)
	return offset, offset + length
}

// getID returns the id field of the cell pointer
func (c *cellptr) getID() uint16 {
	return uint16(*c)
}

// getFlags returns the flags field of the cell pointer
func (c *cellptr) getFlags() uint16 {
	return uint16(*c >> shift2B)
}

// hasFlag tests if the cell pointer has a flag set
func (c *cellptr) hasFlag(flag uint16) bool {
	return uint16(*c>>shift2B)&flag != 0
}

// getOffset returns the record offset stored in the cell pointer
func (c *cellptr) getOffset() uint16 {
	return uint16(*c >> shift4B)
}

// getLength returns the record length stored in the cell pointer
func (c *cellptr) getLength() uint16 {
	return uint16(*c >> shift6B)
}

// setID takes sets the cell pointer id field to the value provided.
func (c *cellptr) setID(n uint16) {
	*c &^= u64mask0to2
	*c |= cellptr(n)
}

// setID takes sets the cell pointer flags field to the value provided.
func (c *cellptr) setFlags(n uint16) {
	*c &^= u64mask2to4
	*c |= cellptr(C_MAGK|n) << shift2B
}

// setOffset takes sets the cell pointer record offset field to the value provided.
func (c *cellptr) setOffset(n uint16) {
	*c &^= u64mask4to6
	*c |= cellptr(n) << shift4B
}

// setLength takes sets the cell pointer record length field to the value provided.
func (c *cellptr) setLength(n uint16) {
	*c &^= u64mask6to8
	*c |= cellptr(n) << shift6B
}

// String is the stringer method for a cell pointer.
func (c *cellptr) String() string {
	return fmt.Sprintf(
		"id=%d, flags=0x%.4x, offset=%d, length=%d", c.getID(), c.getFlags(), c.getOffset(),
		c.getLength(),
	)
}

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

	KeyMask = 0x000f
	ValMask = 0x00f0
)

// RecordHeader is a pageHeader struct for encoding and
// decoding information for a record
type RecordHeader struct {
	Flags  uint16
	KeyLen uint16
	ValLen uint16
}

// Record is a binary type
type Record []byte

// NewRecord initiates and returns a new record using the flags provided
// as the indicators of what types the keys and values should hold.
func NewRecord(flags uint16, key, val []byte) Record {
	rsz := recordHeaderSize + len(key) + len(val)
	rec := make(Record, rsz, rsz)
	rec.encRecordHeader(
		&RecordHeader{
			Flags:  flags,
			KeyLen: uint16(len(key)),
			ValLen: uint16(len(val)),
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+n:], val)
	return rec
}

// NewUintUintRecord creates and returns a new Record that uses uint32's
// as keys and uint32's as values.
func NewUintUintRecord(key uint32, val uint32) Record {
	rsz := recordHeaderSize + 4 + 4
	rec := make(Record, rsz, rsz)
	rec.encRecordHeader(
		&RecordHeader{
			Flags:  RK_NUM | RV_NUM,
			KeyLen: 4,
			ValLen: 4,
		},
	)
	encU32(rec[recordHeaderSize:], key)
	encU32(rec[recordHeaderSize+4:recordHeaderSize+8], val)
	return rec
}

// NewUintCharRecord creates and returns a new Record that uses uint32's
// as keys and []byte slices as values.
func NewUintCharRecord(key uint32, val []byte) Record {
	rsz := recordHeaderSize + 4 + len(val)
	rec := make(Record, rsz, rsz)
	rec.encRecordHeader(
		&RecordHeader{
			Flags:  RK_NUM | RV_STR,
			KeyLen: 4,
			ValLen: uint16(len(val)),
		},
	)
	encU32(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+4:], val)
	return rec
}

// NewCharUintRecord creates and returns a new Record that uses []byte
// slices as keys and uint32's as values.
func NewCharUintRecord(key []byte, val uint32) Record {
	rsz := recordHeaderSize + len(key) + 4
	rec := make(Record, rsz, rsz)
	rec.encRecordHeader(
		&RecordHeader{
			Flags:  RK_STR | RV_NUM,
			KeyLen: uint16(len(key)),
			ValLen: 4,
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	encU32(rec[recordHeaderSize+n:], val)
	return rec
}

// NewCharCharRecord creates and returns a new Record that uses []byte
// slices as keys and []byte slices as values.
func NewCharCharRecord(key []byte, val []byte) Record {
	rsz := recordHeaderSize + len(key) + len(val)
	rec := make(Record, rsz, rsz)
	rec.encRecordHeader(
		&RecordHeader{
			Flags:  RK_STR | RV_STR,
			KeyLen: uint16(len(key)),
			ValLen: uint16(len(val)),
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+n:], val)
	return rec
}

// encRecordHeader takes a pointer to a RecordHeader and encodes it
// directly into the record as a []byte slice
func (r *Record) encRecordHeader(h *RecordHeader) {
	_ = (*r)[recordHeaderSize] // early bounds check
	encU16((*r)[0:2], h.Flags)
	encU16((*r)[2:4], h.KeyLen)
	encU16((*r)[4:6], h.ValLen)
}

// decRecordHeader decodes the header from the Record and fills and
// returns a pointer to the RecordHeader
func (r *Record) decRecordHeader() *RecordHeader {
	_ = (*r)[recordHeaderSize] // early bounds check
	return &RecordHeader{
		Flags:  decU16((*r)[0:2]),
		KeyLen: decU16((*r)[2:4]),
		ValLen: decU16((*r)[4:6]),
	}
}

// Flags returns the underlying uint16 representing the flags set for this Record.
func (r *Record) Flags() uint16 {
	return decU16((*r)[0:2])
}

// Key returns the underlying slice of bytes representing the Record key
func (r *Record) Key() []byte {
	return (*r)[recordHeaderSize : recordHeaderSize+decU16((*r)[2:4])]
}

// Val returns the underlying slice of bytes representing the Record value
func (r *Record) Val() []byte {
	return (*r)[recordHeaderSize+decU16((*r)[2:4]) : recordHeaderSize+decU16((*r)[2:4])+decU16((*r)[4:6])]
}

// KeyType returns the underlying type of the Record key
func (r *Record) KeyType() uint16 {
	return r.Flags() & KeyMask
}

// ValType returns the underlying type of the Record value
func (r *Record) ValType() uint16 {
	return r.Flags() & ValMask
}

// String is the stringer method for a Record
func (r *Record) String() string {
	fl := r.Flags()
	var k, v string
	if (fl & KeyMask) == RK_NUM {
		k = fmt.Sprintf("key: %d", decU32(r.Key()))
	}
	if (fl & KeyMask) == RK_STR {
		k = fmt.Sprintf("key: %q", string(r.Key()))
	}
	if (fl & ValMask) == RV_NUM {
		v = fmt.Sprintf("val: %d", decU32(r.Val()))
	}
	if (fl & ValMask) == RV_STR {
		v = fmt.Sprintf("val: %q", string(r.Val()))
	}
	return fmt.Sprintf("{ flags: %.4x, key: %s, val: %s }", fl, k, v)
}
