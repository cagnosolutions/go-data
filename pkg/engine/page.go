package engine

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"unsafe"
)

var (
	encU16 = binary.LittleEndian.PutUint16
	encU32 = binary.LittleEndian.PutUint32
	encU64 = binary.LittleEndian.PutUint64
	decU16 = binary.LittleEndian.Uint16
	decU32 = binary.LittleEndian.Uint32
	decU64 = binary.LittleEndian.Uint64
)

// page latch
var pgLatch sync.Mutex

// PageID represents a page ID
type PageID = uint32

// CellID represents a cellptr ID (not position)
type CellID = uint16

// RecordID is a record ID type used by a page and is associated
// with a record.
type RecordID struct {
	PageID
	CellID
}

// String is the stringer method for a RecordID
func (r *RecordID) String() string {
	return fmt.Sprintf("{%d, %d}", r.PageID, r.CellID)
}

const (
	// P_FREE and P_USED are flags to be used with the page
	P_FREE uint32 = 0x00000001 // indicates the page is free to use
	P_USED uint32 = 0x00000002 // indicates the page is being used
	P_NODE uint32 = 0x00000010 // indicates the page is an internal node
	P_LEAF uint32 = 0x00000020 // indicates the page is a leaf
	P_ROOT uint32 = 0x00000050 // indicates the page is a root node

	// constants for the page, headers, cellptrs and record sizes
	PageSize         = 16 << 10
	pageHeaderSize   = 24
	pageCellPtrSize  = 8
	recordHeaderSize = 6

	// offsets to be used for decoding and encoding the page header
	offPID      uint32 = 0  // pid=uint32		offs=0-4 	(4 bytes)
	offPrev     uint32 = 4  // prev=uint32 		offs=4-8	(4 bytes)
	offNext     uint32 = 8  // next=uint32 		offs=8-12	(4 bytes)
	offFlags    uint32 = 12 // flags=uint32		offs=12-16	(4 bytes)
	offNumCells uint16 = 16 // numCells=uint16	offs=16-18	(2 bytes)
	offNumFree  uint16 = 18 // numFree=uint16	offs=18-20	(2 bytes)
	offLower    uint16 = 20 // lower=uint16		offs=20-22 	(2 bytes)
	offUpper    uint16 = 22 // upper=uint16		offs=22-24	(2 bytes)
)

// pageHeader represents the header of a page.
type pageHeader struct {
	ID    uint32 // id of page
	Prev  uint32 // prev page pointer
	Next  uint32 // next page pointer
	Flags uint32 // flags and meta data for page
	Cells uint16 // number of cells allocated
	Free  uint16 // number of cells that are free
	Lower uint16 // lower numFree space bound
	Upper uint16 // upper numFree space bound
}

// Size returns the size of the pageHeader in bytes.
func (h *pageHeader) Size() int {
	return int(unsafe.Sizeof(*h))
}

// String is the stringer method for the pageHeader.
func (h *pageHeader) String() string {
	b, err := json.MarshalIndent(h, "", "")
	if err != nil {
		panic("pageHeader:" + err.Error())
	}
	return string(b)
}

// page represents a page
type page []byte

// newPage returns a new page
func newPage(id uint32, flags uint32) page {
	p := make(page, PageSize, PageSize)
	p.setPageHeader(
		&pageHeader{
			ID:    id,
			Prev:  0,
			Next:  0,
			Flags: flags,
			Cells: 0,
			Free:  0,
			Lower: pageHeaderSize,
			Upper: PageSize,
		},
	)
	return p
}

