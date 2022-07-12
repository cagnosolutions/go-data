package pager

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"text/tabwriter"
)

// https://go.dev/play/p/gr8RC8vDuSv

var (
	ErrRecordTooSmall = errors.New("record is too small (under min fileSize allowed)")
	ErrRecordTooBig   = errors.New("record is too big (over max fileSize allowed)")
	ErrNoRoom         = errors.New("the page is full, or has too much fragmentation")
	ErrEmptyPage      = errors.New("the page is empty, cannot get any information")
	ErrInvalidPID     = errors.New("page ID is not valid, or does not match")
	ErrInvalidSID     = errors.New("slot ID is not valid (out of bounds perhaps?)")
	ErrRecordNotFound = errors.New("record not found, it may have been deleted")
)

// Available flag values
const (
	_ uint16 = 0x0004
	_ uint16 = 0x0008
	_ uint16 = 0x0010
	_ uint16 = 0x0020
	_ uint16 = 0x0040
	_ uint16 = 0x0080
	_ uint16 = 0x0100
	_ uint16 = 0x0200
	_ uint16 = 0x0400
	_ uint16 = 0x0800
	_ uint16 = 0x1000
	_ uint16 = 0x2000
	_ uint16 = 0x4000
	_ uint16 = 0x8000
)

// Defaults for page size
const (
	DefaultPageSize        = szPg  // 4KB
	MinPageSize     uint16 = 512   // 512B
	MaxPageSize     uint16 = 65535 // 64KB
)

// Sizes for header, page and records
const (
	szMinRec = 8       // minimum record size
	szMaxRec = 2048    // maximum record size
	szHd     = 16      // fileSize of page header (in bytes)
	szSl     = 6       // fileSize of slot index (in bytes)
	szPg     = 4 << 10 // fileSize of page (default)
)

// Binary offsets for page header
const (
	offPID   uint32 = 0  // pid=uint32		offs=0-4
	offMeta  uint32 = 4  // meta=uint32	offs=4-8
	offStat  uint16 = 8  // status=uint16	offs=8-10
	offSlots uint16 = 10 // slots=uint16	offs=10-12
	offLower uint16 = 12 // lower=uint16	offs=12-14
	offUpper uint16 = 14 // upper=uint16	offs=14-16
)

// Status flags (page or slot status)
const (
	stFree uint16 = 0x0001 // page or slot is free
	stUsed uint16 = 0x0002 // page or slot to use
)

// Meta flags for page
const (
	// Record type
	mdRecFixed uint32 = 0x000004 // fixed sized records
	mdRecDynmc uint32 = 0x000008 // dynamic sized records

	// Slotted page type
	mdSlotted uint32 = 0x000010 // default, general purpose slotted page
	_         uint32 = 0x000020
	_         uint32 = 0x000040
	_         uint32 = 0x000080

	// temp disable
	_ = mdRecFixed
)

// page latch
var pgLatch sync.Mutex

// bin is just a little shorthand if you wish to easily change
// up the encoding and decoding byte order, this variable can
// simply be changed.
var bin = binary.LittleEndian

// slot is a struct which is an index for a record.
type slot struct {
	status uint16
	offset uint16
	length uint16
}

// bounds returns the starting and ending offsets to the
// particular record that this slot index "points" to.
func (s slot) bounds() (uint16, uint16) {
	return s.offset, s.offset + s.length
}

func (s slot) String() string {
	ss := fmt.Sprintf("off=%.4d, len=%.4d, free=%v", s.offset, s.length, s.status == stFree)
	ss += fmt.Sprintf("\t[0x%.4x,0x%.4x,0x%.4x]", s.offset, s.length, s.status)

	return ss
}

// recID is a struct representing a record ID.
type recID struct {
	pid uint32 // page id
	sid uint16 // slot id (slot index)
}

// header is a struct representing a page header.
type header struct {
	pid    uint32 // id of page
	meta   uint32 // meta data for page
	status uint16 // status of page
	slots  uint16 // number of slots
	lower  uint16 // lower free space bound
	upper  uint16 // upper free space bound
}

// page is a page.
type page []byte

func (p *page) size() int {
	if p == nil {
		return 0
	}
	return len(*p)
}

// newEmptyPage returns a new page instance set with the provided page ID,
// with a meta byte status of stFree, denoting it as empty and free to use.
func newEmptyPage(pid uint32) page {
	return newEmptyPageSize(pid, DefaultPageSize)
}

