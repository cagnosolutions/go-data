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

// RE: scratch_184.go
// https://go.dev/play/p/gr8RC8vDuSv

// Various fixed size constants
const (
	recordHeaderSize = 6        // size of a recordHeader
	pageHeaderSize   = 24       // size of a pageHeader
	pageCellSize     = 6        // size of a cell pointer
	pageSize         = 16 << 10 // size of a page
)

// Flags for Page
const (
	P_FREE uint32 = 0x00000001 // indicates the Page is free to use
	P_USED uint32 = 0x00000002 // indicates the Page is being used
)

// Flags for cellPtr
const (
	C_FREE uint16 = 0x0001
	C_USED uint16 = 0x0002
)

func hasFlag16(flags, isSet uint16) bool {
	return (flags & isSet) != 0
}

func hasFlag32(flags, isSet uint32) bool {
	return (flags & isSet) != 0
}

// Binary offsets for Page pageHeader
const (
	offPID      uint32 = 0  // PID=uint32		offs=0-4 	(4 bytes)
	offPrev     uint32 = 4  // prev=uint32 		offs=4-8	(4 bytes)
	offNext     uint32 = 8  // next=uint32 		offs=8-12	(4 bytes)
	offFlags    uint32 = 12 // flags=uint32		offs=12-16	(4 bytes)
	offNumFree  uint16 = 16 // numFree=uint16	offs=16-18	(2 bytes)
	offNumCells uint16 = 18 // numCells=uint16	offs=18-20	(2 bytes)
	offLower    uint16 = 20 // lower=uint16		offs=20-22 	(2 bytes)
	offUpper    uint16 = 22 // upper=uint16		offs=22-24	(2 bytes)
)

// PageID is a PageID type
type PageID = uint32
type SlotID = uint16

// Page latch
var pgLatch sync.Mutex

// bin is just a little shorthand if you wish to easily change
// up the encoding and decoding byte order, this variable can
// simply be changed.
var bin = binary.LittleEndian

// cellPtr is a struct which is an index for a record. **It should be noted
// that a cellPtr pointer may also just end up becoming a single uint64.
type cellPtr struct {
	id     uint16
	flags  uint16
	offset uint16
	length uint16
}

// bounds returns the starting and ending offsets to the
// particular record that this cellPtr index "points" to.
func (c cellPtr) bounds() (uint16, uint16) {
	return c.offset, c.offset + c.length
}

// String is the string method for a cellPtr
func (c cellPtr) String() string {
	ss := fmt.Sprintf("off=%.4d, len=%.4d, numFree=%v", c.offset, c.length, hasFlag16(c.flags, C_FREE))
	ss += fmt.Sprintf("\t[0x%.4x,0x%.4x,0x%.4x]", c.offset, c.length, c.flags)
	return ss
}

// RecID is a struct representing a record ID.
type RecID struct {
	PID uint32 // PID is the Page ID
	CID uint16 // CID is the cellPtr ID (not the index)
}

// pageHeader is a struct representing a Page pageHeader.
type pageHeader struct {
	pid      uint32 // id of Page
	prev     uint32 // prev Page pointer
	next     uint32 // next Page pointer
	flags    uint32 // flags and meta data for Page
	numFree  uint16 // number of numFree cells
	numCells uint16 // number of cells allocated
	lower    uint16 // lower numFree space bound
	upper    uint16 // upper numFree space bound
}

// Page is a Page.
type Page []byte

// NewEmptyPage returns a new Page instance using the default page size. The provided
// page ID is set, along with the flags P_FREE, denoting it as an empty page that is
// free to use.
func NewEmptyPage(pid uint32) Page {
	pg := make(Page, pageSize, pageSize)
	pg.setHeader(
		&pageHeader{
			pid:      pid,
			prev:     0,
			next:     0,
			flags:    P_FREE,
			numFree:  0,
			numCells: 0,
			lower:    pageHeaderSize,
			upper:    pageSize,
		},
	)
	return pg
}