// getPageHeader decodes and returns a pointer to the pageHeader
// directly from the page.
func (p *page) getPageHeader() *pageHeader {
	return &pageHeader{
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

// setPageHeader takes a pointer to a pageHeader and encodes it
// directly into the page.
func (p *page) setPageHeader(h *pageHeader) {
	encU32((*p)[offPID:offPID+4], h.ID)
	encU32((*p)[offPrev:offPrev+4], h.Prev)
	encU32((*p)[offNext:offNext+4], h.Next)
	encU32((*p)[offFlags:offFlags+4], h.Flags)
	encU16((*p)[offNumCells:offNumCells+2], h.Cells)
	encU16((*p)[offNumFree:offNumFree+2], h.Free)
	encU16((*p)[offLower:offLower+2], h.Lower)
	encU16((*p)[offUpper:offUpper+2], h.Upper)
}

// getRecordKeyUsingCellPos takes the position of a cellptr and uses it
// to decode and return the associated record key.
func (p *page) getRecordKeyUsingCellPos(pos uint16) []byte {
	// Decode the cellptr at the provided location.
	cp := p.decCell(pos)
	// Get the record bounds from the decoded cellptr.
	beg, end := cp.getBounds()
	// Return the record that the cellptr points to.
	r := record((*p)[beg:end])
	return r.Key()
}

// getRecordUsingCellPos takes the position of a cellptr and uses it
// to decode and return the associated record.
func (p *page) getRecordUsingCellPos(pos uint16) record {
	// Decode the cellptr at the provided location.
	cp := p.decCell(pos)
	// Get the record bounds from the decoded cellptr.
	beg, end := cp.getBounds()
	// Return the record that the cellptr points to.
	return record((*p)[beg:end])
}

// getRecordKeyUsingCell takes a cellptr and uses it to decode and return
// the associated record key.
func (p *page) getRecordKeyUsingCell(c cellptr) []byte {
	// Get the record bounds from the provided cellptr.
	beg, end := c.getBounds()
	// Return the record that the cellptr points to.
	r := record((*p)[beg:end])
	return r.Key()
}

// getRecordUsingCell takes a cellptr and uses it to decode and return
// the associated record.
func (p *page) getRecordUsingCell(c cellptr) record {
	// Get the record bounds from the provided cellptr.
	beg, end := c.getBounds()
	// Return the record that the cellptr points to.
	return record((*p)[beg:end])
}

// addCell takes the length of a record and creates and returns a new
// cellptr that can be used to represent the record location. The cellptr
// that is returned is encoded, and all operations requiring the pageHeader
// to be updated are all taken care of withint this method before the
// cellptr is returned.
func (p *page) addCell(size uint16) cellptr {
	// Increment the cell count, as well as the lower boundary accordingly,
	// as well as decrementing the upper boundary by the record size.
	p.incrNumCells(1)
	p.incrLower(pageCellPtrSize)
	p.decrUpper(size)
	// Create a new cell, and promptly re-encode it before returning the cell.
	c := newCell(p.getNumCells(), p.getUpper(), size)
	p.encCell(c, p.getNumCells()-1)
	return c
}

// delCell takes a cellptr, and a position and re-encodes the provided cellptr
// as one that can be re-used. It uses the provided position to move it as close
// to the lower boundary as possible to aid in calls to vacuum the page later on.
func (p *page) delCell(c cellptr, pos uint16) {
	// Check the position to ensure it is correct, otherwise panic.
	if pos > p.getNumCells() {
		panic("error: cell position out of bounds")
	}
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

// encCell takes a cellptr and a position and attempts to encode the provided
// cellptr at the provided position. It panics if anything doesn't work.
func (p *page) encCell(c cellptr, pos uint16) {
	// Get the offset from the provided position.
	off := pageHeaderSize + (pos * pageCellPtrSize)
	// Check the offset to ensure it is correct, otherwise panic.
	if off < pageHeaderSize || off > p.getLower() {
		panic("error: cell position out of bounds")
	}
	// Check the cellptr to ensure it is valid, otherwise panic.
	if !c.isValid() {
		panic("error: cell is not a valid cell")
	}
	// Encode the cellptr at the offset.
	encU64((*p)[off:off+8], uint64(c))
}

// decCell takes a position and attempts to decode and return a cellptr
// at the provided position. It panics if anything doesn't work.
func (p *page) decCell(pos uint16) cellptr {
	// Get the offset from the provided position.
	off := pageHeaderSize + (pos * pageCellPtrSize)
	// Check the offset to ensure it is correct, otherwise panic.
	if off < pageHeaderSize || off > p.getLower() {
		panic("error: cell position out of bounds")
	}
	// Decode the cellptr at the offset.
	var cp cellptr
	cp = cellptr(decU64((*p)[off : off+8]))
	// Check the cellptr to ensure it is valid, otherwise panic.
	if !cp.isValid() {
		panic("error: cell is not a valid cell")
	}
	return cp
}

// makeRecordID creates and returns a pointer to a RecordID using the
// provided cellptr.
func (p *page) makeRecordID(c cellptr) *RecordID {
	return &RecordID{
		PageID: PageID(p.getPageID()),
		CellID: CellID(c.getID()),
	}
}

// recycleCell attempts to reuse a free cellptr for a record, if there is a candidate that
// works well. It returns the used cellptr, and a boolean indicating true if it succeeded in
// recycling the cellptr, and false if it could not recycle one.
func (p *page) recycleCell(r record) cellptr {
	// We do, so let's see if we have any candidates for recycling.
	var cp cellptr
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		// for pos := numCells - 1; pos > numCells-freeCells-1; pos-- {
		// Get the free cell at the first location
		cp = p.decCell(pos)
		// Check the cell to see if it's a candidate
		if cp.hasFlag(C_FREE) && cp.canFit(uint16(len(r))) {
			// We have found ourselves a candidate, now we just need to
			// update the cell, and re-encode it.
			cp.setFlags(C_USED)
			cp.setLength(uint16(len(r)))
			p.encCell(cp, pos)
			// And decrement the free cell count in the page header.
			p.decrNumFree(1)
			// Now, we can return the cell
			return cp
		}
	}
	// Otherwise, we could not find one to recycle, so we return 0 because we will be
	// checking this cell to ensure that it reads as a valid one, and 0 will not, so
	// it is perfectly safe to return.
	return 0
}

// checkRecord performs some error checking on the record to ensure it is a good
// record, and that we also have plenty of room for it in the page.
func (p *page) checkRecord(r record) error {
	if r == nil {
		return ErrRecordTooSmall
	}
	if int(p.getUpper()-p.getLower())-pageCellPtrSize < len(r) {
		return ErrNoRoom
	}
	return nil
}

func (p *page) checkRecordID(id *RecordID) error {
	if id.PageID != PageID(p.getPageID()) {
		return ErrInvalidPID
	}
	if id.CellID > p.getNumCells() {
		return ErrInvalidSID
	}
	return nil
}

// addRecord takes a record and attempts to write it to the page. First it will check
// to see if there are any cellptrs that can be re-used, and do so if there are any
// that will accommodate the record size. Otherwise, it will create a new cellptr.
// The cellptrs are always sorted according to the record key, and the record data
// is written to the page last.
func (p *page) addRecord(r record) (*RecordID, error) {
	// latch
	// pgLatch.Lock()
	// defer pgLatch.Unlock()
	// Check the record data to ensure it is not empty, and that we have
	// enough room to add it.
	err := p.checkRecord(r)
	if err != nil {
		return nil, err
	}
	// Get our free cells, and our total cell count.
	freeCells := p.getNumFree()
	// Allocate our cell pointer, we will need to use one no matter what.
	var cp cellptr
	// Before continuing, we must check to see if we can re-use any cells.
	if freeCells > 0 {
		// We do, so let's see if we have any candidates for recycling.
		pgLatch.Lock()
		cp = p.recycleCell(r)
		pgLatch.Unlock()
		// We will be checking below to see if the cell pointer is valid,
		// and it will only be valid if we had a successful time recycling
		// in here, so no need to do anything else, just proceed.
	}
	// This check basically tells us if we have recycled a cell or not.
	if !cp.isValid() {
		// No valid cell pointer found, which means we did not recycle any,
		// and we are free to allocate a fresh one. So that is what we do.
		pgLatch.Lock()
		cp = p.addCell(uint16(len(r)))
		pgLatch.Unlock()
	}
	// We want to ensure we write the record data to the page before we
	// check or try to sort.
	pgLatch.Lock()
	copy((*p)[cp.getOffset():cp.getOffset()+cp.getLength()], r)
	pgLatch.Unlock()
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
		pgLatch.Lock()
		sort.Sort(p)
		pgLatch.Unlock()
	}
	// And finally, return our RecordID
	return &RecordID{p.getPageID(), cp.getID()}, nil
}

// getRecord takes a RecordID, and attempts to locate a cellptr that matches.
// If a match can be located, then the resulting record is returned.
func (p *page) getRecord(id *RecordID) (record, error) {
	// Error check the record ID
	err := p.checkRecordID(id)
	if err != nil {
		return nil, err
	}
	var cp cellptr
	// First, we will attempt to locate the record.
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		// Check the cell at the provided position.
		pgLatch.Lock()
		cp = p.decCell(pos)
		pgLatch.Unlock()
		if cp.getID() == id.CellID {
			// We have located the record. Let's check to make sure it has
			// not been deleted.
			if cp.hasFlag(C_FREE) {
				return nil, ErrRecordNotFound
			}
			pgLatch.Lock()
			// It has not, so we can fetch the record.
			r := p.getRecordUsingCell(cp)
			// We should make a copy of it, so we do not mutate the original.
			rc := make(record, len(r), len(r))
			copy(rc, r)
			pgLatch.Unlock()
			// Return the record
			return rc, nil
		}
	}
	// Otherwise, we did not locate the record
	return nil, ErrRecordNotFound
}

