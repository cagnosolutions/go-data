package pager

type (
	PageID  = pageID
	FrameID = frameID
	Page    = page
)

// BufferPoolManager is a buffer pool implementation that provides working
// memory and cache for operations for sections of memory called pages. It
// is responsible working in conjunction with a StorageManager instance in
// order to accomplish tings such as moving logical page data back and forth
// from an underlying storage medium into main memory and back.
type BufferPoolManager interface {
	// NewPage allocates and returns a new page of data.
	NewPage() (Page, error)

	// FetchPage locates and returns the page matching the supplied page ID.
	FetchPage(pid PageID) (Page, error)

	// UnpinPage locates the page matching the supplied page ID and lets the
	// buffer pool that we are finished using it. A page should be flushed
	// before UnpinPage is call in most situations.
	UnpinPage(pid PageID, isDirty bool) error

	// FlushPage flushes any changes made to the page back to the underlying
	// storage manager.
	FlushPage(pid PageID) error

	// DeletePage locates the page matching the supplied page ID and removes
	// it from the buffer pool. It also removes the page from the underlying
	// storage manager.
	DeletePage(pid PageID) error

	// FlushAll is just like FlushPage except for it will flush all the pages
	// that are currently resident in the buffer pool.
	FlushAll() error
}

// StorageManager is a storage manager type that may have many kinds of
// implementations. It is most often seen as a long-term storage manager
// such as a manager based storage manager.
type StorageManager interface {
	// Allocate attempts to allocate and return the next ID in the sequence.
	Allocate() PageID

	// Deallocate wipes the data in the location matching the provided
	// ID or offset. If the ID is bad, or out of bounds, it will return
	// an error.
	Deallocate(pid PageID) error

	// ReadPage uses the provided page ID to calculate the logical offset
	// where the page should be located and will attempt to read the page
	// data from that location into the provided page. Any errors encountered
	// will be returned immediately.
	ReadPage(pid PageID, pg Page) error

	// WritePage uses the provided page ID to calculate the logical offset
	// where the page should be located and will attempt to write the page
	// data from the provided page to that location. Any errors encountered
	// will be returned immediately.
	WritePage(pid PageID, pg Page) error

	// Close will safely synchronize and shutdown all open file streams.
	Close() error

	// Size returns the number of bytes the storage manager is occupying.
	Size() int
}

// Replacer is a page replacement policy type. For more information on replacement
// policies, see [https://en.wikipedia.org/wiki/Page_replacement_algorithm]
type Replacer interface {
	// Pin takes a frame ID and "pins" it, indicating that the caller is now using
	// it. Because the caller is now using it, the replacer can now remove it to no
	// longer make it available for victimization.
	Pin(fid FrameID)

	// Unpin takes a frame ID and "unpins" it, indicating that the caller is no
	// longer using it. Because the caller is no longer using it, the replacer can
	// now add it to make it available for victimization.
	Unpin(fid FrameID)

	// Victim searches for a frame ID in the replacer that it can victimize and
	// return to the caller. It locates and removes a victim ID (as defined by the
	// policy) and returns it. If there are no frame IDs to victimize, it will simply
	// return nil. In the case of a nil return, the caller will have to figure out
	// how to handle the situation.
	Victim() *FrameID

	// Size returns the number of elements currently in the replacer.
	Size() int
}
