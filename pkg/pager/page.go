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

// addSlot adds a new slot entry to the page. A page slot entry
// has two values, the offset of the record to be inserted, and
// the record length. Both the offset, and length are encoded as
// uint16 types.
func (p *page) addSlot(recSize uint16) uint16 {
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
	// Lastly, we will return the slot index that we just inserted
	// which will always just be p.slots - 1, since we incremented
	// it at the very start of this method.
	return p.slots - 1
}

// delSlot deletes a slot entry on the page. It doesn't actually
// remove the slot, but it marks the slot status field as free.
// This way, we can search for, and use "deleted" slots when we
// want to enter new data (if the record fits in the free slot)
// and if not, we can just add another one, and use the free slot
// another time.
func (p *page) delSlot(sid uint16) error {
	// First, we get the slot entry offset, for ease of use.
	i := hdrSize + (sid * slotSize)
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
func (p *page) getSlot(recSize uint16) uint16 {
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
				// We have located a free slot that can fit the record,
				// so we simply return this slot id.
				return sid
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
