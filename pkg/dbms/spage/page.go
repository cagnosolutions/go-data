package spage

import (
	"encoding/json"
)

// Page is a pageSized data Page
// structure that may contain
// one or more data records
type Page struct {
	header *pageHeader
	slots  []*pageSlot
	data   []byte
}

// NewPage is a new Page constructor
// that creates and returns a new *Page
func NewPage(pid uint32) *Page {
	return &Page{
		header: &pageHeader{
			pageID:         pid,
			nextPageID:     0,
			prevPageID:     0,
			freeSpaceLower: pageHeaderSize,
			freeSpaceUpper: pageSize,
			slotCount:      0,
			freeSlotCount:  0,
			hasOverflow:    0,
			reserved:       0,
		},
		slots: make([]*pageSlot, 0),
		data:  make([]byte, pageSize),
	}
}

// LinkPages links Page "a" with Page "b"; they are
// marked as overflow pages and have their nextPageID
// and prevPageID linked to each other. The next and
// prev pageID's can be used to traverse linked pages
// in the same fashion that a linked list allows you
// to traverse nodes.
func LinkPages(a, b *Page) *Page {
	a.header.nextPageID = b.header.pageID
	a.header.hasOverflow = 1
	b.header.prevPageID = a.header.pageID
	b.header.hasOverflow = 1
	return a
}

// Link links the calling Page to the next Page and is
// provided as an alternate method to LinkPages. All the
// same specs apply.
func (p *Page) Link(next *Page) *Page {
	p.header.nextPageID = next.header.pageID
	p.header.hasOverflow = 1
	next.header.prevPageID = p.header.pageID
	next.header.hasOverflow = 1
	return p
}

// PageID returns the current pageID
func (p *Page) PageID() uint32 {
	return p.header.pageID
}

// PrevID returns the current pageID
func (p *Page) PrevID() uint32 {
	return p.header.prevPageID
}

// NextID returns the current pageID
func (p *Page) NextID() uint32 {
	return p.header.nextPageID
}

// CheckRecord checks if there is room for the record
// but, it also checks if the recordSize is outside
// the bounds of the minimum or maximum record size
// and returns an applicable error if so
func (p *Page) CheckRecord(recordSize uint16) error {
	if recordSize < MinRecordSize {
		return ErrMinRecordSize
	}
	if recordSize > MaxRecordSize {
		return ErrMaxRecordSize
	}
	if !p.hasRoom(recordSize) {
		return ErrNoMoreRoomInPage
	}
	return nil
}

// hasRoom does a simple check to see if there is enough
// room left in the Page to accommodate a recordSized size
// data record
func (p *Page) hasRoom(recordSize uint16) bool {
	return recordSize < p.header.FreeSpace()
}

// getAvailableSlot returns a free Page slot if there is
// one already existing that can be used, otherwise it
// adds a new pageSlot. It returns a *pageSlot to use
// for inserting a new record.
func (p *Page) getAvailableSlot(recordSize uint16) *pageSlot {
	// first check the Page header to see if
	// the freeSlotCount is reporting any
	if p.header.freeSlotCount > 0 {
		// it looks like we might have one, so
		// scan the existing slot set and check
		// for any that are currently marked free
		for i := range p.slots {
			if p.slots[i].itemStatus == itemStatusFree {
				// looks like we found one, lets use it!
				return p.useFreePageSlotRecord(p.slots[i], recordSize)
			}
		}
		// we should NEVER get here
		panic("if you see this, look in Page.go around line 125")
	}
	// otherwise, looks like we aren't reporting
	// that we have any existing pageSlots free,
	// so we should just add a new Page slot record
	// return that, and be done
	return p.addNewPageSlotRecord(recordSize)
}