// delRecord attempts to delete a record using the provided record ID. The
// associated cellptr will be marked as free to re-use, and the record data
// will be overwritten. Any errors will be returned.
func (p *page) delRecord(id *RecordID) error {
	// Error check the record ID
	err := p.checkRecordID(id)
	if err != nil {
		return err
	}
	// First, we will create our cell pointer variable for later.
	var cp cellptr
	// Then, we will attempt to locate the record.
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		// Check the cell at the provided position
		pgLatch.Lock()
		cp = p.decCell(pos)
		pgLatch.Unlock()
		if cp.getID() == id.CellID {
			// We have located the record. Let's check to make sure it has
			// not been deleted.
			if cp.hasFlag(C_USED) {
				// We have located our used record.
				// First, lock, then we can set it free.
				pgLatch.Lock()
				// Get our record boundaries
				beg, end := cp.getBounds()
				// Overwrite the old record
				copy((*p)[beg:end], make([]byte, cp.getLength()))
				// mark the cell free
				cp.setFlags(C_FREE)
				// get the "window" of all the cells
				cells := (*p)[pageHeaderSize:p.getLower()]
				// slide all the cells in the window down
				copy(cells[pos*8:], cells[(pos+1)*8:])
				// re-encode the "deleted" cell, at the front of the list
				p.encCell(cp, p.getNumCells()-1)
				// And increment the free cell count in the page header.
				p.incrNumFree(1)
				// Finally, unlock and return
				pgLatch.Unlock()
				return nil
			}
		}
	}
	// Otherwise, we could not locate the record.
	return ErrRecordNotFound
}

