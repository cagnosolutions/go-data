package pager

import (
	"encoding/binary"
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
)

// slotID is a slot index
type slotID uint16

// slot represents a record slot index
type slot struct {
	id     slotID
	status uint16
	offset uint16
	length uint16
}

// bounds is just a helper method that returns the record offset
// beginning and end offsets.
func (s slot) bounds() (uint16, uint16) {
	return s.offset, s.offset + s.length
}

// recordID is an ID for a record and is composed of a pageID and
// a slotID. The slotID is nothing more than the index of the slot.
type recordID struct {
	pid pageID
	sid slotID
}

// header is the page's header
type header struct {
	pid   pageID // page id
	magic uint16 // status and type (for now, but can include others)
	slots uint16 // number of slots in page
	lower uint16 // free space lower offset
	upper uint16 // free space upper offset
}

// page is a basic representation of a slotted page
type page struct {
	header
	data []byte
}

// newPage instantiates and returns a new pointer to a page.
func newPage(pid pageID) *page {
	return &page{
		header: header{
			pid:   pid,
			magic: statUsed,
			slots: 0,
			lower: hdrSize,
			upper: pageSize,
		},
		data: make([]byte, pageSize),
	}
}

// getPageID returns this page's pageID
func (p *page) getPageID() pageID {
	return p.pid
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
	if recSize < p.upper-p.lower {
		return ErrPossiblePageFull
	}
	return nil
}

// addSlot adds a new slot entry to the page. A page slot entry
// has two values, the offset of the record to be inserted, and
// the record length. Both the offset, and length are encoded as
// uint16 types. **addSlot is called by getSlot.
func (p *page) addSlot(recSize uint16) *slot {
	// **Note, we are assuming that record size checks have been
	// done before calling add slot.
	//
	// First we increment the slot count.
	p.slots++
	// Next, we raise the free space lower bound, because we are
	// adding a new slot.
	p.lower += slotSize
	// Next, we lower the free space upper bound, because we are
	// going to be adding record data.
	p.upper -= recSize
	// Now, we get the slot entry offset, for ease of use.
	i := p.lower - slotSize
	// Next, we add the slot. First encoding the slot status, then
	// the record offset, and then the record length.
	binary.LittleEndian.PutUint16(p.data[i:i+2], statUsed)
	i += 2
	binary.LittleEndian.PutUint16(p.data[i:i+2], p.upper)
	i += 2
	binary.LittleEndian.PutUint16(p.data[i:i+2], recSize)
	// Lastly, we will return the slot that we just inserted as
	// represented by the slot data structure.
	return &slot{
		id:     slotID(p.slots - 1),
		status: statUsed,
		offset: p.upper,
		length: recSize,
	}
}

// delSlot deletes a slot entry on the page. It doesn't actually
// remove the slot, but it marks the slot status field as free.
// This way, we can search for, and use "deleted" slots when we
// want to enter new data (if the record fits in the free slot)
// and if not, we can just add another one, and use the free slot
// another time.
func (p *page) delSlot(sid slotID) error {
	// First, we get the slot entry offset, for ease of use.
	i := hdrSize + (uint16(sid) * slotSize)
	// Next, do some error checking to make sure the slot id
	// provided is not outside the slot bounds.
	if i > p.lower {
		return ErrSlotIDOutOfBounds
	}
	// Then we can overwrite the slot status field, and mark
	// it as a free slot.
	binary.LittleEndian.PutUint16(p.data[i:i+2], statFree)
	// Finally, we return.
	return nil
}

// getSlot returns a slot id we can use to write a new record. It
// will first search through any slots that exist and have been
// marked as free, and it will see if the record size will fit within
// any potential free slots and return one of those id's. If a free
// slot cannot be found, it will call addSlot and simply allocate a
// new one.
func (p *page) getSlot(recSize uint16) *slot {
	// **Note, we are assuming that record size checks have been
	// done before calling add slot.
	//
	// First, we iterate through the slots and check for any
	// slot statuses that are set to statFree.
	var status, size uint16
	for sid := uint16(0); sid < p.slots; sid++ {
		// Get the slot offset, and check each slot status...
		i := hdrSize + (sid * slotSize)
		status = binary.LittleEndian.Uint16(p.data[i : i+2])
		if status == statFree {
			// We found a free slot, check the record length to make
			// sure that it will work as a fit.
			size = binary.LittleEndian.Uint16(p.data[i+4 : i+6])
			if recSize <= size {
				// We have located a free slot that can fit the record.
				// We should update the slot status, and length for the
				// new record.
				binary.LittleEndian.PutUint16(p.data[i:i+2], statUsed)
				binary.LittleEndian.PutUint16(p.data[i+4:i+6], recSize)
				// Then, we can simply return this slot.
				return &slot{
					id:     slotID(sid),
					status: statUsed,
					offset: binary.LittleEndian.Uint16(p.data[i+2 : i+4]),
					length: recSize,
				}
			}
			// If we get here, we found a free slot but the record will
			// not fit here, so we continue on to the next slot.
		}
		// If we get here, this slot is not marked as free.
	}
	// If we get here, then we have not found any free slots, so we must
	// simply allocate and return a new one.
	return p.addSlot(recSize)
}

// findSlot attempts to locate and return the slot matching the
// slotID provided.
func (p *page) findSlot(sid slotID) *slot {
	// Get the slot offset.
	i := hdrSize + (uint16(sid) * slotSize)
	// Decode and return the slot.
	return &slot{
		id:     sid,
		status: binary.LittleEndian.Uint16(p.data[i : i+2]),
		offset: binary.LittleEndian.Uint16(p.data[i+2 : i+4]),
		length: binary.LittleEndian.Uint16(p.data[i+4 : i+6]),
	}
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
	// First, we iterate through the slots and check for any slot stats
	// that say they are set to statFree.
	var status, size uint16
	for sid := uint16(0); sid < p.slots; sid++ {
		// Get the slot offset, and check each slot status...
		i := hdrSize + (sid * slotSize)
		status = binary.LittleEndian.Uint16(p.data[i : i+2])
		if status == statFree {
			// We found a free slot, add the record size to the total
			// fragmented free space and continue.
			size += binary.LittleEndian.Uint16(p.data[i+4 : i+6])
		}
		// If we get here, this slot is not marked as free.
	}
	// Lastly, return our contiguous free space, and our fragmented
	// free space values.
	return p.upper - p.lower, size
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
	sl := p.getSlot(recSize)
	// And then we get the record beginning and ending offsets, for an
	// easier time writing the actual record data to the page.
	beg, end := sl.bounds()
	// Then, we copy the record to the page.
	copy(p.data[beg:end], rec)
	// Lastly, we will create a recordID to be returned.
	rid := &recordID{
		pid: p.pid,
		sid: sl.id,
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
	if rid.pid != p.pid || uint16(rid.sid) > p.slots {
		return nil, ErrBadRID
	}
	// Next, we will locate the appropriate slot.
	sl := p.findSlot(rid.sid)
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
	if rid.pid != p.pid || uint16(rid.sid) > p.slots {
		return ErrBadRID
	}
	// Next, we will locate the appropriate slot.
	sl := p.findSlot(rid.sid)
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
	err := p.delSlot(sl.id)
	if err != nil {
		return err
	}
	return nil
}
