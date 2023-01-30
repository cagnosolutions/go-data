package page

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	// "github.com/cagnosolutions/go-data/pkg/dbms/errs"
)

// RE: scratch_184.go
// https://go.dev/play/p/gr8RC8vDuSv
// https://go.dev/play/p/gr8RC8vDuSv

// Various fixed size constants
const (
	recordHeaderSize = 6        // size of a recordHeader
	pageHeaderSize   = 24       // size of a pageHeader
	pageCellSize     = 8        // size of a cell pointer
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
//
//	type cellPtr struct {
//		id     uint16
//		flags  uint16
//		offset uint16
//		length uint16
//	}
type cellPtr = cell

// bounds returns the starting and ending offsets to the
// particular record that this cellPtr index "points" to.
func (c cellPtr) bounds() (uint16, uint16) {
	return c.getBounds() // c.offset, c.offset + c.length
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
	pg.setPageHeader(
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
	pg.setPageHeader(
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

// setPageHeader encodes the provided pageHeader structure to the underlying
// Page.
func (p *Page) setPageHeader(h *pageHeader) {
	encU32((*p)[offPID:offPID+4], h.pid)                // offset 00: PID 		(00-04) // 4 bytes
	encU32((*p)[offPrev:offPrev+4], h.prev)             // offset 04: prev		(04-08) // 4 bytes
	encU32((*p)[offNext:offNext+4], h.next)             // offset 08: next		(08-12) // 4 bytes
	encU32((*p)[offFlags:offFlags+4], h.flags)          // offset 12: flags		(12-16) // 4 bytes
	encU16((*p)[offNumFree:offNumFree+2], h.numFree)    // offset 16: numFree	(16-18) // 2 bytes
	encU16((*p)[offNumCells:offNumCells+2], h.numCells) // offset 18: numCells	(18-20) // 2 bytes
	encU16((*p)[offLower:offLower+2], h.lower)          // offset 20: lower		(20-22) // 2 bytes
	encU16((*p)[offUpper:offUpper+2], h.upper)          // offset 22: upper		(22-24) // 2 bytes
	// 													// offset 24: 			begin cellPtr list
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

// setNumCells encodes the provided value directly into the pageHeader.
func (p *Page) setNumCells(n uint16) {
	encU16((*p)[offNumCells:offNumCells+2], n)
}

// setLower encodes the provided value directly into the pageHeader.
func (p *Page) setLower(n uint16) {
	encU16((*p)[offLower:offLower+2], n)
}

// setUpper encodes the provided value directly into the pageHeader.
func (p *Page) setUpper(n uint16) {
	encU16((*p)[offUpper:offUpper+2], n)
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
		return ErrRecordTooSmall
	}
	// numFree := p.freeSpace() - pageCellSize
	// util.DEBUG("checkRecord, numFree=%d, prev=%d", numFree, prev)
	if size >= p.freeSpace() {
		return ErrNoRoom
	}
	return nil
}

// getCellPtrs decodes and returns a set of cellPtr pointer for this Page.
// It will return nil if there are no numCells on this Page. Any changes
// made to this list of numCells is not persisted until calling setSlots.
func (p *Page) getCellPtrs() []cellPtr {
	// Check if there are any numCells to return.
	numCells := decU16((*p)[offNumCells : offNumCells+2])
	if numCells < 1 {
		// No numCells to decode
		return nil
	}
	// We have numCells we can decode. Create a set we can append to.
	cells := make([]cellPtr, numCells, numCells)
	// Start looping, decoding, and adding numCells to our cellPtr set.
	for pos := uint16(0); pos < numCells; pos++ {
		// Append the cellPtr to the cellPtr set.
		cells[pos] = p.getCellPtrAt(pos)
	}
	// Finally, return our cellPtr set.
	return cells
}

// __setSlotSet encodes a set of cellPtr pointers into this Page. It will
// return an error if there is not enough room to write the set of numCells
// to the underlying Page.
func (p *Page) __setSlotSet(_ []*cellPtr) error {
	// Not sure if I want to append or overwrite this at the moment...
	return nil
}

// encCellPtr writes the provided cellPtr to the location derived using
// the supplied cellPtr index ID. encCellPtr panics if the provided CID
// is out of bounds.
func (p *Page) _encCellPtr(cp cellPtr, pos uint16) {
	// get the cellPtr offset
	off := pageHeaderSize + (pos * pageCellSize)
	// make sure it is in bounds
	if lo := decU16((*p)[offLower : offLower+2]); off > lo {
		panic(
			fmt.Sprintf(
				"--encCellPtr: cellPtr position (pos=%d, off=%d) is out of bounds (max=%d)", pos,
				off, lo,
			),
		)
	}
	// now we write the cellPtr to the Page
	// encU16((*p)[off:off+2], sl.id)
	// encU16((*p)[off+2:off+4], sl.flags)
	// encU16((*p)[off+4:off+6], sl.offset)
	// encU16((*p)[off+6:off+8], sl.length)
	encU64((*p)[off:off+8], uint64(cp))
}

// encCellPtr adds the provided cellPtr to our cell pointer set, sorts the
// set by the record key, and encodes the entire set back to the page.
func (p *Page) encCellPtr(cp cellPtr) {
	// First, get our set of cell pointers and append our newest cell pointer.
	// Check if there are any numCells to return.
	numCells := p.getNumCells()
	if numCells < 1 {
		// No numCells to decode
		panic("encCellPtr: something terrible happened")
	}
	// We have some cells we can decode. Create a set we can append to.
	var cellPtrs []cellPtr
	// Add our newest member to our set.
	cellPtrs = append(cellPtrs, cp)
	// Then iterate, decode, and adding the other cells to our cellPtr set.
	for pos := uint16(0); pos < numCells; pos++ {
		cellPtrs = append(cellPtrs, p.getCellPtrAt(pos))
	}
	if len(cellPtrs) > 2 {
		// Next, proceed with sorting our cell set.
		p.sortCellPtrs(cellPtrs)
	}
	// And finally, we can iterate our cell set once more, and encode them
	// back to the page in sorted order.
	for pos, c := range cellPtrs {
		off := pageHeaderSize + (pos * pageCellSize)
		encU64((*p)[off:off+8], uint64(c))
	}
}

func (p *Page) sortCellPtrs(cells []cellPtr) {
	sort.SliceStable(
		cells, func(i, j int) bool {
			if cells[i].hasFlag(C_USED) && cells[j].hasFlag(C_FREE) {
				return true
			}
			return bytes.Compare(p.getRecForCell(cells[i])[:4], p.getRecForCell(cells[j])[:4]) < 0
		},
	)
}

// getCellPtrAt reads and returns the cellPtr at the provided location in the
// cellPtr list derived using the supplied cellPtr position pos. This method
// will panic if the provided pos is out of the list bounds.
func (p *Page) getCellPtrAt(pos uint16) cellPtr {
	// Calculate where the proper offset would be for the provided position
	off := pageHeaderSize + (pos * pageCellSize)
	// Make sure it is in bounds
	if lo := decU16((*p)[offLower : offLower+2]); off > lo {
		panic(
			fmt.Sprintf(
				"--getCellPtrAt: cellPtr position (pos=%d, off=%d) is out of bounds (max=%d)", pos,
				off, lo,
			),
		)
	}
	return cellPtr(decU64((*p)[off : off+8]))
}

func (p *Page) getRecForCellPtrAt(pos uint16) []byte {
	// get the cellPtr offset
	off := pageHeaderSize + (pos * pageCellSize)
	// make sure it is in bounds
	if lo := decU16((*p)[offLower : offLower+2]); off > lo {
		panic(
			fmt.Sprintf(
				"--getCellPtrAt: cellPtr position (pos=%d, off=%d) is out of bounds (max=%d)", pos,
				off, lo,
			),
		)
	}
	// now we decode the record offset and length from the page
	recOff := decU16((*p)[off+4 : off+6])
	recLen := decU16((*p)[off+6 : off+8])
	return (*p)[recOff : recOff+recLen]
}

// addCellPtr appends a new cellPtr to the Page
// func (p *Page) addCellPtr(size uint16) (uint16, *cellPtr) {
// 	// get Page pageHeader
// 	h := p.getPageHeader()
// 	// grab the cellPtr id for later
// 	sid := h.numCells
// 	// update Page pageHeader
// 	h.numCells++
// 	h.lower += pageCellSize
// 	h.upper -= size
// 	// create new cellPtr structure
// 	sl := &cellPtr{
// 		id:     h.numCells,
// 		flags:  C_USED,
// 		offset: h.upper,
// 		length: size,
// 	}
// 	// write pageHeader back to Page
// 	p.setHeader(h)
// 	// encode cellPtr onto Page
// 	p.encCellPtr(sl, sid)
// 	// finally, return CID and cellPtr
// 	return sid, sl
// }

// getCellPtrByID attempts to locate and return the desired cellPtr by
// performing a scan through the cellPtr list and finding an ID that
// matches the one provided. This method will return the valid cellPtr
// or an error, but never both.
func (p *Page) getCellPtrByID(id uint16) (cellPtr, error) {
	// First, create a cellPtr instance we can return when we are finished.
	var cp cellPtr
	// Get the max number of usable cells in our list.
	numCells := p.getNumCells() // - p.getNumFree()
	// Start ranging our cell pointer list in search for a matching ID.
	for pos := uint16(0); pos < numCells; pos++ {
		// Attempt to get the cell at this location, then check the ID.
		cp = p.getCellPtrAt(pos)
		log.Printf(">>>>>> trying to find cell with id=%d, have={%s}\n", id, cp)
		if cp.getID() == id {
			// We found it, so we return it
			return cp, nil
		}
	}
	// We did not find it, so we will return an error
	return 0, ErrInvalidCID
}

// getRecForCell returns the record for the given cellPtr
func (p *Page) getRecForCell(cp cellPtr) []byte {
	beg, end := cp.bounds()
	return (*p)[beg:end]
}

// acquireCellPtr attempts to acquire and return a cellPtr that will fit the
// record size provided. It will first try to use an available cell that has
// been marked free, but if there are none, it will go on to add a new one.
// Either way, unless the page is full, this method should always return a
// usable cellPtr, along with a boolean indicating true if a free one was
// chosen to be recycled, and false if a new one had to be created.
func (p *Page) acquireCellPtr(size uint16) (cellPtr, bool) {
	// First, create a cellPtr instance we can return when we are finished.
	var cp cellPtr
	// Now, we will grab our cell count stats, and check to see if there are
	// any free cells we can check to see if they are candidates for use.
	total, free := p.getNumCells(), p.getNumFree()
	if free > 0 {
		// We have some free cells, let's check them out to see if there are
		// any we can re-cycle.
		for pos := total; pos > total-free; pos-- {
			// Attempt to get the cell at this location
			cp = p.getCellPtrAt(pos)
			// Check if the flags are properly set, and if the length is suitable
			if cp.hasFlag(C_FREE) && size <= cp.getLength() {
				// This cell pointer is free and fits our size criteria, so it can
				// be returned for use.
				return cp, true
			}
		}
	}
	// Otherwise, we will just return our new cell
	return cp, false
}

// AddRecord writes a new record to the Page. It returns a *RecID which is a record
// ID, along with any potential errors encountered.
func (p *Page) AddRecord(data []byte) (*RecID, error) {
	// Use our page latches
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// Get our record size
	rsize := uint16(len(data))
	// Perform sanity checks on our data
	err := p.checkRecord(rsize)
	if err != nil {
		return nil, err
	}
	// Get our page header for doing some updates.
	h := p.getPageHeader()
	// Now, we must get a new (or free to use) cellPtr to index the record
	cp, usedFree := p.acquireCellPtr(rsize)
	// Next, update some header values based on what we found.
	if usedFree {
		// We were able to use up one of our free cells, so now we will update
		// the page header and cellPtr accordingly.
		h.numFree--
		cp.setFlags(C_USED)
		cp.setLength(rsize)
	} else {
		// Otherwise, despite having at least one free cellPtr, we were not able
		// to locate one we can re-use. So first, we will have to update our header
		// and cellPtr accordingly.
		h.numCells++
		h.lower += pageCellSize
		h.upper -= rsize
		cp = newCell(h.numCells, h.upper, rsize)
	}
	// Now, we can encode our updated header to the page.
	p.setPageHeader(h)
	// Get our record bounds from the cellPtr index
	beg, end := cp.bounds()
	// Write record data to the Page
	copy((*p)[beg:end], data)
	// Encode and sort our cell pointer set.
	p.encCellPtr(cp)
	// Assemble and return the record ID
	return &RecID{
		PID: p.getPageID(),
		CID: cp.getID(),
	}, nil
}

func (p *Page) iterateCells(fn func(pos uint16, cp cellPtr, rec []byte) bool) {
	cells := p.getNumCells()
	for pos := uint16(0); pos < cells; pos++ {
		cp := p.getCellPtrAt(pos)
		beg, end := cp.bounds()
		if !fn(pos, cp, (*p)[beg:end]) {
			break
		}
	}
}

// checkRID performs error and sanity checking on the provided
// record ID.
func (p *Page) checkRID(rid *RecID) error {
	log.Printf("rid.PID=%d, p.getPageID=%d\n", rid.PID, p.getPageID())
	if rid.PID != p.getPageID() {
		return ErrInvalidPID
	}
	// if rid.CID > p.getNumCells() {
	// 	return ErrInvalidCID
	// }
	return nil
}

// GetRecord reads a record from the Page. It returns the record data
// that is associated with the provided record ID, along with any
// potential errors encountered.
func (p *Page) GetRecord(rid *RecID) ([]byte, error) {
	// Use our page latch, for safety.
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// Sanity check the record ID
	err := p.checkRID(rid)
	if err != nil {
		return nil, err
	}
	// Locate the associated cellPtr
	cp, err := p.getCellPtrByID(rid.CID)
	if err != nil {
		return nil, err
	}
	if cp.hasFlag(C_FREE) {
		// This cell ptr is marked as a free slot, but we are not looking
		// for a free slot, so return a record not found error.
		return nil, ErrRecordNotFound
	}
	// create a buffer to copy the record into (safety)
	buff := make([]byte, cp.getLength())
	// get the record boundary from the cellPtr
	beg, end := cp.bounds()
	// copy the record into the buffer (using record bounds)
	copy(buff, (*p)[beg:end])
	// return the record copy
	return buff, nil
}

// delSlot updates the numFree of the cellPtr found at the provided
// cellPtr ID, and returns the cellPtr for use in the delete operation.
func (p *Page) delSlot(sid uint16) (cellPtr, error) {
	// get the cellPtr using the CID
	cp := p.getCellPtrAt(sid)
	// if the cellPtr numFree is numFree, return nil
	if cp.hasFlag(C_FREE) {
		// This cell ptr is marked as a free slot, but we are not looking
		// for a free slot, so return a record not found error.
		return 0, ErrRecordNotFound
	}
	// update cellPtr numFree
	cp.setFlags(C_FREE)
	// save the numFree of the found cellPtr
	p.encCellPtr(cp)
	// and return
	return cp, nil
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
	cp, err := p.delSlot(rid.CID)
	if err != nil {
		return err
	}
	// create a buffer to overwrite the record with
	buff := make([]byte, cp.getLength())
	// get the record bounds from the cellPtr index
	beg, end := cp.bounds()
	// util.DEBUG("--delRecord(%d) [%x] del_slot=%d, slot_count=%d", id.CID, (*p)[beg:end], id.CID, len(p.getCellPtrs()))
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
	slots    []cellPtr
	index    int
	skipFree bool
}

// newIter instantiates and returns a new iterator. If the Page contains
// no data entries, then it returns a nil iterator along with an error.
func (p *Page) newIter(skipFree bool) (*iterator, error) {
	cells := p.getCellPtrs()
	if cells == nil {
		return nil, ErrEmptyPage
	}
	return &iterator{
		slots:    cells,
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
	cp := it.slots[it.index]
	// check if we should skip any numCells marked numFree.
	if it.skipFree && cp.hasFlag(C_FREE) {
		// cellPtr is numFree, skip it
		return it.next()
	}
	// return our cellPtr
	return &cp
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
		sl := p.getCellPtrAt(sid)
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
