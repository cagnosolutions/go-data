package page

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/cagnosolutions/go-data/pkg/dbms/errs"
)

// https://go.dev/play/p/gr8RC8vDuSv

// Defaults for page size
const (
	DefaultPageSize = szPg // 4KB
	MinPageSize     = szSl
	MaxPageSize     = DefaultPageSize - szHd - szSl
)

const (
	szHd = 24      // fileSize of page header (in bytes)
	szSl = 6       // fileSize of slot index (in bytes)
	szPg = 4 << 10 // fileSize of page (default)
)

// Binary offsets for page header
const (
	offPID      uint32 = 0  // pid=uint32		offs=0-4 	(4 bytes)
	offSize     uint32 = 4  // size=uint32 	offs=4-8	(4 bytes)
	offReserved uint32 = 8  // reserved=uint32 offs=8-12	(4 bytes)
	offMeta     uint32 = 12 // meta=uint32		offs=12-16	(4 bytes)
	offStat     uint16 = 16 // status=uint16	offs=16-18	(2 bytes)
	offSlots    uint16 = 18 // slots=uint16	offs=18-20	(2 bytes)
	offLower    uint16 = 20 // lower=uint16	offs=20-22 	(2 bytes)
	offUpper    uint16 = 22 // upper=uint16	offs=22-24	(2 bytes)
)

// Status flags (page or slot status)
const (
	StatFree uint16 = 0x0001 // page or slot is free
	StatUsed uint16 = 0x0002 // page or slot to use
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
	ss := fmt.Sprintf("off=%.4d, len=%.4d, free=%v", s.offset, s.length, s.status == StatFree)
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
	pid      uint32 // id of page
	size     uint32 // size of the page
	reserved uint32 // reserved for later use
	meta     uint32 // meta data for page
	status   uint16 // status of page
	slots    uint16 // number of slots
	lower    uint16 // lower free space bound
	upper    uint16 // upper free space bound
}

type (
	Page  = page
	RecID = struct {
		PID PageID
		SID SlotID
	}
	PageID = pageID
	SlotID = slotID
)

// pageID is a pageID type
type pageID = uint32
type slotID = uint16

// page is a page.
type page []byte

func (p *Page) Size() int {
	return p.size()
}

func (p *page) size() int {
	if p == nil {
		return 0
	}
	return len(*p)
}

func NewEmptyPage(pid PageID) Page {
	return newEmptyPageSize(pid, DefaultPageSize)
}

// newEmptyPage returns a new page instance set with the provided page ID,
// with a meta byte status of StatFree, denoting it as empty and free to use.
func newEmptyPage(pid uint32) page {
	return newEmptyPageSize(pid, DefaultPageSize)
}

func NewEmptyPageSize(pid PageID, size uint16) Page {
	return newEmptyPageSize(pid, size)
}

// newEmptyPageSize returns a new page instance set with the provided page ID, with
// a meta byte status of StatFree, denoting it as empty and free to use. It is sized
// according to the provided size.
func newEmptyPageSize(pid uint32, size uint16) page {
	pg := make(page, size, size)
	pg.setHeader(
		&header{
			pid:      pid,
			size:     uint32(size),
			reserved: uint32(0),
			meta:     mdSlotted | mdRecDynmc,
			status:   StatFree,
			slots:    uint16(0),
			lower:    szHd,
			upper:    size,
		},
	)
	return pg
}

func NewPage(pid PageID) Page {
	return newPageSize(pid, DefaultPageSize)
}

// newPage returns a new page instance set with the provided page ID.
func newPage(pid uint32) page {
	return newPageSize(pid, DefaultPageSize)
}

func NewPageSize(pid PageID, size uint16) Page {
	return newPageSize(pid, size)
}

// newPageSize returns a new page instance set with the provided page ID, and sized
// according to the provided size.
func newPageSize(pid uint32, size uint16) page {
	pg := make(page, size, size)
	pg.setHeader(
		&header{
			pid:      pid,
			size:     uint32(size),
			reserved: uint32(0),
			meta:     mdSlotted | mdRecDynmc,
			status:   StatUsed,
			slots:    uint16(0),
			lower:    szHd,
			upper:    size,
		},
	)
	return pg
}