// NewPage returns a new Page instance using the default page size. The provided
// page ID is set, along with the flag P_USED, denoting it as a page that is not empty
// and probably (if not now, then momentarily) contains data.
func NewPage(pid uint32) Page {
	pg := make(Page, pageSize, pageSize)
	pg.setHeader(
		&pageHeader{
			pid:      pid,
			prev:     0,
			next:     0,
			flags:    P_USED,
			numFree:  0,
			numCells: 0,
			lower:    pageHeaderSize,
			upper:    pageSize,
		},
	)
	return pg
}

// setHeader encodes the provided pageHeader structure to the underlying
// Page.
func (p *Page) setHeader(h *pageHeader) {
	bin.PutUint32((*p)[offPID:offPID+4], h.pid)                // offset 00: PID 		(00-04) // 4 bytes
	bin.PutUint32((*p)[offPrev:offPrev+4], h.prev)             // offset 04: prev		(04-08) // 4 bytes
	bin.PutUint32((*p)[offNext:offNext+4], h.next)             // offset 08: next	(08-12) // 4 bytes
	bin.PutUint32((*p)[offFlags:offFlags+4], h.flags)          // offset 12: flags		(12-16) // 4 bytes
	bin.PutUint16((*p)[offNumFree:offNumFree+2], h.numFree)    // offset 16: numFree		(16-18) // 2 bytes
	bin.PutUint16((*p)[offNumCells:offNumCells+2], h.numCells) // offset 18: numCells		(18-20) // 2 bytes
	bin.PutUint16((*p)[offLower:offLower+2], h.lower)          // offset 20: lower		(20-22) // 2 bytes
	bin.PutUint16((*p)[offUpper:offUpper+2], h.upper)          // offset 22: upper		(22-24) // 2 bytes

	// 															// offset 24: begin cellPtr list
}

// getPageHeader decodes the pageHeader (from the underlying Page) and returns
// a pointer to a pageHeader structure
func (p *Page) getPageHeader() *pageHeader {
	return &pageHeader{
		pid:      decU32((*p)[offPID : offPID+4]),
		prev:     decU32((*p)[offPrev : offPrev+4]),
		next:     decU32((*p)[offNext : offNext+4]),
		flags:    decU32((*p)[offFlags : offFlags+4]),
		numFree:  decU16((*p)[offNumFree : offNumFree+2]),
		numCells: decU16((*p)[offNumCells : offNumCells+2]),
		lower:    decU16((*p)[offLower : offLower+2]),
		upper:    decU16((*p)[offUpper : offUpper+2]),
	}
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
	return decU16((*p)[offNumFree : offNumFree+4])
}

// getNumCells decodes and returns the number of allocated cells directly from the encoded pageHeader.
func (p *Page) getNumCells() uint16 {
	return decU16((*p)[offNumCells : offNumCells+4])
}

// getLower decodes and returns the lower bound marker directly from the encoded pageHeader.
func (p *Page) getLower() uint16 {
	return decU16((*p)[offLower : offLower+4])
}

// getUpper decodes and returns the upper bound marker directly from the encoded pageHeader.
func (p *Page) getUpper() uint16 {
	return decU16((*p)[offUpper : offUpper+4])
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
	encU16((*p)[offNumFree:offNumFree+4], n)
}

// setNumCells encodes the provided value directly into the pageHeader.
func (p *Page) setNumCells(n uint16) {
	encU16((*p)[offNumCells:offNumCells+4], n)
}

// setLower encodes the provided value directly into the pageHeader.
func (p *Page) setLower(n uint16) {
	encU16((*p)[offLower:offLower+4], n)
}

// setUpper encodes the provided value directly into the pageHeader.
func (p *Page) setUpper(n uint16) {
	encU16((*p)[offUpper:offUpper+4], n)
}

// Size returns the page size in bytes
func (p *Page) Size() int {
	if p == nil {
		return 0
	}
	return len(*p)
}

// freeSpace returns the amount of contiguous numFree space left in the
// Page (space between the upper and lower bounds)
func (p *Page) freeSpace() uint16 {
	free := decU16((*p)[offUpper:offUpper+2]) - decU16((*p)[offLower:offLower+2])
	if p != nil && int(free) > len(*p) {
		return 0
	}
	return free
}

