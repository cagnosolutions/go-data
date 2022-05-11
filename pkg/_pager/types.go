package _pager

type (
	PageID   uint16
	RecordID uint32
	FrameID  uint32
)

type Slot interface {
	GetRecord() (uint16, uint16)
}

type PageHeader interface {
	FreeSpace() uint16
}

type Page interface {
	GetPageHeader() PageHeader
	SetPageHeader(ph PageHeader)
	GetPageID() PageID
	IsDirty() bool
	Data() []byte
}

// Replacer represents a page replacement strategy.
type Replacer interface {

	// Victim removes the victim frame as defined by the replacement
	// policy.
	Victim() *FrameID

	// Unpin unpins a frame, indicating that it can now be victimized.
	Unpin(id FrameID)

	// Pin pins a frame, indicating that it should not be victimized
	// until it is unpinned.
	Pin(id FrameID)

	// Size returns the size of the replacer.
	Size() uint32

	// SetSize sets the replacement size.
	SetSize(size uint32)
}

// A PageBuffer works in conjunction with the DiskManager to provide
// full page and memory management. The page buffer maintains a page
// table and holds the page frames in main memory. It utilizes calls
// made by the DiskManager in order to allocate new pages, and read and
// write pages.
type PageBuffer interface {

	// NewPage allocates a new page and pins it to a frame. If we did not
	// find an open frame we will proceed by attempting to victimize the
	// current frame.
	NewPage() Page

	// FetchPage fetches the requested page from the buffer pool. If the
	// page is in cache, it is returned immediately. If not, it will be
	// found by the DiskManager and loaded into the cache and returned.
	FetchPage(pid PageID) Page

	// UnpinPage unpins the target page from the buffer pool. It indicates
	// that the page is not used any more for the current requesting thread.
	// If no more threads are using this page, the page is considered for
	// eviction (victim).
	UnpinPage(pid PageID) error

	// FlushPage flushes the target page that is in the cache onto the
	// underlying medium. It also decrements the pin count and unpins it
	// from the holding frame and unsets the dirty bit.
	FlushPage(pid PageID) bool

	// DeletePage deletes a page from the buffer pool. Once removed, it
	// marks the holding fame as free to use.
	DeletePage(pid PageID) error

	// GetFrameID returns a frame ID from the free list, or by using the
	// replacement policy if the free list is full along with a boolean
	// indicating true if the frame ID was returned using the free list
	// and false if it was returned by using the replacement policy.
	GetFrameID() (FrameID, bool)
}

// A DiskManager manages the creation and the swapping in and out of pages
// to and from the underlying medium. It is responsible for maintaining
// the sequential PageID's and providing save reads and writes of entire
// pages. It also allocates space in the underlying medium.
type DiskManager interface {

	// AllocatePage creates a new page on the underlying medium and returns
	// the PageID of the freshly allocated page. It will grow the underlying
	// medium if necessary.
	AllocatePage() PageID

	// ReadPage reads a page from the underlying medium by mapping the PageID
	// provided to a page offset. It reads the page into the dst space which
	// is typically going to be a frame maintained by the PageBuffer. If the
	// provided PageID cannot be located, an error will be returned.
	ReadPage(pid PageID, dst []byte) error

	// WritePage writes a page to the underlying medium by mapping the PageID
	// provided to a page offset. It writes the page from the src space which
	// is typically going to be a frame maintained by the PageBuffer. If the
	// provided PageID cannot be located, an error will be returned.
	WritePage(pid PageID, src []byte) error

	// DeallocatePage marks the page mapped to by the provided PageID as a free
	// page that can be used later on.
	DeallocatePage(pid PageID)

	// GetNumReads returns the number of reads the DiskManager has performed.
	GetNumReads() uint

	// GetNumWrites returns the number of writes the DiskManager has performed.
	GetNumWrites() uint

	// ShutDown gracefully shuts down the DiskManager. Any allocated pages that
	// are being held by the PageBuffer that are dirty or pinned are not always
	// guaranteed to be synchronized on the underlying medium unless ShutDown
	// is called by the PageBuffer.
	ShutDown() error

	// Size returns the size (in bytes) of the underlying medium.
	Size() int64
}