// setHeader encodes the provided header structure to the underlying
// page.
func (p *page) setHeader(h *header) {
	bin.PutUint32((*p)[offPID:offPID+4], h.pid)                // offset 00: pid 		(00-04) // 4 bytes
	bin.PutUint32((*p)[offSize:offSize+4], h.size)             // offset 04: size		(04-08) // 4 bytes
	bin.PutUint32((*p)[offReserved:offReserved+4], h.reserved) // offset 08: reserved	(08-12) // 4 bytes
	bin.PutUint32((*p)[offMeta:offMeta+4], h.meta)             // offset 12: meta		(12-16) // 4 bytes
	bin.PutUint16((*p)[offStat:offStat+2], h.status)           // offset 16: status		(16-18) // 2 bytes
	bin.PutUint16((*p)[offSlots:offSlots+2], h.slots)          // offset 18: slots		(18-20) // 2 bytes
	bin.PutUint16((*p)[offLower:offLower+2], h.lower)          // offset 20: lower		(20-22) // 2 bytes
	bin.PutUint16((*p)[offUpper:offUpper+2], h.upper)          // offset 22: upper		(22-24) // 2 bytes

	// 															// offset 24: begin slot list
}

func (p *page) printHeader() {
	h := p.getHeader()
	fmt.Printf("header:\n")
	fmt.Printf(
		"\tpid=%d\t\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.pid, 4, offPID, offPID+4,
		(*p)[offPID:offPID+4],
	)
	fmt.Printf(
		"\tsize=%d\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.size, 4, offSize, offSize+4,
		(*p)[offSize:offSize+4],
	)
	fmt.Printf(
		"\treserved=%d\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.reserved, 4, offReserved, offReserved+4,
		(*p)[offReserved:offReserved+4],
	)
	fmt.Printf(
		"\tmeta=%d\t\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.meta, 4, offMeta, offMeta+4,
		(*p)[offMeta:offMeta+4],
	)
	fmt.Printf(
		"\tstatus=%d\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.status, 2, offStat, offStat+2,
		(*p)[offStat:offStat+2],
	)
	fmt.Printf(
		"\tslots=%d\t\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.slots, 2, offSlots, offSlots+2,
		(*p)[offSlots:offSlots+2],
	)
	fmt.Printf(
		"\tlower=%d\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.lower, 2, offLower, offLower+2,
		(*p)[offLower:offLower+2],
	)
	fmt.Printf(
		"\tupper=%d\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.upper, 2, offUpper, offUpper+2,
		(*p)[offUpper:offUpper+2],
	)
}

// getHeader decodes (from the underlying page) and returns a
// pointer to a header structure
func (p *page) getHeader() *header {
	return &header{
		pid:      bin.Uint32((*p)[offPID : offPID+4]),
		size:     bin.Uint32((*p)[offSize : offSize+4]),
		reserved: bin.Uint32((*p)[offReserved : offReserved+4]),
		meta:     bin.Uint32((*p)[offMeta : offMeta+4]),
		status:   bin.Uint16((*p)[offStat : offStat+2]),
		slots:    bin.Uint16((*p)[offSlots : offSlots+2]),
		lower:    bin.Uint16((*p)[offLower : offLower+2]),
		upper:    bin.Uint16((*p)[offUpper : offUpper+2]),
	}
}

func (p *Page) GetPageID() PageID {
	return p.getPageID()
}

// getPageID returns the current page ID.
func (p *page) getPageID() uint32 {
	return bin.Uint32((*p)[offPID : offPID+4])
}

// getPageSize returns the page size.
func (p *page) getPageSize() uint32 {
	return bin.Uint32((*p)[offSize : offSize+4])
}

// freeSpace returns the amount of contiguous free space left in the
// page (space between the upper and lower bounds)
func (p *page) freeSpace() uint16 {
	free := bin.Uint16((*p)[offUpper:offUpper+2]) - bin.Uint16((*p)[offLower:offLower+2])
	if p != nil && int(free) > len(*p) {
		return 0
	}
	return free
}

// checkRecord performs sanity and error checking on a record size
func (p *page) checkRecord(size uint16) error {
	if size < szSl {
		return errs.ErrRecordTooSmall
	}
	// free := p.freeSpace() - szSl
	// util.DEBUG("checkRecord, free=%d, size=%d", free, size)
	if size >= p.freeSpace() {
		return errs.ErrNoRoom
	}
	return nil
}

