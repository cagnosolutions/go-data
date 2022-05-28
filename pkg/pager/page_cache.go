package pager

type frame struct {
	// metadata
	// embedded page
}

// pageCache is a definition of a page buffer pool and all the basic functions that
// it is required to provide.
type pageCache interface {

	// newPage allocates a new page in the page cache using the provided pageID. The
	// page will be allocated and loaded into a frame that will then be returned. It
	// should be noted that allocating a new page from the page cache does not mean
	// it resides on the disk. It will remain in the page cache until it is flushed.
	newPage(pid uint32) *frame

	// fetchPage attempts to fetch the frame (that contains the requested page) from
	// the page cache. If the page is not currently loaded into a frame, it will be
	// located on disk by the disk manager and loaded into a frame, which will then
	// be returned. If there are no available frames, one must be victimized. The pin
	// count on the frame will be incremented every time fetchPage is called.
	fetchPage(pid uint32) *frame

	// flushPage attempts to flush the page that has a pageID that matches the provided
	// one. If there is not a matching page that is currently loaded into a frame, then
	// flushPage will return a ErrPageNotFound error. If a loaded page is not considered
	// dirty it will simply return a nil error. Once the page is flushed, the pin count
	// decrements by one. **Consider only flushing the page if the pin count is 1 or 0.
	flushPage(pid uint32) error

	// deletePage attempts to remove the page that has a pageID that matches the provided
	// one. If there is not a matching page that is currently loaded into a frame, then
	// deletePage will simply return a nil error. If the pin count on the found page is
	// greater than zero, deletePage will return a ErrPageInUse error. Assuming the page
	// is located, and there are no errors, the page will be removed from the page cache.
	// It should be noted that this will NOT remove the page from the disk.
	deletePage(pid uint32) error

	// unpinPage indicates the page is no longer in use by the requesting thread and is making
	// a request for the pin count to decrement. It instructs the page cache to locate the page
	// that has a pageID matching the one provided. If one cannot be found, then a ErrPageNotFound
	// error is returned. If the page is located, the pin count is decremented and if it reaches
	// zero (it means no more threads are using this page) the page is considered for eviction and
	// added to the replacer.
	unpinPage(pid uint32, isDirty bool) error
}