// getRecordByKey attempts to locate and return a record using the provided
// record key. It performs a binary search, since the record cellptrs are
// always kept in a sorted order, attempts to return a matching record. If
// a there is more than one record in the page that has the same key then
// it will return the first one it locates.
func (p *page) getRecordByKey(key []byte) *record {
	// latch
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// First, we will attempt to locate the record.
	pos := p.findCellPos(key)
	// Otherwise, we have located it. Let's check to make sure it has
	// not been deleted.
	c := p.decCell(uint16(pos))
	if c.hasFlag(C_FREE) {
		return nil
	}
	// It has not, so we will get it.
	r := p.getRecordUsingCell(c)
	// We should make a copy of it, so we do not mutate the original.
	rc := make(record, len(r), len(r))
	copy(rc, r)
	return &rc
}

// hasKey returns a boolean indicating true if the key that is provided
// is found within the current page, and false if it is not found. The
// key must be strictly equal.
func (p *page) hasKey(k []byte) bool {
	// latch
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// create local cell pointer variable
	var cp cellptr
	// loop through the used cells
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		cp = p.decCell(pos)
		if cp.hasFlag(C_FREE) {
			continue
		}
		if bytes.Equal(k, p.getRecordKeyUsingCell(cp)) {
			return true
		}
	}
	return false
}

var skipRecord = errors.New("skip this record")

// rangeRecords is an iterator methods that uses a simple callback. It
// returns any errors encountered.
func (p *page) rangeRecords(fn func(r *record) error) error {
	// latch
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// iterate
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		c := p.decCell(pos)
		if c.hasFlag(C_FREE) {
			continue
		}
		r := p.getRecordUsingCell(c)
		if err := fn(&r); err != nil {
			if err == skipRecord {
				continue
			}
			return err
		}
	}
	return nil
}

// rangeNRecords is a bounded iterator methods that uses a simple callback. It
// returns any errors encountered.
func (p *page) rangeNRecords(beg, end int, fn func(r *record) error) error {
	// latch
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// error check
	if beg < 0 {
		beg = 0
	}
	if beg > int(p.getNumCells()) {
		beg = int(p.getNumCells() - 1)
	}
	if end < beg {
		end = beg + 1
	}
	if end > int(p.getNumCells()) {
		end = int(p.getNumCells())
	}
	// iterate
	for pos := uint16(beg); pos < uint16(end); pos++ {
		c := p.decCell(pos)
		if c.hasFlag(C_FREE) {
			continue
		}
		r := p.getRecordUsingCell(c)
		if err := fn(&r); err != nil {
			if err == skipRecord {
				continue
			}
			return err
		}
	}
	return nil
}

// Len implements the sort.Sort interface for sorting the cellptrs
// according to the record key.
func (p *page) Len() int {
	return int(p.getNumCells() - p.getNumFree())
}

// Less implements the sort.Sort interface for sorting the cellptrs
// according to the record key.
func (p *page) Less(i, j int) bool {
	r1 := p.getRecordUsingCellPos(uint16(i))
	r2 := p.getRecordUsingCellPos(uint16(j))
	return bytes.Compare(r1.Key(), r2.Key()) < 0
}

// Swap implements the sort.Sort interface for sorting the cellptrs
// according to the record key.
func (p *page) Swap(i, j int) {
	cp1 := p.decCell(uint16(i))
	cp2 := p.decCell(uint16(j))
	p.encCell(cp1, uint16(j))
	p.encCell(cp2, uint16(i))
}

