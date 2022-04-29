package spage

const (
	// used in Page
	pageSize       = 8 << 10 // 8 KB
	pageHeaderSize = 24      // 24 bytes
	pageSlotSize   = 8       // 8 bytes
	MinRecordSize  = pageSlotSize
	MaxRecordSize  = pageSize - pageHeaderSize - pageSlotSize

	// used in Cache
	cacheSize      = 2 << 20
	cachePageCount = cacheSize / pageSize

	// used in PageBuffer
	defaultBufferedPageCount = 8
)

const (
	itemStatusFree uint16 = iota
	itemStatusUsed
)

func align(n int, size int) int {
	return (n + size) &^ size
}

const (
	// header offsets within page
	offPageID         = 0  // +4
	offNextPageID     = 4  // +4
	offPrevPageID     = 8  // +4
	offFreeSpaceLower = 12 // +2
	offFreeSpaceUpper = 14 // +2
	offSlotCount      = 16 // +2
	offFreeSlotCount  = 18 // +2
	offHasOverflow    = 20 // +2
	offReserved       = 22 // +2
	offStartSlots     = 24

	// entry offsets within slot space
	offSlotEntryID     = 0 // +2
	offSlotEntryStatus = 2 // +2
	offSlotEntryOffset = 4 // +2
	offSlotEntryLength = 6 // +2
)
