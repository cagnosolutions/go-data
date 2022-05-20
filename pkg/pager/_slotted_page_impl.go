package pager

import (
	"encoding/binary"
	"errors"
)

var (
	ErrRecordTooSmall = errors.New("record is too small (under min size allowed)")
	ErrRecordTooBig   = errors.New("record is too big (over max size allowed)")
	ErrNoRoom         = errors.New("the page is full, or has too much fragmentation")
	ErrInvalidPID     = errors.New("page ID is not valid, or does not match")
	ErrInvalidSID     = errors.New("slot ID is not valid (out of bounds perhaps?)")
	ErrRecordNotFound = errors.New("record not found, it may have been deleted")
)

const (
	stFree = 0x0001
	stUsed = 0x0002

	minSize = 8
	maxSize = 2048

	szHd = 12
	szSl = 6
	szPg = 4096

	offPID   = 0
	offMagic = 4
	offSlots = 6
	offLower = 8
	offUpper = 10
)

var bin = binary.LittleEndian

// slot is a struct which is an index for a record
type slot struct {
	status uint16
	offset uint16
	length uint16
}

// bounds returns the starting and ending offsets to the
// particular record that this slot index "points" to
func (s slot) bounds() (uint16, uint16) {
	return s.offset, s.offset + s.length
}

// rid is a struct representing a record ID
type rid struct {
	pid uint32 // page id
	sid uint16 // slot id (slot index)
}

// header is a struct representing a page header
type header struct {
	pid   uint32 // id of page
	magic uint16 // status and type
	slots uint16 // number of slots
	lower uint16 // lower free space bound
	upper uint16 // upper free space bound
}

// page is a page
type page []byte

// newPage returns a new page instance set with the provided page ID
func newPage(pid uint32) page {
	pg := make(page, szPg, szPg)
	pg.setHeader(
		&header{
			pid:   pid,
			magic: stUsed,
			slots: 0,
			lower: szHd,
			upper: szPg,
		},
	)
	return pg
}

// setHeader encodes the provided header structure to the underlying
// page
func (p *page) setHeader(h *header) {
	bin.PutUint32((*p)[offPID:], h.pid)
	bin.PutUint16((*p)[offMagic:], h.magic)
	bin.PutUint16((*p)[offSlots:], h.slots)
	bin.PutUint16((*p)[offLower:], h.lower)
	bin.PutUint16((*p)[offUpper:], h.upper)
}

// getHeader decodes (from the underlying page) and returns a
// pointer to a header structure
func (p *page) getHeader() *header {
	return &header{
		pid:   bin.Uint32((*p)[offPID:]),
		magic: bin.Uint16((*p)[offMagic:]),
		slots: bin.Uint16((*p)[offSlots:]),
		lower: bin.Uint16((*p)[offLower:]),
		upper: bin.Uint16((*p)[offUpper:]),
	}
}

// freeSpace returns the amount of contiguous free space left in the
// page (space between the upper and lower bounds)
func (p *page) freeSpace() uint16 {
	return bin.Uint16((*p)[offUpper:]) - bin.Uint16((*p)[offLower:])
}

// checkRecord performs sanity and error checking on a record size
func (p *page) checkRecord(size uint16) error {
	if size < minSize {
		return ErrRecordTooSmall
	}
	if size > maxSize {
		return ErrRecordTooBig
	}
	if size > p.freeSpace() {
		return ErrNoRoom
	}
	return nil
}

// setSlot writes the provided slot to the location derived using
// the supplied slot index ID. setSlot panics if the provided sid
// is out of bounds.
func (p *page) setSlot(sl *slot, sid uint16) {
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

// getSlot reads the slot at the provided slot location derived using
// the supplied slot index ID. getSlot panics if the provided sid
// is out of bounds.
func (p *page) getSlot(sid uint16) *slot {
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
	return &slot{
		status: bin.Uint16((*p)[offStatus:]),
		offset: bin.Uint16((*p)[offOffset:]),
		length: bin.Uint16((*p)[offLength:]),
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
		status: stUsed,
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

// getSlot adds or locates a free slot that will fit the record size
// provided. It returns a slot along with the slot index.
func (p *page) locateSlot(size uint16) (uint16, *slot) {
	// try to find a free slot we can use
	slotCount := bin.Uint16((*p)[offSlots:])
	for sid := uint16(0); sid < slotCount; sid++ {
		sl := p.getSlot(sid)
		if sl.status == stFree && size <= sl.length {
			// we can use this slot, but first we update and save it
			sl.status = stUsed
			sl.length = size
			p.setSlot(sl, sid)
			// and then we return it for use
			return sid, sl
		}
	}
	// otherwise, we append and return a fresh slot
	return p.addSlot(size)
}

// addRecord writes a new record to the page. It returns a *rid which
// is a record ID, along with any potential errors encountered.
func (p *page) addRecord(data []byte) (*rid, error) {
	// get the record size
	rsize := uint16(len(data))
	// sanity check the record
	err := p.checkRecord(rsize)
	if err != nil {
		return nil, err
	}
	// get a new (or used) slot to index the record
	sid, sl := p.locateSlot(rsize)
	// get the record bounds from the slot index
	beg, end := sl.bounds()
	// write the record to the page (using record bounds)
	copy((*p)[beg:end], data)
	// assemble and return the record ID
	return &rid{
		pid: bin.Uint32((*p)[offPID:]),
		sid: sid,
	}, nil

}

// checkRID performs error and sanity checking on the provided
// record ID.
func (p *page) checkRID(id *rid) error {
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
func (p *page) getRecord(id *rid) ([]byte, error) {
	// sanity check the record ID
	err := p.checkRID(id)
	if err != nil {
		return nil, err
	}
	// find the associated slot index (ensure it is a used slot)
	sl := p.getSlot(id.sid)
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

// delSlot updates the status of the slot found at the provided
// slot ID, and returns the slot for use in the delete operation.
func (p *page) delSlot(sid uint16) *slot {
	// get the slot using the sid
	sl := p.getSlot(sid)
	// if the slot status is free, return nil
	if sl.status == stFree {
		return nil
	}
	// save the status of the found slot
	p.setSlot(sl, sid)
	// and return
	return sl
}

// delRecord removes a record from the page. It overwrites the record
// data with zeros and marks the slot index as a "free to use" slot
// so that it can be re-used at a later date if there is another
// record that can occupy the same (or less) space. It returns any
// errors encountered.
func (p *page) delRecord(id *rid) error {
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
