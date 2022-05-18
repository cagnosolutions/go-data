package pager

import (
	"encoding/hex"
	"fmt"
)

// flags for setting values on the page headers magic bits
const (
	// **Remember, a nibble can only hold 16 digits (0-15), so do not have
	// any single nibble (half-byte) go higher than 8 or alternatively, do
	// not have any combined nibbles add up to anything more than 15, or
	// they will carry over!

	statFree = 0x0001
	statUsed = 0x0002
	statSpec = 0x0004

	typeA = 0x0010
	typeB = 0x0020
	typeC = 0x0040
	typeD = 0x0080

	// field offsets
	offPID   = 0
	offMagic = 4
	offSlots = 6
	offLower = 8
	offUpper = 10
)

// slotID is a slot index
type slotID uint16

// slot represents a record slot index
type slot struct {
	status uint16
	offset uint16
	length uint16
}

func (s slot) encode(b []byte) {
	_ = b[5] // bounds check hint to compiler
	b[0] = byte(s.status)
	b[1] = byte(s.status >> 8)
	b[2] = byte(s.offset)
	b[3] = byte(s.offset >> 8)
	b[4] = byte(s.length)
	b[5] = byte(s.length >> 8)
}

func (s slot) decode(b []byte) {
	_ = b[6] // bounds check hint to compiler
	s.status = uint16(b[0]) | uint16(b[1])<<8
	s.offset = uint16(b[2]) | uint16(b[3])<<8
	s.length = uint16(b[4]) | uint16(b[5])<<8
}

// bounds is just a helper method that returns the record offset
// beginning and end offsets.
func (s slot) bounds() (uint16, uint16) {
	return s.offset, s.offset + s.length
}

func (s slot) String() string {
	return fmt.Sprintf("off=%.4d, len=%.4d, free=%v", s.offset, s.length, s.status == statFree)
}

// recordID is an ID for a record and is composed of a pageID and
// a slotID. The slotID is nothing more than the index of the slot.
type recordID struct {
	pid pageID
	sid slotID
}

// header is the page's header
type header struct {
	pid   uint32 // page id
	magic uint16 // status and type (for now, but can include others)
	slots uint16 // number of slots in page
	lower uint16 // free space lower offset
	upper uint16 // free space upper offset
}

// page is a basic representation of a slotted page
type page struct {
	header
	sls  []*slot
	data []byte
}

// newPage instantiates and returns a new pointer to a page.
func newPage(pid pageID) *page {
	return &page{
		header: header{
			pid:   uint32(pid),
			magic: statUsed,
			slots: 0,
			lower: hdrSize,
			upper: pageSize,
		},
		sls:  make([]*slot, 0),
		data: make([]byte, pageSize),
	}
}

// getPageID returns this page's pageID
func (p *page) getPageID() pageID {
	return pageID(p.header.pid)
}

// isFree returns a boolean value indicating true if the page
// is currently free to use.
func (p *page) isFree() bool {
	return p.magic&statFree > 0
}

// checkRecord checks to see fi there is room for the record but,
// it also checks if the size of the record is less than the
// minimum or greater than the maximum allowed record size.
func (p *page) checkRecord(recSize uint16) error {
	if recSize < minRecSize {
		return ErrMinRecSize
	}
	if recSize > maxRecSize {
		return ErrMaxRecSize
	}
	if recSize > p.upper-p.lower {
		return ErrPossiblePageFull
	}
	return nil
}

// addSlot adds a new slot entry to the page.
func (p *page) addSlot(recSize uint16) (slotID, *slot) {
	// To add a new page slot we must increment the slot count,
	// raise the lower bound, and lower the upper bound.
	p.header.slots++
	p.header.lower += slotSize
	p.header.upper -= recSize
	// Then we can create a new slot, and add it to the slot list.
	sl := &slot{
		status: statUsed,
		offset: p.header.upper,
		length: recSize,
	}
	p.sls = append(p.sls, sl)
	// Encode slot
	// sl.encode(p.data[p.header.slots*slotSize:])
	// Finally, we return the new slot pointer.
	return slotID(p.header.slots - 1), sl
}

// delSlot deletes a slot entry on the page. It doesn't actually
// remove the slot, but it marks the slot status field as free.
// This way, we can search for, and use "deleted" slots when we
// want to enter new data (if the record fits in the free slot)
// and if not, we can just add another one, and use the free slot
// another time.
func (p *page) delSlot(sid slotID) error {
	// Locate the slot using the slot id provided.
	sl, err := p.findSlot(sid)
	if err != nil {
		return err
	}
	// Then, update the slot's status.
	sl.status = statFree
	// Finally, we return.
	return nil
}

// getSlot returns a slot id we can use to write a new record. It
// will first search through any slots that exist and have been
// marked as free, and it will see if the record size will fit within
// any potential free slots and return one of those id's. If a free
// slot cannot be found, it will call addSlot and simply allocate a
// new one.
func (p *page) getSlot(recSize uint16) (slotID, *slot) {
	// First, search for a free slot.
	for sid := range p.sls {
		if p.sls[sid].status == statFree {
			// Found a free slot...
			if recSize <= p.sls[sid].length {
				// ... and the record fits; update and return it.
				p.sls[sid].status = statUsed
				return slotID(sid), p.sls[sid]
			}
		}
	}
	// We did not find a free slot, so we will add a new one.
	return p.addSlot(recSize)
}

