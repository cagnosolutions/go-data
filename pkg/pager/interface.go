package pager

// BufferPoolManager is a buffer pool implementation that provides working memory
// and cache for operations for disk resident files. It is responsible for moving
// physical pages back and forth from main memory to disk.
type BufferPoolManager interface {

	// NewPage allocates and returns a new page with the provided page ID. The page
	// that is returned will remain in memory until it is flushed at which point it
	// will be disk resident. In the event that there is no more room in the memory
	// pool, an unused page frame will be selected from the free list or the replacer,
	// and it will be flushed (if necessary) and evicted to make room.
	NewPage(pid pageID) (page, error)

	// FetchPage tries to locate and return the page matching the provided page ID.
	// If the page is not currently loaded into memory, it will be paged in from the
	// DiskManager. In the event that there is no more room in the memory pool, an
	// unused page frame will be selected from the free list or the replacer, and it
	// will be flushed (if necessary) and evicted to make room.
	FetchPage(pid pageID) (page, error)

	// FlushPage tries to locate and flush the page matching the provided page ID.
	// If a matching page is not found an ErrPageNotFound error will be returned. It
	// passes the dirty, memory resident page to the DiskManager and requests it to
	// write. Any errors encountered by the DiskManager call will be returned. Once
	// the page is flushed, the pin count decreases by one.
	FlushPage(pid pageID) error

	// DeletePage tries to locate and remove the page matching the provided page ID.
	// If the pin count on the found page is greater than zero, an ErrPageInUse error
	// will be returned. If there are no errors encountered the page will be removed
	// from the resident memory, and the frame previously holding the page will be
	// added to the free list. This function will NOT remove the logical page data
	// from the disk.
	DeletePage(pid pageID) error

	// UnpinPage tries to locate and decrement the pin count for the page matching the
	// provided page ID. If the page could not be found, an ErrPageNotFound error is
	// returned. If the pin count reaches zero, the page frame is put in the replacer
	// and, it is now considered for eviction. The isDirty flag will indicate that this
	// page frame MUST be flushed to disk before being used as a victim.
	UnpinPage(pid pageID, isDirty bool) error
}

type DiskManager interface {

	// AllocateSegment allocates a segment in the file on disk . A segment is 2MB.
	// of 512 pages. If the page is 4KB, a segment is about 2MB.
	AllocateSegment() error

	// AllocatePage tries to allocate a new page on disk.
	AllocatePage() (pageID, error)

	DeAllocatePage(pid pageID) error

	// WritePage tries to write the page data to the disk. It uses the provided page ID
	// to calculate the logical offset of the disk resident page. It attempts to write
	// the data from the provided page. If no errors are encountered, the function will
	// return a nil error.
	WritePage(pid pageID, pg page) error

	// ReadPage tries to read the page data from the disk. It uses the provided page ID
	// to calculate the logical offset of the disk resident page. It attempts to read
	// the data into the provided page. If no errors are encountered, the function will
	// return a nil error.
	ReadPage(pid pageID, pg page) error

	// Close will safely synchronize and shutdown all file streams.
	Close() error
}