// checkRecord performs sanity and error checking on a record size
func (p *Page) checkRecord(size uint16) error {
	if size < pageCellSize {
		return errs.ErrRecordTooSmall
	}
	// numFree := p.freeSpace() - pageCellSize
	// util.DEBUG("checkRecord, numFree=%d, prev=%d", numFree, prev)
	if size >= p.freeSpace() {
		return ErrNoRoom
	}
	return nil
}

// getSlotSet decodes and returns a set of cellPtr pointer for this Page.
// It will return nil if there are no numCells on this Page. Any changes
// made to this list of numCells is not persisted until calling setSlots.
func (p *Page) getSlotSet() []*cellPtr {
	// Check if there are any numCells to return.
	slotCount := decU16((*p)[offNumCells : offNumCells+2])
	if slotCount < 1 {
		// No numCells to decode
		return nil
	}
	// We have numCells we can decode. Create a set we can append to.
	var slots []*cellPtr
	// Start looping, decoding, and adding numCells to our cellPtr set.
	for sid := uint16(0); sid < slotCount; sid++ {
		// Append the cellPtr to the cellPtr set.
		slots = append(slots, p.getSlot(sid))
	}
	// Finally, return our cellPtr set.
	return slots
}

// __setSlotSet encodes a set of cellPtr pointers into this Page. It will
// return an error if there is not enough room to write the set of numCells
// to the underlying Page.
func (p *Page) __setSlotSet(_ []*cellPtr) error {
	// Not sure if I want to append or overwrite this at the moment...
	return nil
}

// setSlot writes the provided cellPtr to the location derived using
// the supplied cellPtr index ID. setSlot panics if the provided CID
// is out of bounds.
func (p *Page) setSlot(sl *cellPtr, sid uint16) {
	// get the cellPtr offset
	off := pageHeaderSize + (sid * pageCellSize)
	// make sure it is in bounds
	if lo := decU16((*p)[offLower : offLower+2]); off > lo {
		panic(fmt.Sprintf("--setSlot: cellPtr id (%d) or offset (%d) is out of bounds (%d)", sid, off, lo))
	}
	// for clarity
	offStatus := off
	offOffset := off + 2
	offLength := off + 4
	// now we write the cellPtr to the Page
	bin.PutUint16((*p)[offStatus:offStatus+2], sl.flags)
	bin.PutUint16((*p)[offOffset:offOffset+2], sl.offset)
	bin.PutUint16((*p)[offLength:offLength+2], sl.length)
}

// getSlot reads the cellPtr at the provided cellPtr location derived using
// the supplied cellPtr index ID. getSlot panics if the provided CID
// is out of bounds.
func (p *Page) getSlot(sid uint16) *cellPtr {
	// get the cellPtr offset
	off := pageHeaderSize + (sid * pageCellSize)
	// make sure it is in bounds
	if lo := decU16((*p)[offLower : offLower+2]); off > lo {
		panic(fmt.Sprintf("--getSlot: cellPtr id (%d) or offset (%d) is out of bounds (%d)", sid, off, lo))
	}
	// for clarity
	offStatus := off
	offOffset := off + 2
	offLength := off + 4
	// now we write the cellPtr to the Page
	return &cellPtr{
		flags:  decU16((*p)[offStatus : offStatus+2]),
		offset: decU16((*p)[offOffset : offOffset+2]),
		length: decU16((*p)[offLength : offLength+2]),
	}
}

// addSlot appends a new cellPtr to the Page
func (p *Page) addSlot(size uint16) (uint16, *cellPtr) {
	// get Page pageHeader
	h := p.getPageHeader()
	// grab the cellPtr id for later
	sid := h.numCells
	// update Page pageHeader
	h.numCells++
	h.lower += pageCellSize
	h.upper -= size
	// create new cellPtr structure
	sl := &cellPtr{
		id:     h.numCells,
		flags:  C_USED,
		offset: h.upper,
		length: size,
	}
	// write pageHeader back to Page
	p.setHeader(h)
	// encode cellPtr onto Page
	p.setSlot(sl, sid)
	// finally, return CID and cellPtr
	return sid, sl
}

