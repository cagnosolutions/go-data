package engine

// // bufferPool is an interface describing the basic operations that the buffer pool
// // is responsible for handling. The bufferPool is used by the bufferPoolManager.
// type bufferPool interface {
// 	// getFrameID attempts to return a frame.FrameID from the free list. If
// 	// one is found it will return it along with a boolean indicating true.
// 	getFrameID() (*buffer.FrameID, bool)
// 	// addFrameID takes a FrameID and adds it to the set of free frames list.
// 	addFrameID(fid buffer.FrameID)
// 	// getUsableFrameID attempts to return a frame.FrameID. It first checks
// 	// calls getFrameID to try to return one from the freeList. If the first
// 	// call fails, it will then go on to the replacer in search of one.
// 	getUsableFrameID() (*buffer.FrameID, error)
// }

// // bufferPoolManager is an interface for describing the basic operations that
// // the buffer pool manager is responsible for handling.
// type bufferPoolManager interface {
// 	bufferPool
// 	replacer
// 	diskManager
// 	// newPage returns a new "fresh" page.Page for use.
// 	NewPage() Page
// 	// FetchPage takes a page.PageID, and attempts to locate it (either in the
// 	// buffer pool, or on io) and return the associated page.Page.
// 	FetchPage(pid PageID) Page
// 	// UnpinPage takes a page.PageID, and a boolean hinting at the page.Page
// 	// associated with the supplied page.pageID being dirty or not. It instructs
// 	// the replacer to unpin the page making it available for victimization.
// 	UnpinPage(pid PageID, isDirty bool) error
// 	// FlushPage takes a page.PageID, and attempts to locate and flush the
// 	// associated page.Page to io but, it does not remove it from the pageTable.
// 	FlushPage(pid PageID) error
// 	// DeletePage takes a page.PageID and attempts to locate and remove the
// 	// associated page.Page from the pageTable (if it is not pinned) and also
// 	// clears it on the io.
// 	DeletePage(pid PageID) error
// 	// Close flushes and dirty page.Page data to the underlying io, and then
// 	// shuts down the bufferPoolManager.
// 	Close() error
// }

// var InitializerError = func(err error) error {
// 	return fmt.Errorf("initializer error: %w", err)
// }
//
// type Initializer interface {
// 	Init(args any) error
// }
//
// // Loader is an interface that should have a method for loading
// // and a method to check to see if something is loaded.
// type Loader interface {
// 	// IsLoaded should return a boolean indicating the status
// 	// of whether Loader has successfully been called.
// 	IsLoaded() bool
//
// 	// Load should do any necessary work and return any errors
// 	// encountered.
// 	Load() error
// }
