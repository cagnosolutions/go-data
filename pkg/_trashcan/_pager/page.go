package _pager

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

var offset = binary.LittleEndian

const (
	headerSize    = uint16(unsafe.Sizeof(header{}))
	pageSize      = 4096
	slotSize      = 8
	minRecordSize = 8
	maxRecordSize = pageSize - headerSize - slotSize

	slotStatusFree uint16 = 2<<iota + 1
	slotStatusUsed
)

const (
	offsPageID     = 0
	offNumSlots    = 4
	offFreeSlots   = 6
	offFreeSpaceLo = 8
	offFreeSpaceHi = 10
)

type RID struct {
	PageID uint32
	SlotID uint16
}

type slot struct {
	id     uint16
	status uint16
	offset uint16
	length uint16
}

type header struct {
	pageID      uint32
	numSlots    uint16
	freeSlots   uint16
	freeSpaceLo uint16
	freeSpaceHi uint16
}

type page []byte

func newPage(pid uint32) page {
	p := make(page, pageSize, pageSize)
	p.setPageHeader(
		header{
			pageID:      pid,
			numSlots:    0,
			freeSlots:   0,
			freeSpaceLo: headerSize,
			freeSpaceHi: pageSize,
		},
	)
	return p
}

func (p page) setPageHeader(h header) {
	offset.PutUint32(p[offsPageID:], h.pageID)
	offset.PutUint16(p[offNumSlots:], h.numSlots)
	offset.PutUint16(p[offFreeSlots:], h.freeSlots)
	offset.PutUint16(p[offFreeSpaceLo:], h.freeSpaceLo)
	offset.PutUint16(p[offFreeSpaceHi:], h.freeSpaceHi)
}

func (p page) getPageHeader() header {
	return header{
		pageID:      offset.Uint32(p[offsPageID:]),
		numSlots:    offset.Uint16(p[offNumSlots:]),
		freeSlots:   offset.Uint16(p[offFreeSlots:]),
		freeSpaceLo: offset.Uint16(p[offFreeSpaceLo:]),
		freeSpaceHi: offset.Uint16(p[offFreeSpaceHi:]),
	}
}

func (p page) getPageID() uint32 {
	return offset.Uint32(p[offsPageID:])
}

func (p page) setPageID(pid uint32) {
	offset.PutUint32(p[offsPageID:], pid)
}

func (p page) getNumSlots() uint16 {
	return offset.Uint16(p[offNumSlots:])
}

func (p page) setNumSlots(n uint16) {
	offset.PutUint16(p[offNumSlots:], n)
}

func (p page) getFreeSlots() uint16 {
	return offset.Uint16(p[offFreeSlots:])
}

func (p page) setFreeSlots(n uint16) {
	offset.PutUint16(p[offFreeSlots:], n)
}

func (p page) getFreeSpaceLo() uint16 {
	return offset.Uint16(p[offFreeSpaceLo:])
}

func (p page) setFreeSpaceLo(n uint16) {
	offset.PutUint16(p[offFreeSpaceLo:], n)
}

func (p page) getFreeSpaceHi() uint16 {
	return offset.Uint16(p[offFreeSpaceHi:])
}

func (p page) setFreeSpaceHi(n uint16) {
	offset.PutUint16(p[offFreeSpaceHi:], n)
}

func (p page) getSlotN(n uint16) *slot {
	if n > p.getFreeSpaceLo() {
		return nil
	}
	i := (n * slotSize) + headerSize
	return &slot{
		id:     offset.Uint16(p[i : i+2]),
		status: offset.Uint16(p[i+2 : i+4]),
		offset: offset.Uint16(p[i+4 : i+6]),
		length: offset.Uint16(p[i+6 : i+8]),
	}
}

func (p page) setSlotN(n uint16, s *slot) bool {
	if n > p.getFreeSpaceLo() {
		return false
	}
	i := (n * slotSize) + headerSize
	offset.PutUint16(p[i:i+2], s.id)
	offset.PutUint16(p[i+2:i+4], s.status)
	offset.PutUint16(p[i+4:i+6], s.offset)
	offset.PutUint16(p[i+6:i+8], s.length)
	return true
}