// acquireSlot adds or locates a numFree cellPtr that will fit the record size
// provided. It returns a cellPtr along with the cellPtr index.
func (p *Page) acquireSlot(size uint16) (uint16, *cellPtr) {
	// try to find a numFree cellPtr we can use
	slotCount := decU16((*p)[offNumCells : offNumCells+2])
	// first we check to see if the cellPtr count is zero
	if slotCount == 0 {
		// we can skip the mess if this is the case
		return p.addSlot(size)
	}
	//
	//
	// NOTE: we might want to keep a first open cellPtr id in
	// the pageHeader, as well as know if we are dealing with
	// fixed prev data or not. If we are dealing with fixed
	// sized data, then we can count forward from the first
	// open cellPtr because we know no matter what, we can
	// fit our data into any cellPtr. Otherwise, if we are
	// dealing with dynamic sized data, then we may want to
	// count backwards because we will have a higher chance
	// of encountering (not always, but sometimes) an empty
	// cellPtr that may fit our dynamic sized data sooner.
	//
	// NOTE: also, we may want to implement a cellPtr sorting
	// function (for fixed sized data entries) that does not
	// change the cellPtr id of anything, but simply re-orders
	// the offsets to point to the data in sorted order. Or
	// maybe it's even faster to do that operation when
	// acquiring or adding a new cellPtr.
	//
	//
	// we try from the last one available and work backwards
	// for CID := slotCount; CID > 0; CID-- {
	for sid := uint16(0); sid < slotCount; sid++ {
		sl := p.getSlot(sid)
		if hasFlag16(sl.flags, C_FREE) && size <= sl.length {
			// we can use this cellPtr, but first we update and save it
			sl.flags = C_USED
			sl.length = size
			p.setSlot(sl, sid)
			// and then we return it for use
			return sid, sl
		}
	}
	// otherwise, we append and return a fresh cellPtr
	return p.addSlot(size)
}

// AddRecord writes a new record to the Page. It returns a *RecID which
// is a record ID, along with any potential errors encountered.
func (p *Page) AddRecord(data []byte) (*RecID, error) {
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// get the record prev
	rsize := uint16(len(data))
	// sanity check the record
	err := p.checkRecord(rsize)
	// util.DEBUG("--AddRecord(%q) rsize=%d, numFree=%d, checkRecordErr=%v", data, rsize, p.freeSpace(), err)
	if err != nil {
		return nil, err
	}
	// get a new (or used) cellPtr to index the record
	sid, sl := p.acquireSlot(rsize)
	// util.DEBUG("--AddRecord(%x) rsize=%d, use_slot=%d", data, rsize, CID)
	// get the record bounds from the cellPtr index
	beg, end := sl.bounds()
	// write the record to the Page (using record bounds)
	copy((*p)[beg:end], data)
	// assemble and return the record ID
	return &RecID{
		PID: decU32((*p)[offPID : offPID+4]),
		CID: sid,
	}, nil
}

// checkRID performs error and sanity checking on the provided
// record ID.
func (p *Page) checkRID(rid *RecID) error {
	if rid.PID != decU32((*p)[offPID:offPID+4]) {
		return ErrInvalidPID
	}
	if rid.CID > decU16((*p)[offNumCells:offNumCells+2]) {
		return ErrInvalidSID
	}
	return nil
}

// GetRecord reads a record from the Page. It returns the record data
// that is associated with the provided record ID, along with any
// potential errors encountered.
func (p *Page) GetRecord(rid *RecID) ([]byte, error) {
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// sanity check the record ID
	err := p.checkRID(rid)
	if err != nil {
		return nil, err
	}
	// find the associated cellPtr index (ensure it is a used cellPtr)
	sl := p.getSlot(rid.CID)
	if hasFlag16(sl.flags, C_FREE) {
		return nil, ErrRecordNotFound
	}
	// create a buffer to copy the record into (safety)
	buff := make([]byte, sl.length)
	// get the record bounds from the cellPtr index
	beg, end := sl.bounds()
	// copy the record into the buffer (using record bounds)
	copy(buff, (*p)[beg:end])
	// return the record copy
	return buff, nil
}