// findSlot attempts to locate and return the slot matching the
// slotID provided.
func (p *page) findSlot(sid slotID) (*slot, error) {
	// First, check for any errors.
	if int(sid) > len(p.sls) {
		return nil, ErrSlotIDOutOfBounds
	}
	// Finally, we return the isolated slot from the slot list,
	// and we return it.
	return p.sls[sid], nil
}

// getFreeSpace calculates and returns the total free space left in
// the page. The first number that is returned is the contiguous free
// space left in the page and the seconds number that is returned is
// the total fragmented free space that is not contiguous that is left
// in the page. This may be useful for determining if the page needs
// to be compacted, or if there is even enough free space left to fit
// any more records. **Note, these values may need to be set in the
// header long term if this method ends up consuming too many resources.
func (p *page) getFreeSpace() (uint16, uint16) {
	// Calculate the total fragmented free space from any free slots.
	var frag uint16
	for sid := range p.sls {
		if p.sls[sid].status == statFree {
			// Found a free slot...
			frag += p.sls[sid].length
		}
	}
	// Finally, return the contiguous free space and the fragmented
	// free space.
	return p.upper - p.lower, frag
}

// addRecord adds a new record to the page if it has room. The page will
// attempt to reuse any available slots, and will return any errors it
// encounters, if any, along the way. One of the potential errors that
// may be returned is an ErrPossiblePageFull error which usually indicates
// that the page may be full, but that there may be possible fragmented
// space available if the page were to be compacted. Page compaction is
// a separate method that can be called externally and optionally if the
// need arises.
func (p *page) addRecord(rec []byte) (*recordID, error) {
	// First, we get the size of the record to be added.
	recSize := uint16(len(rec))
	// Next, we do a record sanity check.
	err := p.checkRecord(recSize)
	if err != nil {
		return nil, err
	}
	// Next, we get a slot. At this point a new slot has been allocated
	// and written to the page, or an existing one has been used, in
	// which case it will have been updated on the page.
	sid, sl := p.getSlot(recSize)
	// And then we get the record beginning and ending offsets, for an
	// easier time writing the actual record data to the page.
	beg, end := sl.bounds()
	// Then, we copy the record to the page.
	copy(p.data[beg:end], rec)
	// Lastly, we will create a recordID to be returned.
	rid := &recordID{
		pid: pageID(p.header.pid),
		sid: sid,
	}
	// And finally, we will return the recordID, and a nil error.
	return rid, nil
}

// getRecord searches for a record matching the provided recordID in
// the page and returns it if it can be located, otherwise and error
// will be returned.
func (p *page) getRecord(rid *recordID) ([]byte, error) {
	// First, we check to make sure the recordID provided is actually
	// a valid recordID, and if not we will return an error.
	if rid.pid != p.getPageID() {
		return nil, ErrBadRID
	}
	// Next, we will locate the appropriate slot.
	sl, err := p.findSlot(rid.sid)
	if err != nil {
		return nil, err
	}
	// Ensure that the slot has not been marked as free.
	if sl.status == statFree {
		// Looks like it is a free slot, which means it has been removed.
		return nil, ErrRecNotFound
	}
	// Otherwise, we have located the record. We will create a temporary
	// buffer to copy the record data into (so that we are not returning
	// a slice of the underlying page data, which would be unsafe) and
	// then we can copy and return the record.
	buf := make([]byte, sl.length)
	// Now we get the record bounds and copy into the buffer.
	beg, end := sl.bounds()
	copy(buf, p.data[beg:end])
	// Finally, we return the record data (which is completely safe) to
	// use, mutate or otherwise do whatever we wish to because it is a
	// fresh copy.
	return buf, nil
}

// delRecord removes a record (by overwriting the record with zeroed out
// data) from the page. It will update the slot, marking it as free and
// return any errors encountered along the way.
func (p *page) delRecord(rid *recordID) error {
	// First, we check to make sure the recordID provided is actually
	// a valid recordID, and if not we will return an error.
	if rid.pid != p.getPageID() {
		return ErrBadRID
	}
	// Next, we will locate the appropriate slot.
	sl, err := p.findSlot(rid.sid)
	if err != nil {
		return err
	}
	// If the slot is already marked as free, we can simply return.
	if sl.status == statFree {
		// Looks like it is a free slot, nothing to do here.
		return nil
	}
	// Otherwise, we have located the record. We will create a zeroed
	// out record buffer to overwrite the existing record with, and
	// then we can copy it over the record.
	buf := make([]byte, sl.length)
	// Now we get the record bounds.
	beg, end := sl.bounds()
	// And copy zeroed out data over the exising record on the page.
	copy(p.data[beg:end], buf)
	// Now, we must update the slot, and then we can return.
	err = p.delSlot(rid.sid)
	if err != nil {
		return err
	}
	return nil
}

func (p *page) String() string {
	ss := fmt.Sprintf("          +------------------[ header ]------------------+\n")
	ss += fmt.Sprintf(
		"          |\tpid=%.2d, slots=%.2d, lo=%.2d, hi=%.2d\n",
		p.header.pid, p.header.slots, p.header.lower, p.header.upper,
	)
	ss += fmt.Sprintf("          +------------------[ pslcnt ]------------------+\n")
	for _, sl := range p.sls {
		ss += fmt.Sprintf("          |\t%s\n", sl)
	}
	ss += fmt.Sprintf("          +------------------[  data  ]------------------+\n")
	ss += fmt.Sprintf(hex.Dump(p.data))
	ss += fmt.Sprintf("          +----------------------------------------------+\n")
	return ss
}