// newEmptyPageSize returns a new page instance set with the provided page ID, with
// a meta byte status of stFree, denoting it as empty and free to use. It is sized
// according to the provided size.
func newEmptyPageSize(pid uint32, size uint16) page {
	pg := make(page, size, size)
	pg.setHeader(
		&header{
			pid:    pid,
			meta:   mdSlotted | mdRecDynmc,
			status: stFree,
			slots:  0,
			lower:  szHd,
			upper:  size,
		},
	)
	return pg
}

// newPage returns a new page instance set with the provided page ID.
func newPage(pid uint32) page {
	return newPageSize(pid, DefaultPageSize)
}

// newPageSize returns a new page instance set with the provided page ID, and sized
// according to the provided size.
func newPageSize(pid uint32, size uint16) page {
	pg := make(page, size, size)
	pg.setHeader(
		&header{
			pid:    pid,
			meta:   mdSlotted | mdRecDynmc,
			status: stUsed,
			slots:  0,
			lower:  szHd,
			upper:  size,
		},
	)
	return pg
}

// setHeader encodes the provided header structure to the underlying
// page.
func (p *page) setHeader(h *header) {
	bin.PutUint32((*p)[offPID:offPID+4], h.pid)
	bin.PutUint32((*p)[offMeta:offMeta+4], h.meta)
	bin.PutUint16((*p)[offStat:offStat+2], h.status)
	bin.PutUint16((*p)[offSlots:offSlots+2], h.slots)
	bin.PutUint16((*p)[offLower:offLower+2], h.lower)
	bin.PutUint16((*p)[offUpper:offUpper+2], h.upper)
}

// getHeader decodes (from the underlying page) and returns a
// pointer to a header structure
func (p *page) getHeader() *header {
	return &header{
		pid:    bin.Uint32((*p)[offPID:]),
		meta:   bin.Uint32((*p)[offMeta:]),
		status: bin.Uint16((*p)[offStat:]),
		slots:  bin.Uint16((*p)[offSlots:]),
		lower:  bin.Uint16((*p)[offLower:]),
		upper:  bin.Uint16((*p)[offUpper:]),
	}
}

// getPageID returns the current page ID.
func (p *page) getPageID() uint32 {
	return bin.Uint32((*p)[offPID:])
}

// freeSpace returns the amount of contiguous free space left in the
// page (space between the upper and lower bounds)
func (p *page) freeSpace() uint16 {
	return bin.Uint16((*p)[offUpper:]) - bin.Uint16((*p)[offLower:])
}

// checkRecord performs sanity and error checking on a record size
func (p *page) checkRecord(size uint16) error {
	if size < szMinRec {
		return ErrRecordTooSmall
	}
	if size > szMaxRec {
		return ErrRecordTooBig
	}
	if size >= p.freeSpace()-szSl {
		return ErrNoRoom
	}
	return nil
}

// getSlotSet decodes and returns a set of slot pointer for this page.
// It will return nil if there are no slots on this page. Any changes
// made to this list of slots is not persisted until calling setSlots.
func (p *page) getSlotSet() []*slot {
	// Check if there are any slots to return.
	slotCount := bin.Uint16((*p)[offSlots:])
	if slotCount < 1 {
		// No slots to decode
		return nil
	}
	// We have slots we can decode. Create a set we can append to.
	var slots []*slot
	// Start looping, decoding, and adding slots to our slot set.
	for sid := uint16(0); sid < slotCount; sid++ {
		// Append the slot to the slot set.
		slots = append(slots, p._getSlot(sid))
	}
	// Finally, return our slot set.
	return slots
}

// setSlotSet encodes a set of slot pointers into this page. It will
// return an error if there is not enough room to write the set of slots
// to the underlying page.
func (p *page) setSlotSet(_ []*slot) error {
	// Not sure if I want to append or overwrite this at the moment...
	return nil
}

// _setSlot writes the provided slot to the location derived using
// the supplied slot index ID. _setSlot panics if the provided sid
// is out of bounds.
func (p *page) _setSlot(sl *slot, sid uint16) {
	// get the slot offset
	off := szHd + (sid * szSl)
	// make sure it is in bounds
	if off > bin.Uint16((*p)[offLower:]) {
		panic("slot id or offset is out of bounds")
	}
	// for clarity
	offStatus := off
	offOffset := off + 2
	offLength := off + 4
	// now we write the slot to the page
	bin.PutUint16((*p)[offStatus:], sl.status)
	bin.PutUint16((*p)[offOffset:], sl.offset)
	bin.PutUint16((*p)[offLength:], sl.length)
}