// delSlot updates the numFree of the cellPtr found at the provided
// cellPtr ID, and returns the cellPtr for use in the delete operation.
func (p *Page) delSlot(sid uint16) *cellPtr {
	// get the cellPtr using the CID
	sl := p.getSlot(sid)
	// if the cellPtr numFree is numFree, return nil
	if hasFlag16(sl.flags, C_FREE) {
		return nil
	}
	// update cellPtr numFree
	sl.flags = C_FREE
	// save the numFree of the found cellPtr
	p.setSlot(sl, sid)
	// and return
	return sl
}

// delRecord removes a record from the Page. It overwrites the record
// data with zeros and marks the cellPtr index as a "numFree to use" cellPtr
// so that it can be re-used at a later date if there is another
// record that can occupy the same (or less) space. It returns any
// errors encountered.
func (p *Page) delRecord(rid *RecID) error {
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// sanity check the record ID
	err := p.checkRID(rid)
	if err != nil {
		return err
	}
	// find the associated cellPtr index (ensure it is a used cellPtr)
	sl := p.delSlot(rid.CID)
	if sl == nil {
		return nil
	}
	// create a buffer to overwrite the record with
	buff := make([]byte, sl.length)
	// get the record bounds from the cellPtr index
	beg, end := sl.bounds()
	// util.DEBUG("--delRecord(%d) [%x] del_slot=%d, slot_count=%d", id.CID, (*p)[beg:end], id.CID, len(p.getSlotSet()))
	// copy the buffer over the record (using record bounds)
	copy((*p)[beg:end], buff)
	// return nil error
	return nil
}

// clear wipes all the data and returns the Page to its original form
func (p *Page) clear() {
	*p = NewEmptyPage(p.getPageID())
}

// iterator is a basic iterator
type iterator struct {
	slots    []*cellPtr
	index    int
	skipFree bool
}

// newIter instantiates and returns a new iterator. If the Page contains
// no data entries, then it returns a nil iterator along with an error.
func (p *Page) newIter(skipFree bool) (*iterator, error) {
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

// next returns the next record in this Page.
func (it *iterator) next() *cellPtr {
	// increment to the next cellPtr
	it.index++
	// make sure we do not overflow the numCells index
	if !it.hasMore() {
		fmt.Println(">>> hit it <<<")
		return nil
	}
	// get the cellPtr at this index, then check to make sure that this
	// cellPtr is not a numFree cellPtr; if it is a numFree cellPtr, then skip it.
	sl := it.slots[it.index]
	// check if we should skip any numCells marked numFree.
	if it.skipFree && hasFlag16(sl.flags, C_FREE) {
		// cellPtr is numFree, skip it
		return it.next()
	}
	// return our cellPtr
	return sl
}

// hasMore returns a boolean indicating true if this Page contains one or more "next"
// returns that it can return. Otherwise, it returns false.
func (it *iterator) hasMore() bool {
	return it.index < len(it.slots)
}

// compact triggers a record compaction
func (p *Page) compact() error {
	// First, we must allocate a new Page to copy data into.
	pg := NewPage(p.getPageID())
	// Next, we will get an iterator and check for any errors.
	it, err := p.newIter(true)
	if err != nil {
		return err
	}
	var n int
	// Now, we iterate the numCells of the current Page. The iterator skips all
	// records that are marked numFree.
	for sl := it.next(); it.hasMore() == true; sl = it.next() {
		// Get the record bounds for the current record on the current Page.
		beg, end := sl.bounds()
		// Call AddRecord from the new Page passing the record slice
		// in from the current Page.
		_, err = pg.AddRecord((*p)[beg:end])
		if err != nil {
			return err
		}
		n++
	}
	// Make sure the iterator gets marked for collection.
	it = nil
	// Finished adding records to the new Page, now swap the pages.
	*p = pg
	// Call the GC directly here
	runtime.GC()
	// Return our nil error
	fmt.Printf("wrote %d records\n", n)
	return nil
}

func (p *Page) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 16, 4, 0, ' ', tabwriter.DiscardEmptyColumns)
	var err error
	_, err = fmt.Fprintf(w, "PID\tflags\tnumFree\tnumCells\tlower\tupper\n")
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
	h := p.getPageHeader()
	_, err = fmt.Fprintf(
		w, "%.4d\t%.4d\t%.4d\t%.4d\t%.4d\t%.4d", h.pid, h.flags, h.numFree, h.numCells, h.lower, h.upper,
	)
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

