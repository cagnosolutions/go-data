package spage

// RecordID represents the
// unique id for a single
// data record held within
// a Page
type RecordID struct {
	PageID uint32
	SlotID uint16
}

// pageHeader is a header structure for a Page
type pageHeader struct {
	pageID         uint32
	nextPageID     uint32
	prevPageID     uint32
	freeSpaceLower uint16
	freeSpaceUpper uint16
	slotCount      uint16
	freeSlotCount  uint16
	hasOverflow    uint16
	reserved       uint16
}

// FreeSpace returns the total (contiguous) free
// space in bytes that is left in this Page
func (h *pageHeader) FreeSpace() uint16 {
	return h.freeSpaceUpper - h.freeSpaceLower // - (pageSlotSize * 1 * h.slotCount)
}

// PageIsFree reports if the Page has been allocated
// but is now available and free to use
func (h *pageHeader) PageIsFree() bool {
	return h.freeSlotCount == h.slotCount
}