// useFreePageSlotRecord uses an existing Page slot record provided. it
// attempts to use the same record offset (if it will fit) otherwise, it
// will find another location in the Page and update the header accordingly
func (p *Page) useFreePageSlotRecord(
	slot *pageSlot, recordSize uint16,
) *pageSlot {
	// no need to increment the slotCount however
	// we do need to decrement the freeSlotCount
	p.header.freeSlotCount--
	// let's check to see if the last record offset
	// had enough space to fit this record in
	if recordSize <= slot.itemLength {
		// it will fit, nice! let us update the slot
		// to fit the new record information--remember
		// the only things that will change are the
		// itemStatus, and the itemLength
		slot.itemStatus = itemStatusUsed
		slot.itemLength = recordSize
		// all done--it fit nicely, so we just return
		return slot
	}
	// now, if we are here it means that the last
	// record was not large enough to house this
	// record data, so first we just need to allocate
	// a whole new offset to store the record that
	// is the size of the new record
	p.header.freeSpaceUpper -= recordSize
	// we do not need to raise the free space
	// lower bound because we are not adding a
	// new slot--but now we need to update the
	// slot itemStatus, itemOffset, and itemLength...
	slot.itemStatus = itemStatusUsed
	slot.itemOffset = p.header.freeSpaceUpper
	slot.itemLength = recordSize
	// we should be all done, return the slot
	return slot
}

// addNewPageSlotRecord appends a new pageSlot to the slots list,
// and it updates the pageHeader in memory, incrementing
// the slotCount, growing the freeSpaceLower bound and
// shrinking the freeSpaceUpper bound. addNewPageSlot returns
// a pointer to the newly added pageSlot.
func (p *Page) addNewPageSlotRecord(recordSize uint16) *pageSlot {
	// increment the slot count
	p.header.slotCount++
	// raise the free space lower bound
	// because we are adding a new slot
	p.header.freeSpaceLower += pageSlotSize
	// lower the free space upper bound
	// because we are adding record data
	p.header.freeSpaceUpper -= recordSize
	// create a new Page slot recording
	// the byte offset where the record
	// will be copied to within the Page
	// along with the length of the record
	p.slots = append(
		p.slots, &pageSlot{
			itemID:     p.header.slotCount - 1,
			itemStatus: itemStatusUsed,
			itemOffset: p.header.freeSpaceUpper,
			itemLength: recordSize,
		},
	)
	// return the last pageSlot we entered
	return p.slots[p.header.slotCount-1]
}

// SortRecords is a convenience wrapper for
// the internal sortSlotsByRecordPrefix call
func (p *Page) SortRecords() {
	p.sortSlotsByRecordPrefix()
}

// AddRecord adds a new record to the Page, if
// there is not enough room for the record to
// fit within the Page or the remaining Page's
// available space, an error will be returned.
//
// **It should be noted that (on insertion of
// a record) all pages slots are sorted
// lexicography by the prefix of the record
// data that they point to.
func (p *Page) AddRecord(r []byte) (*RecordID, error) {
	// get record size for check
	recordSize := uint16(len(r))
	// run the necessary checks on the record
	// to make sure we are good to go
	err := p.CheckRecord(recordSize)
	if err != nil {
		return nil, err
	}
	// get a fresh (or used free one, if there
	// are any) and update the Page header
	s := p.getAvailableSlot(recordSize)
	// get the new record offsets
	beg, end := s.itemBounds()
	// copy the record to the Page
	copy(p.data[beg:end], r)
	// before we return (this does not affect
	// the slotID) we should sort the slot
	// pointers, so all the record pointers
	// are in the proper order.
	p.sortSlotsByRecordPrefix()
	// all went well, return the newly added
	// records' id along with a nil error
	return &RecordID{
		PageID: p.header.pageID,
		SlotID: s.itemID,
	}, nil
}

// recordIDIsValid reports whether the
// provided *RecordID is valid or invalid
func (p *Page) recordIDIsValid(rid *RecordID) bool {
	return rid.PageID == p.header.pageID && int(rid.SlotID) < len(p.slots)
}