func (p *Page) DumpPage(showPageData bool) string {
	h := p.getPageHeader()
	ss := fmt.Sprintf("+------------------[ Page pageHeader ]------------------+\n")
	ss += fmt.Sprintf(
		"PID=%.2d, prev=%.2d, next=%.2d, flags=%.2d, numFree=%.2d, numCells=%.2d, "+
			"lo=%.3d, hi=%.4d [0x%.8x,0x%.4x,0x%.4x,0x%.4x]\n",
		h.pid, h.prev, h.next, h.flags, h.numFree, h.numCells, h.lower, h.upper, h.pid, h.numCells, h.lower, h.upper,
	)
	ss += fmt.Sprintf("+------------------[ numCells index ]------------------+\n")
	for sid := uint16(0); sid < h.numCells; sid++ {
		sl := p.getSlot(sid)
		ss += fmt.Sprintf("%s\n", sl)
	}
	if showPageData {
		ss += fmt.Sprintf("+------------------[  Page data  ]------------------+\n")
		hf := hex.Dump(*p)
		ss += fmt.Sprintf(hf[h.upper:])
	}
	ss += fmt.Sprintf("+----------------------------------------------+\n")
	return ss
}

func (p *Page) FillPercent() float32 {
	/*
		fmt.Printf("pageSize=%d, used=%d, unused=%d, percent=%.2f%%\n",
				pgsize, datused, pgsize-datused, (float32(datused)/float32(pgsize))*100)
	*/
	pgsize, unused := p.Size(), p.freeSpace()
	return ((float32(pgsize) - float32(unused)) / float32(pgsize)) * 100
}

func (p *Page) Statistics() string {
	size := p.Size()
	free := uint32(p.freeSpace())
	used := size - int(free)
	percentFull := float32(((float32(size) - float32(free)) / float32(size)) * 100)
	percentFree := float32((float32(free) / float32(size)) * 100)
	ss := fmt.Sprintf("~~~~[Page %d statistics]~~~~\n", p.getPageID())

	ss += fmt.Sprintf("   page_size:\t%d\t(%.2fkB)\n", size, float32(size)/1024)
	ss += fmt.Sprintf("  bytes_used:\t%d\t(%.2fkB)\n", used, float32(used)/1024)
	ss += fmt.Sprintf("  bytes_free:\t%d\t(%.2fkB)\n", free, float32(free)/1024)
	ss += fmt.Sprintf("percent_full:\t%.2f\n", percentFull)
	ss += fmt.Sprintf("percent_free:\t%.2f\n", percentFree)

	ss += fmt.Sprintf("~~~~[Page %d statistics]~~~~\n", p.getPageID())
	return ss
}

// printHeader prints the page header
func (p *Page) printHeader() {
	h := p.getPageHeader()
	fmt.Printf("pageHeader:\n")
	fmt.Printf(
		"\tPID=%d\t\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.pid, 4, offPID, offPID+4,
		(*p)[offPID:offPID+4],
	)
	fmt.Printf(
		"\tprev=%d\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.prev, 4, offPrev, offPrev+4,
		(*p)[offPrev:offPrev+4],
	)
	fmt.Printf(
		"\tnext=%d\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.next, 4, offNext, offNext+4,
		(*p)[offNext:offNext+4],
	)
	fmt.Printf(
		"\tflags=%d\t\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.flags, 4, offFlags, offFlags+4,
		(*p)[offFlags:offFlags+4],
	)
	fmt.Printf(
		"\tnumFree=%d\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.numFree, 2, offNumFree, offNumFree+2,
		(*p)[offNumFree:offNumFree+2],
	)
	fmt.Printf(
		"\tnumCells=%d\t\t\t(%d bytes, offs=%.2d-%.2d, data=%v)\n", h.numCells, 2, offNumCells, offNumCells+2,
		(*p)[offNumCells:offNumCells+2],
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