// _getSlot reads the slot at the provided slot location derived using
// the supplied slot index ID. getSlot panics if the provided sid
// is out of bounds.
func (p *page) _getSlot(sid uint16) *slot {
	// get the slot offset
	off := szHd + (sid * szSl)
	// make sure it is in bounds
	if lo := bin.Uint16((*p)[offLower:]); off > lo {
		// log.Printf(">>>>>>>> sid=%d, off=%d, lower=%d\n", sid, off, bin.Uint16((*p)[offLower:]))
		log.Println(">>>", p)
		panic("slot id or offset is out of bounds")
	}
	// for clarity
	offStatus := off
	offOffset := off + 2
	offLength := off + 4
	// now we write the slot to the page
	return &slot{
		status: bin.Uint16((*p)[offStatus:]),
		offset: bin.Uint16((*p)[offOffset:]),
		length: bin.Uint16((*p)[offLength:]),
	}
}

// _addSlot appends a new slot to the page
func (p *page) _addSlot(size uint16) (uint16, *slot) {
	// get page header
	h := p.getHeader()
	// grab the slot id for later
	sid := h.slots
	// update page header
	h.slots++
	h.lower += szSl
	h.upper -= size
	// create new slot structure
	sl := &slot{
		status: stUsed,
		offset: h.upper,
		length: size,
	}
	// write header back to page
	p.setHeader(h)
	// encode slot onto page
	p._setSlot(sl, sid)
	// finally, return sid and slot
	return sid, sl
}

// _acquireSlot adds or locates a free slot that will fit the record size
// provided. It returns a slot along with the slot index.
func (p *page) _acquireSlot(size uint16) (uint16, *slot) {
	// try to find a free slot we can use
	slotCount := bin.Uint16((*p)[offSlots:])
	for sid := uint16(0); sid < slotCount; sid++ {
		sl := p._getSlot(sid)
		if sl.status == stFree && size <= sl.length {
			// we can use this slot, but first we update and save it
			sl.status = stUsed
			sl.length = size
			p._setSlot(sl, sid)
			// and then we return it for use
			return sid, sl
		}
	}
	// otherwise, we append and return a fresh slot
	return p._addSlot(size)
}

// addRecord writes a new record to the page. It returns a *recID which
// is a record ID, along with any potential errors encountered.
func (p *page) addRecord(data []byte) (*recID, error) {
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// get the record size
	rsize := uint16(len(data))
	// sanity check the record
	err := p.checkRecord(rsize)
	if err != nil {
		return nil, err
	}
	// get a new (or used) slot to index the record
	sid, sl := p._acquireSlot(rsize)
	// get the record bounds from the slot index
	beg, end := sl.bounds()
	// write the record to the page (using record bounds)
	copy((*p)[beg:end], data)
	// assemble and return the record ID
	return &recID{
		pid: bin.Uint32((*p)[offPID:]),
		sid: sid,
	}, nil
}

// checkRID performs error and sanity checking on the provided
// record ID.
func (p *page) checkRID(id *recID) error {
	if id.pid != bin.Uint32((*p)[offPID:]) {
		return ErrInvalidPID
	}
	if id.sid > bin.Uint16((*p)[offSlots:]) {
		return ErrInvalidSID
	}
	return nil
}

// getRecord reads a record from the page. It returns the record data
// that is associated with the provided record ID, along with any
// potential errors encountered.
func (p *page) getRecord(id *recID) ([]byte, error) {
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// sanity check the record ID
	err := p.checkRID(id)
	if err != nil {
		return nil, err
	}
	// find the associated slot index (ensure it is a used slot)
	sl := p._getSlot(id.sid)
	if sl.status == stFree {
		return nil, ErrRecordNotFound
	}
	// create a buffer to copy the record into (safety)
	buff := make([]byte, sl.length)
	// get the record bounds from the slot index
	beg, end := sl.bounds()
	// copy the record into the buffer (using record bounds)
	copy(buff, (*p)[beg:end])
	// return the record copy
	return buff, nil
}

// _delSlot updates the status of the slot found at the provided
// slot ID, and returns the slot for use in the delete operation.
func (p *page) _delSlot(sid uint16) *slot {
	// get the slot using the sid
	sl := p._getSlot(sid)
	// if the slot status is free, return nil
	if sl.status == stFree {
		return nil
	}
	// update slot status
	sl.status = stFree
	// save the status of the found slot
	p._setSlot(sl, sid)
	// and return
	return sl
}