// getSlotSet decodes and returns a set of slot pointer for this page.
// It will return nil if there are no slots on this page. Any changes
// made to this list of slots is not persisted until calling setSlots.
func (p *page) getSlotSet() []*slot {
	// Check if there are any slots to return.
	slotCount := bin.Uint16((*p)[offSlots : offSlots+2])
	if slotCount < 1 {
		// No slots to decode
		return nil
	}
	// We have slots we can decode. Create a set we can append to.
	var slots []*slot
	// Start looping, decoding, and adding slots to our slot set.
	for sid := uint16(0); sid < slotCount; sid++ {
		// Append the slot to the slot set.
		slots = append(slots, p.getSlot(sid))
	}
	// Finally, return our slot set.
	return slots
}

// __setSlotSet encodes a set of slot pointers into this page. It will
// return an error if there is not enough room to write the set of slots
// to the underlying page.
func (p *page) __setSlotSet(_ []*slot) error {
	// Not sure if I want to append or overwrite this at the moment...
	return nil
}

// setSlot writes the provided slot to the location derived using
// the supplied slot index ID. setSlot panics if the provided sid
// is out of bounds.
func (p *page) setSlot(sl *slot, sid uint16) {
	// get the slot offset
	off := szHd + (sid * szSl)
	// make sure it is in bounds
	if lo := bin.Uint16((*p)[offLower : offLower+2]); off > lo {
		panic(fmt.Sprintf("--setSlot: slot id (%d) or offset (%d) is out of bounds (%d)", sid, off, lo))
	}
	// for clarity
	offStatus := off
	offOffset := off + 2
	offLength := off + 4
	// now we write the slot to the page
	bin.PutUint16((*p)[offStatus:offStatus+2], sl.status)
	bin.PutUint16((*p)[offOffset:offOffset+2], sl.offset)
	bin.PutUint16((*p)[offLength:offLength+2], sl.length)
}

// getSlot reads the slot at the provided slot location derived using
// the supplied slot index ID. getSlot panics if the provided sid
// is out of bounds.
func (p *page) getSlot(sid uint16) *slot {
	// get the slot offset
	off := szHd + (sid * szSl)
	// make sure it is in bounds
	if lo := bin.Uint16((*p)[offLower : offLower+2]); off > lo {
		panic(fmt.Sprintf("--getSlot: slot id (%d) or offset (%d) is out of bounds (%d)", sid, off, lo))
	}
	// for clarity
	offStatus := off
	offOffset := off + 2
	offLength := off + 4
	// now we write the slot to the page
	return &slot{
		status: bin.Uint16((*p)[offStatus : offStatus+2]),
		offset: bin.Uint16((*p)[offOffset : offOffset+2]),
		length: bin.Uint16((*p)[offLength : offLength+2]),
	}
}

// addSlot appends a new slot to the page
func (p *page) addSlot(size uint16) (uint16, *slot) {
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
		status: StatUsed,
		offset: h.upper,
		length: size,
	}
	// write header back to page
	p.setHeader(h)
	// encode slot onto page
	p.setSlot(sl, sid)
	// finally, return sid and slot
	return sid, sl
}

// acquireSlot adds or locates a free slot that will fit the record size
// provided. It returns a slot along with the slot index.
func (p *page) acquireSlot(size uint16) (uint16, *slot) {
	// try to find a free slot we can use
	slotCount := bin.Uint16((*p)[offSlots : offSlots+2])
	// first we check to see if the slot count is zero
	if slotCount == 0 {
		// we can skip the mess if this is the case
		return p.addSlot(size)
	}
	// we try from the last one available and work backwards
	for sid := slotCount; sid > 0; sid-- {
		// for sid := uint16(0); sid < slotCount; sid++ {
		sl := p.getSlot(sid)
		if sl.status == StatFree && size <= sl.length {
			// we can use this slot, but first we update and save it
			sl.status = StatUsed
			sl.length = size
			p.setSlot(sl, sid)
			// and then we return it for use
			return sid, sl
		}
	}
	// otherwise, we append and return a fresh slot
	return p.addSlot(size)
}

