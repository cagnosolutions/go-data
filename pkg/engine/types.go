package engine

import "github.com/cagnosolutions/go-data/pkg/engine/page"

// bufferPool is an interface describing the basic operations that the buffer pool
// is responsible for handling. The bufferPool is used by the bufferPoolManager.
type bufferPool interface {
	// getFrameID attempts to return a frame.FrameID from the free list. If
	// one is found it will return it along with a boolean indicating true.
	getFrameID() (*frameID, bool)
	// addFrameID takes a frameID and adds it to the set of free frames list.
	addFrameID(fid frameID)
	// getUsableFrameID attempts to return a frame.FrameID. It first checks
	// calls getFrameID to try to return one from the freeList. If the first
	// call fails, it will then go on to the replacer in search of one.
	getUsableFrameID() (*frameID, error)
}

// replacer is an interface describing the basic operations that make up a replacement
// policy. The replacer is used by the bufferPoolManager.
type replacer interface {
	// Pin pins the frame matching the supplied frame ID, indicating that it should
	// not be victimized until it is unpinned.
	Pin(fid frameID)
	// Victim removes and returns the next "victim frame", as defined by the policy.
	Victim() *frameID
	// Unpin unpins the frame matching the supplied frame ID, indicating that it may
	// now be victimized.
	Unpin(fid frameID)
}

// diskManager is an interface describing the basic operations that the
// io manager is responsible for handling. The diskManager is usually
// something that is used by a bufferPoolManager.
type diskManager interface {
	// AllocatePage allocates and returns the next sequential page.PageID.
	// in some cases, if there are a lot of empty fragmented pages, it may
	// return a non-sequential page.PageID.
	AllocatePage() page.PageID
	// DeallocatePage takes a page.PageID and attempts to locate and mark
	// the associated page status as free to use in the future. The data
	// may be wiped, so this is a destructive call and should be used with
	// care.
	DeallocatePage(pid page.PageID) error
	// ReadPage takes a page.PageID, as well as a (preferably empty) page.Page,
	// attempts to locate and copy the contents into p.
	ReadPage(pid page.PageID, p page.Page) error
	// WritePage takes a page.PageID, as well as a page.Page, attempts to locate
	// and copy and flush the contents of p onto the io.
	WritePage(pid page.PageID, p page.Page) error
	// Close closes the io manager.
	Close() error
}

// bufferPoolManager is an interface for describing the basic operations that
// the buffer pool manager is responsible for handling.
type bufferPoolManager interface {
	bufferPool
	replacer
	diskManager
	// NewPage returns a new "fresh" page.Page for use.
	NewPage() page.Page
	// FetchPage takes a page.PageID, and attempts to locate it (either in the
	// buffer pool, or on io) and return the associated page.Page.
	FetchPage(pid page.PageID) page.Page
	// UnpinPage takes a page.PageID, and a boolean hinting at the page.Page
	// associated with the supplied page.pageID being dirty or not. It instructs
	// the replacer to unpin the page making it available for victimization.
	UnpinPage(pid page.PageID, isDirty bool) error
	// FlushPage takes a page.PageID, and attempts to locate and flush the
	// associated page.Page to io but, it does not remove it from the pageTable.
	FlushPage(pid page.PageID) error
	// DeletePage takes a page.PageID and attempts to locate and remove the
	// associated page.Page from the pageTable (if it is not pinned) and also
	// clears it on the io.
	DeletePage(pid page.PageID) error
	// Close flushes and dirty page.Page data to the underlying io, and then
	// shuts down the bufferPoolManager.
	Close() error
}