// findCellPos performs a binary search through the cellptrs using
// the provided record key, and attempts to find the first cellptr
// that contains a record key that matches.
func (p *page) findCellPos(k []byte) uint16 {
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
	return uint16(i) // , i < n && at == 0
}

// clear resets the entire page. It wipes all the data, but retains the same ID.
func (p *page) clear() {
	// latch
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// clear the page out
	*p = newPage(p.getPageID(), P_FREE)
}

// Vacuum is a method that sucks up any free space within the page, removing any
// gaps, and essentially compacting the page, so it can be better utilized if it
// is getting full. This method must be called manually.
func (p *page) Vacuum() {
	// latch
	pgLatch.Lock()
	defer pgLatch.Unlock()
	// First, we must allocate a new page to copy data into.
	np := newPage(p.getPageID(), p.getFlags())
	// We will initialize local states here, so we only have to set them once.
	numCells, lowerBound, upperBound := uint16(0), uint16(pageHeaderSize), uint16(PageSize)
	// Next we iterate the current non-free cells and add the records to the new page.
	for pos := uint16(0); pos < p.getNumCells(); pos++ {
		// Get the cell for the current position.
		c := p.decCell(pos)
		if c.hasFlag(C_USED) {
			// We will only add valid, used, records. So we need to grab the record.
			r := p.getRecordUsingCell(c)
			// Create a new cell in the new page for adding the record. Don't forget
			// to update the local page states for the new page, and then we can set
			// them once, outside the loop.
			numCells++
			lowerBound += pageCellPtrSize
			upperBound -= uint16(len(r))
			// Create a new cell, and promptly re-encode it in the new page (and take
			// note that we are using the old cellID)
			nc := newCell(c.getID(), upperBound, uint16(len(r)))
			// Calculate offset (we cannot use encCell here, because we are optimizing
			// for batch inserting) so we can manually encode.
			off := pageHeaderSize + ((numCells - 1) * pageCellPtrSize)
			encU64(np[off:off+8], uint64(nc))
			// Write the old record data to the new page, using the new cell.
			copy(np[nc.getOffset():nc.getOffset()+nc.getLength()], r)
		}
		// Otherwise, we are not writing a record, if the cell is not being used.
	}
	// Don't forget to update the local page state in the new page.
	np.setNumCells(numCells)
	np.setLower(lowerBound)
	np.setUpper(upperBound)
	// Then, we will sort the cells according to their key, just one time
	// at the very end.
	sort.Sort(&np)
	// And now we are finished compacting everything, so we will swap the page
	// pointers.
	*p = np
	// We will call the GC directly here, and return
	runtime.GC()
}

// getPageID decodes and returns the page ID directly from the encoded pageHeader.
func (p *page) getPageID() uint32 {
	return decU32((*p)[offPID : offPID+4])
}

// getPrev decodes and returns the previous pointer directly from the encoded pageHeader.
func (p *page) getPrev() uint32 {
	return decU32((*p)[offPrev : offPrev+4])
}

// getNext decodes and returns the next pointer directly from the encoded pageHeader.
func (p *page) getNext() uint32 {
	return decU32((*p)[offNext : offNext+4])
}

// getFlags decodes and returns the flags field directly from the encoded pageHeader.
func (p *page) getFlags() uint32 {
	return decU32((*p)[offFlags : offFlags+4])
}

// hasFlag tests if the cell pointer has a flag set
func (p *page) hasFlag(flag uint32) bool {
	return decU32((*p)[offFlags:offFlags+4])&flag != 0
}

// getNumFree decodes and returns the number of free cells directly from the encoded pageHeader.
func (p *page) getNumFree() uint16 {
	return decU16((*p)[offNumFree : offNumFree+2])
}

// getNumCells decodes and returns the number of allocated cells directly from the encoded pageHeader.
func (p *page) getNumCells() uint16 {
	return decU16((*p)[offNumCells : offNumCells+2])
}

// getLower decodes and returns the lower bound marker directly from the encoded pageHeader.
func (p *page) getLower() uint16 {
	return decU16((*p)[offLower : offLower+2])
}

// getUpper decodes and returns the upper bound marker directly from the encoded pageHeader.
func (p *page) getUpper() uint16 {
	return decU16((*p)[offUpper : offUpper+2])
}

// setPageID encodes the provided value directly into the pageHeader.
func (p *page) setPageID(n uint32) {
	encU32((*p)[offPID:offPID+4], n)
}

// setPrev encodes the provided value directly into the pageHeader.
func (p *page) setPrev(n uint32) {
	encU32((*p)[offPrev:offPrev+4], n)
}

// setNext encodes the provided value directly into the pageHeader.
func (p *page) setNext(n uint32) {
	encU32((*p)[offNext:offNext+4], n)
}

