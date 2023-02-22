package _trash

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

// page is a basic representation of a slotted page
type page struct {
	data []byte
}

// newPage instantiates and returns a new pointer to a page.
func newPage(pid pageID) *page {
	data := make([]byte, pageSize)
	BinMapU32(data[offPID:], SetU32(pid))
	BinMapU16(data[offMagic:], SetU16(statUsed))
	BinMapU16(data[offSlots:], SetU16(0))
	BinMapU16(data[offLower:], SetU16(hdrSize))
	BinMapU16(data[offUpper:], SetU16(pageSize))
	return &page{
		data: data,
	}
}

// getPageID returns this page's pageID
func (p *page) getPageID() pageID {
	var pid uint32
	BinMapU32(p.data[offPid:], GetU32(&pid))
	return pageID(pid)
}

// isFree returns a boolean value indicating true if the page
// is currently free to use.
func (p *page) isFree() bool {
	var magic uint16
	BinMapU16(p.data[offPid:], GetU16(&magic))
	return magic&statFree > 0
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
	if recSize > diffU16(p.data, offUpper, offLower) {
		return ErrPossiblePageFull
	}
	return nil
}

func (p *page) AddSlot(size uint16) *slot {
	// increment the slot count by one
	BinMapU16(p.data[offSlots:], IncrU16(1))
	// raise the free space lower bound (by a slot size)
	BinMapU16(p.data[offLower:], IncrU16(slotSize))
	// lower the free space upper bound (by the record size)
	BinMapU16(p.data[offUpper:], DecrU16(recSize))
	// get the upper and lower bounds and create a slot
	var lower, upper uint16
	BinMapU16(p.data[offLower:], GetU16(&lower))
	BinMapU16(p.data[offUpper:], GetU16(&upper))
	sl := &slot{
		status: statUsed,
		offset: upper,
		length: size,
	}
	// next we get the slot (to write) offset and
	// encode the new slot.
	offset := lower - slotSize
	BinMapU16(p.data[offset:], SetU16(statUsed))
	BinMapU16(p.data[offset+2:], SetU16(upper))
	BinMapU16(p.data[offset+4:], SetU16(size))
	// finally, return the slot
	return sl
}

// addSlot adds a new slot entry to the page. A page slot entry
// has two values, the offset of the record to be inserted, and
// the record length. Both the offset, and length are encoded as
// uint16 types. **addSlot is called by getSlot.
func (p *page) addSlot(recSize uint16) *slot {
	// **Note, we are assuming that record size checks have been
	// done before calling add slot.
	//
	// First we get the slot count for the sid when we return.
	sid := getU16(p.data, offSlots)
	// Next we increment the slot count.
	incrU16(p.data, offSlots, 1)
	// Next, we raise the free space lower bound, because we are
	// adding a new slot.
	incrU16(p.data, offLower, slotSize)
	// Next, we lower the free space upper bound, because we are
	// going to be adding record data.
	decrU16(p.data, offUpper, recSize)
	// Now, we get the slot entry offset, for ease of use.
	offSlot := int(diffU16(p.data, offLower, slotSize))
	// We want to get the upper bound to set in the slot and for
	// later when we return this slot.
	upper := getU16(p.data, offUpper)
	// Next, we add the slot. First encoding the slot status, then
	// the record offset, and then the record length.
	putU16(p.data, offSlot, statUsed)
	offSlot += 2
	putU16(p.data, offSlot, upper)
	offSlot += 2
	putU16(p.data, offSlot, recSize)
	// Lastly, we will return the slot that we just inserted as
	// represented by the slot data structure.
	return &slot{
		id:     slotID(sid),
		status: statUsed,
		offset: upper,
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
	off := int(hdrSize + (uint16(sid) * slotSize))
	// Next, do some error checking to make sure the slot id
	// provided is not outside the slot bounds.
	if off > int(getU16(p.data, offLower)) {
		return ErrSlotIDOutOfBounds
	}
	// Then we can overwrite the slot status field, and mark
	// it as a free slot.
	putU16(p.data, off, statFree)
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
	var status, size, slotCount uint16
	// get the slot count
	slotCount = getU16(p.data, offSlots)
	for sid := uint16(0); sid < slotCount; sid++ {
		// get the slot offset, and check each slot status...
		off := int(hdrSize + (sid * slotSize))
		status = getU16(p.data, off)
		if status == statFree {
			// We found a free slot, check the record length to make
			// sure that it will work as a fit.
			size = getU16(p.data, off+4)
			if recSize <= size {
				// We have located a free slot that can fit the record.
				// We should update the slot status, and length for the
				// new record.
				putU16(p.data, off, statUsed)
				putU16(p.data, off+4, recSize)
				// Then, we can simply return this slot.
				return &slot{
					id:     slotID(sid),
					status: statUsed,
					offset: getU16(p.data, off+2),
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
	// get the slot offset.
	off := int(hdrSize + (uint16(sid) * slotSize))
	// Decode and return the slot.
	return &slot{
		id:     sid,
		status: getU16(p.data, off),
		offset: getU16(p.data, off+2),
		length: getU16(p.data, off+4),
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
	var status, size, slotCount uint16
	// get the slot count
	slotCount = getU16(p.data, offSlots)
	for sid := uint16(0); sid < slotCount; sid++ {
		// get the slot offset, and check each slot status...
		off := int(hdrSize + (sid * slotSize))
		status = getU16(p.data, off)
		if status == statFree {
			// We found a free slot, add the record size to the total
			// fragmented free space and continue.
			size += getU16(p.data, off+4)
		}
		// If we get here, this slot is not marked as free.
	}
	// Lastly, return our contiguous free space, and our fragmented
	// free space values.
	return diffU16(p.data, offUpper, offLower), size
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
		pid: p.getPageID(),
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
	if rid.pid != p.getPageID() || uint16(rid.sid) > getU16(p.data, offSlots) {
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
	if rid.pid != p.getPageID() || uint16(rid.sid) > getU16(p.data, offSlots) {
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

func (p *page) String() string {
	pid := getU32(p.data, offPID)
	// magic := getU16(p.data, offMagic)
	slots := getU16(p.data, offSlots)
	lower := getU16(p.data, offLower)
	upper := getU16(p.data, offUpper)
	ss := fmt.Sprintf("+--------------[ header ]--------------+\n")
	ss += fmt.Sprintf("|\tpid=%.2d, slots=%.2d, lo=%.2d, hi=%.2d\n", pid, slots, lower, upper)
	// ss += fmt.Sprintf("|   ID=0x%.4x (%d)\n", pid, pid)
	// ss += fmt.Sprintf("| Magic=0x%.2x (%d)\n", magic, magic)
	// ss += fmt.Sprintf("| Slots=0x%.2x (%d)\n", slots, slots)
	// ss += fmt.Sprintf("| Lower=0x%.2x (%d)\n", lower, lower)
	// ss += fmt.Sprintf("| Upper=0x%.2x (%d)\n", upper, upper)
	ss += fmt.Sprintf("+--------------[ pslcnt ]--------------+\n")
	for i := 0; i < int(slots); i++ {
		sl := p.findSlot(slotID(i))
		ss += fmt.Sprintf("|\t%s\n", sl)
	}
	ss += fmt.Sprintf("          +--------------------[  data  ]--------------------+\n")
	ss += fmt.Sprintf(hex.Dump(p.data))
	ss += fmt.Sprintf("+--------------------------------------+\n")
	return ss
}
