package pager

type Pager struct {
	bp *bufferPoolManager
}

func NewPager(path string) *Pager {
	dm := newTempDiskManager(path)
	bp := newBufferPoolManager(16, dm)
	return &Pager{
		bp: bp,
	}
}

// NewPage allocates a new page in the page cache using the provided pageID. The
// page will be allocated and loaded into a frame that will then be returned. It
// should be noted that allocating a new page from the page cache does not mean
// it resides on the disk. It will remain in the page cache until it is flushed.
func (p *Pager) NewPage() (page, error) {
	pg := p.bp.newPage()
	if pg == nil {
		return nil, ErrPageNotFound
	}
	return pg, nil
}

// fetchPage attempts to fetch the frame (that contains the requested page) from
// the page cache. If the page is not currently loaded into a frame, it will be
// located on disk by the disk manager and loaded into a frame, which will then
// be returned. If there are no available frames, one must be victimized. The pin
// count on the frame will be incremented every time fetchPage is called. In the
// event in which a page must be victimized, it will first check to see if the
// frame considered for reuse is marked as dirty or not. If it is dirty, it will
// automatically be flushed before the frame is re-used.
func (p *Pager) fetchPage(pid uint32) (page, error) {
	return nil, nil
}

// flushPage attempts to flush the page that has a pageID that matches the provided
// one. If there is not a matching page that is currently loaded into a frame, then
// flushPage will return a ErrPageNotFound error. If a loaded page is not considered
// dirty it will simply return a nil error. Once the page is flushed, the pin count
// decrements by one. **Consider only flushing the page if the pin count is 1 or 0.
func (p *Pager) flushPage(pid uint32) error {
	return nil
}

// deletePage attempts to remove the page that has a pageID that matches the provided
// one. If there is not a matching page that is currently loaded into a frame, then
// deletePage will simply return a nil error. If the pin count on the found page is
// greater than zero, deletePage will return a ErrPageInUse error. Assuming the page
// is located, and there are no errors, the page will be removed from the page cache.
// It should be noted that this will NOT remove the page from the disk.
func (p *Pager) deletePage(pid uint32) error {
	return nil
}

// unpinPage indicates the page is no longer in use by the requesting thread and is making
// a request for the pin count to decrement. It instructs the page cache to locate the page
// that has a pageID matching the one provided. If one cannot be found, then a ErrPageNotFound
// error is returned. If the page is located, the pin count is decremented and if it reaches
// zero (it means no more threads are using this page) the page is considered for eviction.
// If the isDirty flag is set to true, it means that if this page frame is chosen as a victim,
// (in the event that all the page frames are in-use) the isDirty flag will indicate that this
// page frame MUST be flushed to disk before being used as a victim.
func (p *Pager) unpinPage(pid uint32, isDirty bool) error {
	return nil
}
