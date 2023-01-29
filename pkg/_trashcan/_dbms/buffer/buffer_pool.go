package buffer

import (
	"errors"

	"github.com/cagnosolutions/go-data/pkg/dbms/buffer/prp"
	"github.com/cagnosolutions/go-data/pkg/dbms/disk"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

// BufferPoolManager represents the buffer pool manager
type BufferPoolManager struct {
	dm       disk.DiskManager            // manages I/O for pages on disk
	pages    []*page.Page                // list of pages in memory
	replacer *prp.ClockReplacer          // page replacement policy interface
	free     []prp.FrameID               // free list of page frames
	pt       map[page.PageID]prp.FrameID // page table mapping
}

// NewBufferPoolManager returns a empty buffer pool manager
func NewBufferPoolManager(size uint32, dm disk.DiskManager) *BufferPoolManager {
	b := &BufferPoolManager{
		dm:       dm,
		pages:    make([]*page.Page, size),
		replacer: prp.NewClockReplacer(size),
		free:     make([]prp.FrameID, size),
		pt:       make(map[page.PageID]prp.FrameID),
	}
	for i := uint32(0); i < size; i++ {
		b.free[i] = prp.FrameID(i)
		b.pages[i] = nil
	}
	return b
}

// FetchPage fetches the requested page from the buffer pool. If the page is in cache,
// it is returned immediately. If not, it will be found on disk, loaded into the cache
// and returned.
func (b *BufferPoolManager) FetchPage(pid page.PageID) *page.Page {
	// If the requested page is currently in the page table; return it.
	if fid, found := b.pt[pid]; found {
		pg := b.pages[fid]
		pg.IncPinCount()
		(*b.replacer).Pin(fid)
		return pg
	}
	// Otherwise, we will request a new page fame that we can use to load the page into.
	fid, fromFree := b.getFrameID()
	if fid == nil {
		// The buffer is full, it can't find a frame
		// NOTE: handle error? potentially return page not found?
		return nil
	}
	// We should now have a frame in which we can load a page into. If we did not find
	// an open frame in the free list we will proceed by attempting to victimize the
	// current frame.
	if !fromFree {
		// Remove page from current frame.
		pg := b.pages[*fid]
		if pg != nil {
			// If the page is dirty, flush page data to disk before
			// removing the page from the page table.
			if pg.IsDirty() {
				// NOTE: consider handling error here?
				b.dm.WritePage(pg.ID(), pg.Data()[:])
			}
			// And remove page from the page table.
			delete(b.pt, pg.ID())
		}
	}
	// At this point we will be reading the page from the underlying medium using the
	// disk manager. First we allocate a bew page sized buffer, and read our page into
	// the buffer.
	data := make([]byte, page.PageSize)
	err := b.dm.ReadPage(pid, data)
	if err != nil {
		// NOTE: handle error?
		return nil
	}
	// Next we ensure that the data we read fits a page sized chunk and allocated a
	// fresh new page to copy the data from the disk manager into.
	var pageData [page.PageSize]byte
	copy(pageData[:], data)
	pg := page.NewPage(pid, false, &pageData)
	// Finally, we update our page table entry, and add our page to our page set.
	b.pt[pid] = *fid
	b.pages[*fid] = pg
	// Return our "fetched" page.
	return pg
}

// UnpinPage unpins the target page from the buffer pool. It indicates that the page
// is not used any more for the current requesting thread. If no more threads are
// using this page, the page is considered for eviction (victim).
func (b *BufferPoolManager) UnpinPage(pid page.PageID, isDirty bool) error {
	// Attempt to locate the correct page frame using the page table.
	fid, found := b.pt[pid]
	if !found {
		return ErrPageNotFound
	}
	// Grab the page and decrement the pin count.
	pg := b.pages[fid]
	pg.DecPinCount()
	// If the pin count is less than one, we can simply unpin the frame.
	if pg.PinCount() <= 0 {
		(*b.replacer).Unpin(fid)
	}
	// Set the dirty status on the page. NOTE: potentially refactor this dirty call.
	if pg.IsDirty() || isDirty {
		pg.SetIsDirty(true)
	} else {
		pg.SetIsDirty(false)
	}
	return nil
}

// FlushPage flushes the target page that is in the cache onto the underlying medium.
func (b *BufferPoolManager) FlushPage(pid page.PageID) bool {
	// Attempt to locate the correct page frame using the page table.
	fid, found := b.pt[pid]
	if !found {
		// the buffer is full, it can't find a frame
		return false
	}
	// Grab the page and decrement the pin count.
	pg := b.pages[fid]
	pg.DecPinCount()
	// Write the page data to the underlying medium via the disk manager.
	b.dm.WritePage(pid, pg.Data()[:])
	// Unset dirty bit
	pg.SetIsDirty(false)
	// Return success!
	return true
}

// NewPage allocates a new page in the buffer pool with the disk manager help
func (b *BufferPoolManager) NewPage() *page.Page {
	// Request a new page fame that we can use to load the page into.
	fid, fromFree := b.getFrameID()
	if fid == nil {
		// the buffer is full, it can't find a frame
		return nil
	}
	// We should now have a frame in which we can load a page into. If we did not find
	// an open frame in the free list we will proceed by attempting to victimize the
	// current frame.

	if !fromFree {
		// Remove page from current frame.
		pg := b.pages[*fid]
		if pg != nil {
			// If the page is dirty, flush page data to disk before
			// removing the page from the page table.
			if pg.IsDirty() {
				// NOTE: consider handling error here?
				b.dm.WritePage(pg.ID(), pg.Data()[:])
			}
			// And remove page from the page table.
			delete(b.pt, pg.ID())
		}
	}
	// Allocate a new page.
	pid := b.dm.AllocatePage()
	pg := page.NewEmptyPage(pid)

	// Update page table before returning.
	b.pt[pid] = *fid
	b.pages[*fid] = pg

	return pg
}

// DeletePage deletes a page from the buffer pool.
func (b *BufferPoolManager) DeletePage(pid page.PageID) error {
	var frameID prp.FrameID
	var ok bool
	if frameID, ok = b.pt[pid]; !ok {
		return nil
	}
	page := b.pages[frameID]
	if page.PinCount() > 0 {
		return errors.New("Pin count greater than 0")
	}
	delete(b.pt, page.ID())
	(*b.replacer).Pin(frameID)
	b.dm.DeallocatePage(pid)

	b.free = append(b.free, frameID)

	return nil

}

// FlushAllPages flushes all the pages in the buffer pool to disk.
func (b *BufferPoolManager) FlushAllPages() {
	for pageID := range b.pt {
		b.FlushPage(pageID)
	}
}

// getFrameID returns a frame ID from the free list, or by using the
// replacement policy if the free list is full along with a boolean
// indicating true if the frame ID was returned using the free list
// and false if it was returned by using the replacement policy.
func (b *BufferPoolManager) getFrameID() (*prp.FrameID, bool) {
	// Check the free list. If there is no room use the replacement
	// policy to return the next victim.
	if len(b.free) < 1 {
		return (*b.replacer).Victim(), false
	}
	// Otherwise, get the oldest frame in the list, and update list.
	fid, newFree := b.free[0], b.free[1:]
	b.free = newFree
	// Finally, return the frame ID.
	return &fid, true
}

func (b *BufferPoolManager) _getFrameID() (*prp.FrameID, bool) {
	if len(b.free) > 0 {
		fid, newFree := b.free[0], b.free[1:]
		b.free = newFree
		return &fid, true
	}
	return (*b.replacer).Victim(), false
}