func (p page) getSlots() []*slot {
	// start at i, and go to j
	i, j := headerSize, p.getFreeSpaceLo()
	// create a new slot set
	var slots []*slot
	// loop over slots
	for i < j {
		slots = append(
			slots, &slot{
				id:     offset.Uint16(p[i : i+2]),
				status: offset.Uint16(p[i+2 : i+4]),
				offset: offset.Uint16(p[i+4 : i+6]),
				length: offset.Uint16(p[i+6 : i+8]),
			},
		)
		// increment i to get to next slot
		i += slotSize
	}
	// return slot set
	return slots
}

func (p page) locateFreeSlot() *slot {
	// Check the free slot count
	if p.getFreeSlots() < 1 {
		return nil
	}
	// Start at i, and go to j
	i, j := headerSize, p.getFreeSpaceLo()
	// Loop over slots
	for i < j {
		// Found free slot
		if offset.Uint16(p[i+2:i+4]) == slotStatusFree {
			return &slot{
				id:     offset.Uint16(p[i : i+2]),
				status: offset.Uint16(p[i+2 : i+4]),
				offset: offset.Uint16(p[i+4 : i+6]),
				length: offset.Uint16(p[i+6 : i+8]),
			}
		}
		// No match, so increment to next slot
		i += slotSize
	}
	return nil
}

func (p page) setSlots(slots []*slot) {
	var i int
	for j := range slots {
		offset.PutUint16(p[i:i+2], slots[j].id)
		offset.PutUint16(p[i+2:i+4], slots[j].status)
		offset.PutUint16(p[i+4:i+6], slots[j].offset)
		offset.PutUint16(p[i+6:i+8], slots[j].length)
		i += slotSize
	}
}

func (p page) IsDirty() bool {
	return true
}

func (p page) Data() []byte {
	return p
}

func (p page) Read(rid *RID) ([]byte, error) {

	// return
	return nil, nil
}

func (p page) Write(src []byte) (*RID, error) {

	// return
	return nil, nil
}

func (p page) Remove(rid *RID) error {
	// Ensure provided ID is valid.
	if rid == nil || rid.PageID != p.getPageID() {
		return ErrInvalidRecordID
	}
	// get the slot for the record.
	s := p.getSlotN(rid.SlotID)
	// Update the slot to reflect the record removal.
	s.status = slotStatusFree
	// Determine if the record is at the high-water
	// mark, because if it is, then we do not need
	// to perform a compaction.
	if s.offset != p.getFreeSpaceHi() {
		// We do not need to compact, so let us simply
		// remove the slot and return.
		p.removeSlot(s.id)
		return nil
	}
	// Otherwise, we must perform a compaction.
	p.compact()
	// And then, remove the slot.
	p.removeSlot(s.id)
	// Finally, we can return.
	return nil
}

func (p page) removeSlot(sid uint16) {
	// make this work.
}

// compact iterates through the records within the page using the slot
// directory. If a slot with a status of slotStatusFree is found, the
// address of the associated record is located and the space that the
// record occupied is compacted.
func (p page) compact() {
	var beg, end int
	_ = beg
	_ = end
	// Loop over the slots
	for _, s := range p.getSlots() {
		// Compact contiguous slots
		s.status = slotStatusFree
	}
}

func (p page) PageHeaderString() string {
	ss := fmt.Sprintf("pageID=%d\n", p.getPageID())
	ss += fmt.Sprintf("numSlots=%d\n", p.getNumSlots())
	ss += fmt.Sprintf("freeSlots=%d\n", p.getFreeSlots())
	ss += fmt.Sprintf("freeSpaceLo=%d\n", p.getFreeSpaceLo())
	ss += fmt.Sprintf("freeSpaceHi=%d\n", p.getFreeSpaceHi())
	return ss
}