// setFlags encodes the provided value directly into the pageHeader.
func (p *page) setFlags(n uint32) {
	encU32((*p)[offFlags:offFlags+4], n)
}

// setNumFree encodes the provided value directly into the pageHeader.
func (p *page) setNumFree(n uint16) {
	encU16((*p)[offNumFree:offNumFree+2], n)
}

// incrNumFree increments the free cell count by the amount provided and encodes
// directly into the pageHeader.
func (p *page) incrNumFree(n uint16) {
	encU16((*p)[offNumFree:offNumFree+2], decU16((*p)[offNumFree:offNumFree+2])+n)
}

// decrNumFree decrements the free cell count by the amount provided and encodes
// directly into the pageHeader.
func (p *page) decrNumFree(n uint16) {
	encU16((*p)[offNumFree:offNumFree+2], decU16((*p)[offNumFree:offNumFree+2])-n)
}

// setNumCells encodes the provided value directly into the pageHeader.
func (p *page) setNumCells(n uint16) {
	encU16((*p)[offNumCells:offNumCells+2], n)
}

// incrNumCells increments the cell count by the amount provided and encodes
// directly into the pageHeader.
func (p *page) incrNumCells(n uint16) {
	encU16((*p)[offNumCells:offNumCells+2], decU16((*p)[offNumCells:offNumCells+2])+n)
}

// decrNumCells decrements the cell count by the amount provided and encodes
// directly into the pageHeader.
func (p *page) decrNumCells(n uint16) {
	encU16((*p)[offNumCells:offNumCells+2], decU16((*p)[offNumCells:offNumCells+2])-n)
}

// setLower encodes the provided value directly into the pageHeader.
func (p *page) setLower(n uint16) {
	encU16((*p)[offLower:offLower+2], n)
}

// incrLower increments the lower boundary by the amount provided and encodes
// directly into the pageHeader.
func (p *page) incrLower(n uint16) {
	encU16((*p)[offLower:offLower+2], decU16((*p)[offLower:offLower+2])+n)
}

// decrLower decrements the lower boundary by the amount provided and encodes
// directly into the pageHeader.
func (p *page) decrLower(n uint16) {
	encU16((*p)[offLower:offLower+2], decU16((*p)[offLower:offLower+2])-n)
}

// setUpper encodes the provided value directly into the pageHeader.
func (p *page) setUpper(n uint16) {
	encU16((*p)[offUpper:offUpper+2], n)
}

// incrUpper increments the upper boundary by the amount provided and encodes
// directly into the pageHeader.
func (p *page) incrUpper(n uint16) {
	encU16((*p)[offUpper:offUpper+2], decU16((*p)[offUpper:offUpper+2])+n)
}

// decrUpper increments the upper boundary by the amount provided and encodes
// directly into the pageHeader.
func (p *page) decrUpper(n uint16) {
	encU16((*p)[offUpper:offUpper+2], decU16((*p)[offUpper:offUpper+2])-n)
}

// getFreeSpace returns the free space left in the page
func (p *page) getFreeSpace() uint16 {
	free := p.getUpper() - p.getLower()
	if p != nil && int(free) > len(*p) {
		return 0
	}
	return free
}

// size returns the page size
func (p *page) size() int {
	return len(*p)
}