// delRecord removes a record from the page. It overwrites the record
// data with zeros and marks the slot index as a "free to use" slot
// so that it can be re-used at a later date if there is another
// record that can occupy the same (or less) space. It returns any
// errors encountered.
func (p *page) delRecord(id *recID) error {
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// sanity check the record ID
	err := p.checkRID(id)
	if err != nil {
		return err
	}
	// find the associated slot index (ensure it is a used slot)
	sl := p._delSlot(id.sid)
	if sl == nil {
		return nil
	}
	// create a buffer to overwrite the record with
	buff := make([]byte, sl.length)
	// get the record bounds from the slot index
	beg, end := sl.bounds()
	// copy the buffer over the record (using record bounds)
	copy((*p)[beg:end], buff)
	// return nil error
	return nil
}

// iterator is a basic iterator
type iterator struct {
	slots    []*slot
	index    int
	skipFree bool
}

// newIter instantiates and returns a new iterator. If the page contains
// no data entries, then it returns a nil iterator along with an error.
func (p *page) newIter(skipFree bool) (*iterator, error) {
	slots := p.getSlotSet()
	if slots == nil {
		return nil, ErrEmptyPage
	}
	return &iterator{
		slots:    slots,
		index:    -1,
		skipFree: skipFree,
	}, nil
}

// next returns the next record in this page.
func (it *iterator) next() *slot {
	// increment to the next slot
	it.index++
	// make sure we do not overflow the slots index
	if !it.hasMore() {
		fmt.Println(">>> hit it <<<")
		return nil
	}
	// get the slot at this index, then check to make sure that this
	// slot is not a free slot; if it is a free slot, then skip it.
	sl := it.slots[it.index]
	// check if we should skip any slots marked free.
	if it.skipFree && sl.status == stFree {
		// slot is free, skip it
		return it.next()
	}
	// return our slot
	return sl
}

// hasMore returns a boolean indicating true if this page contains one or more "next"
// returns that it can return. Otherwise, it returns false.
func (it *iterator) hasMore() bool {
	return it.index < len(it.slots)
}

// compact triggers a record compaction
func (p *page) compact() error {
	// First, we must allocate a new page to copy data into.
	pg := newPage(p.getPageID())
	// Next, we will get an iterator and check for any errors.
	it, err := p.newIter(true)
	if err != nil {
		return err
	}
	var n int
	// Now, we iterate the slots of the current page. The iterator skips all
	// records that are marked free.
	for sl := it.next(); it.hasMore() == true; sl = it.next() {
		// Get the record bounds for the current record on the current page.
		beg, end := sl.bounds()
		// Call addRecord from the new page passing the record slice
		// in from the current page.
		_, err = pg.addRecord((*p)[beg:end])
		if err != nil {
			return err
		}
		n++
	}
	// Make sure the iterator gets marked for collection.
	it = nil
	// Finished adding records to the new page, now swap the pages.
	*p = pg
	// Call the GC directly here
	runtime.GC()
	// Return our nil error
	fmt.Printf("wrote %d records\n", n)
	return nil
}

func (p page) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 16, 4, 0, ' ', tabwriter.DiscardEmptyColumns)
	var err error
	_, err = fmt.Fprintf(w, "pid\tmeta\tstatus\tslots\tlower\tupper\n")
	if err != nil {
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}
	n := buf.Len()
	hrule := strings.Repeat("-", n)
	_, err = fmt.Fprintf(w, "%s\n", hrule)
	if err != nil {
		panic(err)
	}
	h := p.getHeader()
	_, err = fmt.Fprintf(w, "%.4d\t%.4d\t%.4d\t%.4d\t%.4d\t%.4d", h.pid, h.meta, h.status, h.slots, h.lower, h.upper)
	if err != nil {
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}
	data := buf.String()
	return fmt.Sprintf("%s\n%s\n", hrule, data)
}

func (p *page) DumpPage(showPageData bool) string {
	h := p.getHeader()
	ss := fmt.Sprintf("+------------------[ page header ]------------------+\n")
	ss += fmt.Sprintf(
		"pid=%.2d, meta=%.2d, slots=%.2d, lo=%.3d, hi=%.4d [0x%.8x,0x%.4x,0x%.4x,0x%.4x]\n",
		h.pid, h.meta, h.slots, h.lower, h.upper, h.pid, h.slots, h.lower, h.upper,
	)
	ss += fmt.Sprintf("+------------------[ slots index ]------------------+\n")
	for sid := uint16(0); sid < h.slots; sid++ {
		sl := p._getSlot(sid)
		ss += fmt.Sprintf("%s\n", sl)
	}
	if showPageData {
		ss += fmt.Sprintf("+------------------[  page data  ]------------------+\n")
		hf := hex.Dump(*p)
		ss += fmt.Sprintf(hf[h.upper:])
	}
	ss += fmt.Sprintf("+----------------------------------------------+\n")
	return ss
}