// GetRecord attempts to return the record data
// for a record found within this *Page using the
// provided *RecordID. If the record cannot be
// located, nil data and an error will be returned
func (p *Page) GetRecord(rid *RecordID) ([]byte, error) {
	// check to make sure the RecordID
	// is not an invalid record id
	if !p.recordIDIsValid(rid) {
		return nil, ErrInvalidRecordID
	}
	// locate the proper slot in the
	// Page using the supplied *RecordID
	slot := p.slots[rid.SlotID]
	// check the item status in the found slot
	// to ensure it has not already been marked
	// as a free slot (aka, can still be used)
	if slot.itemStatus == itemStatusFree {
		// item status has been marked free
		// which means it has been freed up
		// or removed
		return nil, ErrRecordHasBeenMarkedFree
	}
	// create a new buffer to copy the
	// record data into (so we are not
	// returning a pointer to the base
	// data, which would be unsafe)
	data := make([]byte, slot.itemLength)
	// get the record offsets for
	// an easier time copying
	beg, end := slot.itemBounds()
	// copy the record data into the
	// newly created buffer, and return
	copy(data, p.data[beg:end])
	// return the record data along
	// with a nil error
	return data, nil
}

// DelRecord removes a record from a Page. It will
// preserve the slot for later use.
func (p *Page) DelRecord(rid *RecordID) error {
	// check to make sure the RecordID
	// is not an invalid record id
	if !p.recordIDIsValid(rid) {
		return ErrInvalidRecordID
	}
	// locate the proper slot in the
	// Page using the supplied *RecordID
	slot := p.slots[rid.SlotID]
	// check the item status in the found slot
	// to ensure it has not already been marked
	// as a free slot (aka, can still be used)
	if slot.itemStatus == itemStatusFree {
		// item status has been marked free
		// which means it has already been
		// freed up or removed, nothing
		// else to do here, our job is done.
		return nil
	}
	// otherwise, we must now mark the found
	// slot as a free item which is now the
	// in pool to be re-used at a later date.
	slot.itemStatus = itemStatusFree
	// next, we should overwrite the item
	// record with zero's to minimize the
	// potential for data corruption if
	// the space is ever reused or compacted.
	zeros := make([]byte, slot.itemLength)
	// get the record offsets for
	// an easier time copying
	beg, end := slot.itemBounds()
	// copy the record data into the
	// newly created buffer, and return
	copy(p.data[beg:end], zeros)
	// make sure to increment the free
	// slot count
	p.header.freeSlotCount++
	// return the record data along
	// return a nil error
	return nil
}

// Range is a record iterator method for a Page's records
func (p *Page) Range(fn func(rid *RecordID) bool) {
	for i := range p.slots {
		if p.slots[i].itemStatus == itemStatusFree {
			continue
		}
		if !fn(
			&RecordID{
				PageID: p.header.pageID,
				SlotID: uint16(i),
			},
		) {
			break
		}
	}
}

// Reset resets the Page, all data and header information
// will return to the same state it was in when it was created.
func (p *Page) Reset() {
	// TODO: implement...
}

// String is a Page stringer method
func (p *Page) String() string {
	printHeader := struct {
		PageID         uint32 `json:"page_id"`
		NextPageID     uint32 `json:"next_page_id"`
		FreeSpaceLower uint16 `json:"free_space_lower"`
		FreeSpaceUpper uint16 `json:"free_space_upper"`
		SlotCount      uint16 `json:"slot_count"`
		FreeSlotCount  uint16 `json:"free_slot_count"`
	}{
		PageID:         p.header.pageID,
		NextPageID:     p.header.nextPageID,
		FreeSpaceLower: p.header.freeSpaceLower,
		FreeSpaceUpper: p.header.freeSpaceUpper,
		SlotCount:      p.header.slotCount,
		FreeSlotCount:  p.header.freeSlotCount,
	}
	b, err := json.MarshalIndent(printHeader, "", "\t")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