// String is the stringer method for the page
func (p *page) String() string {
	ss := fmt.Sprintf("%10v +---------+\n", "")
	ss += fmt.Sprintf("%10v |%2v%v%-2v|\n", "", "", " Page", "")
	ss += fmt.Sprintf("%10v +---------+\n", "")
	h := p.getPageHeader()
	hv := reflect.ValueOf(h).Elem()
	for i := 0; i < hv.NumField(); i++ {
		sf := hv.Type().Field(i)
		ss += fmt.Sprintf("%10v |%2v%5v%-2v|\n", strings.ToLower(sf.Name), "", hv.Field(i), "")
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

const (
	// C_FREE C_USED and C_MAGK are flags for the cellptrs
	C_MAGK uint16 = 0x5a00

	C_FREE uint16 = 0x0001
	C_USED uint16 = 0x0002

	C_PTR uint16 = 0x0010

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

func (c *cellptr) fmtFlag() string {
	if c.hasFlag(C_USED) {
		return "USED"
	}
	if c.hasFlag(C_FREE) {
		return "FREE"
	}
	return ""
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

const (
	R_BTND = 0x0100 // b+tree node (root or inner)
	R_BTLN = 0x0200 // b+tree leaf node
	R_NUM  = 0x01   // number type
	R_STR  = 0x02   // string type
	R_PTR  = 0x04   // pointer type
)

func checkRecordFlags(rtyp, ktyp, vtyp uint16) uint16 {
	if rtyp&0x0f00 == 0 {
		panic("bad record type flag")
	}
	if ktyp&0x000f == 0 || ktyp&0x000f == R_PTR {
		panic("bad key type flag")
	}
	if (vtyp<<4)&0x00f0 == 0 {
		panic("bad val type flag")
	}
	if rtyp&0x0f00 == R_BTND && vtyp != R_PTR {
		panic("bad val type flag, must have use a pointer flag with this record type")
	}
	var f uint16
	f |= rtyp
	f |= (ktyp << 0)
	f |= (vtyp << 4)
	return f
}

func main() {
	r := checkRecordFlags(R_BTND, R_NUM, R_PTR)
	ftox(r)
}

func ftox(f uint16) {
	var ss string
	var mask uint16
	ss += fmt.Sprintf("got flag: '0x%.4x'\n", f)
	for i := 4; i > 0; i -= 2 {
		switch i {
		case 4:
			mask = 0xff00
		case 2:
			mask = 0x00ff
		}
		ss += fmt.Sprintf("bits %d-%d: ", i-2, i)
		ss += fmt.Sprintf("0x%.2x", (f & mask))[:4]
		ss += "\n"
	}
	fmt.Println(ss)
}

// record flags
const (
	R_KEYREC = 0x00010000
	R_KEYVAL = 0x00020000

	RK_NUM = 0x00000001 //	00000000000000000000000000000001
	RK_STR = 0x00000002 //	00000000000000000000000000000010
	// RK_PTR = 0x00000004 // 	00000000000000000000000000000100
	// RK_CUS = 0x00000008 // 	00000000000000000000000000001000

	RV_NUM = 0x00000010 // 	00000000000000000000000000010000
	RV_STR = 0x00000020 // 	00000000000000000000000000100000
	RV_PTR = 0x00000040 // 	00000000000000000000000001000000
	// RV_CUS = 0x00000080 //  00000000000000000000000010000000

	PTR_RENT = 0x00000100 // 00000000000000000000000100000000
	PTR_CHLD = 0x00000200 // 00000000000000000000001000000000
	PTR_PREV = 0x00000400 // 00000000000000000000010000000000
	PTR_NEXT = 0x00000800 // 00000000000000000000100000000000

	keyMask = 0x0000000f //  00000000000000000000000000001111
	valMask = 0x000000f0 //  00000000000000000000000011110000
	ptrMask = 0x00000f00 //  00000000000000000000111100000000
	cusMask = 0x0000f000 //  00000000000000001111000000000000
)

// checkRecordFlags checks the flags that are set in the
// recordHeader to see if there are any that are set that
// should not be set or that have incorrect values.
func checkRecordFlags(flags uint32) error {
	if (flags&keyMask)&RK_NUM == 0 &&
		(flags&keyMask)&RK_STR == 0 {
		// (((flags&keyMask)&RK_PTR != 0) && (flags&ptrMask == 0)) {
		return ErrBadKeyFlagInRecord
	}
	if (flags&valMask)&RV_NUM == 0 &&
		(flags&valMask)&RV_STR == 0 &&
		(((flags&valMask)&RV_PTR != 0) && (flags&ptrMask == 0)) {
		return ErrBadValFlagInRecord
	}
	// if (flags&cusMask != 0) &&
	// 	(((flags&keyMask)&RK_CUS == 0) ||
	// 		((flags&valMask)&RV_CUS == 0)) {
	// 	return ErrBadCustomFlagInRecord
	// }
	return nil
}

// recordHeader is a pageHeader struct for encoding and
// decoding information for a record
type recordHeader struct {
	Flags  uint32
	KeyLen uint16
	ValLen uint16
}

// record is a binary type
type record []byte

// newRecord initiates and returns a new record using the flags provided
// as the indicators of what types the keys and values should hold.
func newRecord(flags uint32, key, val []byte) record {
	err := checkRecordFlags(flags)
	if err != nil {
		panic(err)
	}
	rsz := recordHeaderSize + len(key) + len(val)
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			Flags:  flags,
			KeyLen: uint16(len(key)),
			ValLen: uint16(len(val)),
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+n:], val)
	return rec
}

// func newPtrRecord(kind uint32, ptr uint32) record {
// 	var flags uint16
// 	if kind == PTR_RENT ||
// 		kind == PTR_CHLD ||
// 		kind == PTR_PREV ||
// 		kind == PTR_NEXT {
// 		flags = kind | RK_PTR | RV_PTR
// 	}
// 	rsz := recordHeaderSize + 4 + 4
// 	rec := make(record, rsz, rsz)
// 	rec.encRecordHeader(
// 		&recordHeader{
// 			Flags:  flags,
// 			KeyLen: 4,
// 			ValLen: 4,
// 		},
// 	)
// 	err := checkRecordFlags(flags)
// 	if err != nil {
// 		panic(err)
// 	}
// 	encU32(rec[recordHeaderSize:recordHeaderSize+4], ptr)
// 	encU32(rec[recordHeaderSize+4:recordHeaderSize+8], ptr)
// 	return rec
// }

// newUintUintRecord creates and returns a new record that uses uint32's
// as keys and uint32's as values.
func _(key uint32, val uint32) record {
	rsz := recordHeaderSize + 4 + 4
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			Flags:  RK_NUM | RV_NUM,
			KeyLen: 4,
			ValLen: 4,
		},
	)
	encU32(rec[recordHeaderSize:], key)
	encU32(rec[recordHeaderSize+4:recordHeaderSize+8], val)
	return rec
}