func (p *Page) AddRecord(data []byte) (*RecID, error) {
	rid, err := p.addRecord(data)
	if err != nil {
		return nil, err
	}
	return &RecID{
		PID: rid.pid,
		SID: rid.sid,
	}, nil
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
	// util.DEBUG("--addRecord(%q) rsize=%d, free=%d, checkRecordErr=%v", data, rsize, p.freeSpace(), err)
	if err != nil {
		return nil, err
	}
	// get a new (or used) slot to index the record
	sid, sl := p.acquireSlot(rsize)
	// get the record bounds from the slot index
	beg, end := sl.bounds()
	// write the record to the page (using record bounds)
	copy((*p)[beg:end], data)
	// assemble and return the record ID
	return &recID{
		pid: bin.Uint32((*p)[offPID : offPID+4]),
		sid: sid,
	}, nil
}

// checkRID performs error and sanity checking on the provided
// record ID.
func (p *page) checkRID(id *recID) error {
	if id.pid != bin.Uint32((*p)[offPID:offPID+4]) {
		return errs.ErrInvalidPID
	}
	if id.sid > bin.Uint16((*p)[offSlots:offSlots+2]) {
		return errs.ErrInvalidSID
	}
	return nil
}

func (p *Page) GetRecord(rid *RecID) ([]byte, error) {
	return p.getRecord(&recID{
		pid: rid.PID,
		sid: rid.SID,
	})
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
	sl := p.getSlot(id.sid)
	if sl.status == StatFree {
		return nil, errs.ErrRecordNotFound
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

// delSlot updates the status of the slot found at the provided
// slot ID, and returns the slot for use in the delete operation.
func (p *page) delSlot(sid uint16) *slot {
	// get the slot using the sid
	sl := p.getSlot(sid)
	// if the slot status is free, return nil
	if sl.status == StatFree {
		return nil
	}
	// update slot status
	sl.status = StatFree
	// save the status of the found slot
	p.setSlot(sl, sid)
	// and return
	return sl
}

func (p *Page) DelRecord(rid *RecID) error {
	return p.delRecord(&recID{
		pid: rid.PID,
		sid: rid.SID,
	})
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
	sl := p.delSlot(id.sid)
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

// clear wipes all the data and returns the page to its original form
func (p *page) clear() {
	*p = newEmptyPageSize(p.getPageID(), uint16(p.getPageSize()))
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
		return nil, errs.ErrEmptyPage
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
	if it.skipFree && sl.status == StatFree {
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

func (p *page) String() string {
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
		"pid=%.2d, size=%.2d, reserved=%.2d, meta=%.2d, status=%.2d, slots=%.2d, "+
			"lo=%.3d, hi=%.4d [0x%.8x,0x%.4x,0x%.4x,0x%.4x]\n",
		h.pid, h.size, h.reserved, h.meta, h.status, h.slots, h.lower, h.upper, h.pid, h.slots, h.lower, h.upper,
	)
	ss += fmt.Sprintf("+------------------[ slots index ]------------------+\n")
	for sid := uint16(0); sid < h.slots; sid++ {
		sl := p.getSlot(sid)
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

func (p *page) FillPercent() float32 {
	/*
		fmt.Printf("pageSize=%d, used=%d, unused=%d, percent=%.2f%%\n",
				pgsize, datused, pgsize-datused, (float32(datused)/float32(pgsize))*100)
	*/
	pgsize, unused := p.getPageSize(), p.freeSpace()
	return ((float32(pgsize) - float32(unused)) / float32(pgsize)) * 100
}

func (p *page) Statistics() string {
	size := p.getPageSize()
	free := uint32(p.freeSpace())
	used := size - free
	percentFull := float32(((float32(size) - float32(free)) / float32(size)) * 100)
	percentFree := float32((float32(free) / float32(size)) * 100)
	ss := fmt.Sprintf("~~~~[page %d statistics]~~~~\n", p.getPageID())

	ss += fmt.Sprintf("   page_size:\t%d\t(%.2fkB)\n", size, float32(size)/1024)
	ss += fmt.Sprintf("  bytes_used:\t%d\t(%.2fkB)\n", used, float32(used)/1024)
	ss += fmt.Sprintf("  bytes_free:\t%d\t(%.2fkB)\n", free, float32(free)/1024)
	ss += fmt.Sprintf("percent_full:\t%.2f\n", percentFull)
	ss += fmt.Sprintf("percent_free:\t%.2f\n", percentFree)

	ss += fmt.Sprintf("~~~~[page %d statistics]~~~~\n", p.getPageID())
	return ss
}