// newUintCharRecord creates and returns a new record that uses uint32's
// as keys and []byte slices as values.
func _(key uint32, val []byte) record {
	rsz := recordHeaderSize + 4 + len(val)
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			Flags:  RK_NUM | RV_STR,
			KeyLen: 4,
			ValLen: uint16(len(val)),
		},
	)
	encU32(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+4:], val)
	return rec
}

// newCharUintRecord creates and returns a new record that uses []byte
// slices as keys and uint32's as values.
func _(key []byte, val uint32) record {
	rsz := recordHeaderSize + len(key) + 4
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			Flags:  RK_STR | RV_NUM,
			KeyLen: uint16(len(key)),
			ValLen: 4,
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	encU32(rec[recordHeaderSize+n:], val)
	return rec
}

// newCharCharRecord creates and returns a new record that uses []byte
// slices as keys and []byte slices as values.
func _(key []byte, val []byte) record {
	rsz := recordHeaderSize + len(key) + len(val)
	rec := make(record, rsz, rsz)
	rec.encRecordHeader(
		&recordHeader{
			Flags:  RK_STR | RV_STR,
			KeyLen: uint16(len(key)),
			ValLen: uint16(len(val)),
		},
	)
	n := copy(rec[recordHeaderSize:], key)
	copy(rec[recordHeaderSize+n:], val)
	return rec
}

// encRecordHeader takes a pointer to a recordHeader and encodes it
// directly into the record as a []byte slice
func (r *record) encRecordHeader(h *recordHeader) {
	_ = (*r)[recordHeaderSize] // early bounds check
	encU32((*r)[0:4], h.Flags)
	encU16((*r)[4:6], h.KeyLen)
	encU16((*r)[6:8], h.ValLen)
}

// decRecordHeader decodes the header from the record and fills and
// returns a pointer to the recordHeader
func (r *record) decRecordHeader() *recordHeader {
	_ = (*r)[recordHeaderSize] // early bounds check
	return &recordHeader{
		Flags:  decU32((*r)[0:4]),
		KeyLen: decU16((*r)[4:6]),
		ValLen: decU16((*r)[6:8]),
	}
}

// Flags returns the underlying uint16 representing the flags set for this record.
func (r *record) Flags() uint32 {
	return decU32((*r)[0:4])
}

func (r *record) hasFlag(flag uint32) bool {
	return decU32((*r)[0:4])&flag != 0
}

// Key returns the underlying slice of bytes representing the record key
func (r *record) Key() []byte {
	return (*r)[recordHeaderSize : recordHeaderSize+decU16((*r)[2:4])]
}

// Val returns the underlying slice of bytes representing the record value
func (r *record) Val() []byte {
	return (*r)[recordHeaderSize+decU16((*r)[2:4]) : recordHeaderSize+decU16((*r)[2:4])+decU16((*r)[4:6])]
}

// KeyType returns the underlying type of the record key
func (r *record) KeyType() uint32 {
	return r.Flags() & keyMask
}

// ValType returns the underlying type of the record value
func (r *record) ValType() uint32 {
	return r.Flags() & valMask
}

// String is the stringer method for a record
func (r *record) String() string {
	fl := r.Flags()
	var k, v string
	if (fl & keyMask) == RK_NUM {
		k = fmt.Sprintf("key: %d", decU32(r.Key()))
	}
	if (fl & keyMask) == RK_STR {
		k = fmt.Sprintf("key: %q", string(r.Key()))
	}
	if (fl & valMask) == RV_NUM {
		v = fmt.Sprintf("val: %d", decU32(r.Val()))
	}
	if (fl & valMask) == RV_STR {
		v = fmt.Sprintf("val: %q", string(r.Val()))
	}
	return fmt.Sprintf("{ flags: %.4x, key: %s, val: %s }", fl, k, v)
}
