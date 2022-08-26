package engine

import (
	"sync"
)

// PageCache is the access level structure wrapping up the bufferPool, and DiskManager,
// along with a page table, and replacement policy.
type PageCache struct {
	latch     sync.Mutex
	pool      []frame            // buffer pool page frames
	replacer  *ClockReplacer     // page replacement policy structure
	io        *DiskManager       // underlying current manager
	freeList  []frameID          // list of frames that are free to use
	pageTable map[PageID]frameID // table of the current page to frame mappings
}

// OpenPageCache opens an existing storage manager instance if one exists with the same namespace
// otherwise it creates a new instance and returns it.
func OpenPageCache(base string, pageCount uint16) (*PageCache, error) {
	// open current manager
	fm, err := OpenDiskManager(base)
	if err != nil {
		return nil, err
	}
	// create buffer manager instance
	bm := &PageCache{
		pool:      make([]frame, pageCount, pageCount),
		replacer:  NewClockReplacer(pageCount),
		io:        fm,
		freeList:  make([]frameID, pageCount),
		pageTable: make(map[PageID]frameID),
	}
	// initialize the pool in the buffer manager
	for i := uint16(0); i < pageCount; i++ {
		bm.pool[i] = frame{
			PID:      0,
			FID:      0,
			PinCount: 0,
			IsDirty:  false,
			Page:     nil,
		}
		bm.freeList[i] = frameID(i)
	}
	// return buffer manager
	return bm, nil
}

// NewPage returns a fresh empty page from the pool.
func (m *PageCache) NewPage() Page {
	// First we must acquire a Frame in order to store our page. Calling
	// GetUsableFrame first checks our freeList and if we cannot find one in there
	// our replacement policy is used to locate a victimized one Frame.
	fid, err := m.getUsableFrameID()
	if err != nil {
		// This can happen when the PageCache is full, so let's make sure that
		// it's something like that, and not something more sinister.
		if len(m.freeList) == 0 && m.replacer.size() == 0 {
			return nil
		}
		// Nope, it's something more sinister... shoot.
		DefaultLogger.Panic("%s", err)
	}
	// Allocate (get the next sequential PageID) so we can use it to initialize
	// the next page we will use.
	pid := m.io.AllocatePage()
	// Create a new Frame initialized with our PageID and Page.
	pf := newFrame(pid, *fid, PageSize)
	pg := NewPage(pid)
	copy(pf.Page, pg)
	// Add an entry to our pageTable
	m.pageTable[pid] = *fid
	// And update the pool
	m.pool[*fid] = pf
	// Finally, return our Page for use
	return pf.Page
}

// FetchPage retrieves specific page from the pool, or storage medium by the page ID.
func (m *PageCache) FetchPage(pid PageID) Page {
	// Check to see if the PageID is located in the pageTable.
	if fid, found := m.pageTable[pid]; found {
		// We located it, so now we access the Frame and ensure that it will
		// not be a victim candidate by our replacement policy.
		pf := m.pool[fid]
		pf.PinCount++
		m.replacer.Pin(fid)
		// And now, we can safely return our Page.
		return pf.Page
	}
	// A match was not found in our pageTable, so now we must swap the Page in
	// from io. But first, we must get a Frame to hold our Page. We will
	// call on GetUsableFrame to check our freeList, and then potentially move on to
	// return a victimized Frame if we need to.
	fid, err := m.getUsableFrameID()
	if err != nil {
		// TODO: think about a more graceful way of handling this whole situation
		// Check the EXACT error
		if err != ErrUsableFrameNotFound {
			// Something went terribly wrong if this happens.
			DefaultLogger.Panic("%s", err)
		}
		return nil
	}
	// Now, we will swap the Page in from the io using the DiskManager.
	data := make([]byte, PageSize)
	err = m.io.ReadPage(pid, data)
	if err != nil {
		// Something went terribly wrong if this happens.
		DefaultLogger.Panic("%s", err)
	}
	// Create a new frame, so we can copy the page data we just swapped
	// in from off the io and add the Frame to the pageTable.
	pf := newFrame(pid, *fid, PageSize)
	copy(pf.Page, data)
	// Add the entry to our pageTable
	m.pageTable[pid] = *fid
	// And update the pool
	m.pool[*fid] = pf
	// Finally, return our Page for use
	return pf.Page
}

// UnpinPage allows for manual unpinning of a specific page from the pool by the page ID.
func (m *PageCache) UnpinPage(pid PageID, isDirty bool) error {
	// Check to see if the PageID is located in the pageTable.
	fid, found := m.pageTable[pid]
	if !found {
		// We have not located it, we will return an error.
		return ErrPageNotFound
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// ensure that it can be used as a victim candidate by our replacement policy.
	pf := m.pool[fid]
	pf.decrPinCount()
	if pf.PinCount <= 0 {
		// After we decrement the pin count, check to see if it is low enough to
		// completely unpin it, and if so, unpin it.
		m.replacer.Unpin(fid)
	}
	// Now, check to see if the dirty bit needs to be set.
	if pf.IsDirty || isDirty {
		pf.IsDirty = true
		return nil
	}
	// If not, we can make sure to unset the dirty bit.
	pf.IsDirty = false
	return nil
}

// FlushPage forces a page to be written onto the storage medium, and decrements the
// pin count on the frame potentially enabling the frame to be reused.
func (m *PageCache) FlushPage(pid PageID) error {
	// Check to see if the PageID is located in the pageTable.
	fid, found := m.pageTable[pid]
	if !found {
		// We have not located it, we will return an error.
		return ErrPageNotFound
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// ensure that it can be used as a victim candidate by our replacement policy.
	pf := m.pool[fid]
	pf.decrPinCount()
	// Now, we can make sure we flush it to the io using the DiskManager.
	err := m.io.WritePage(pf.PID, pf.Page)
	if err != nil {
		// Something went terribly wrong if this happens.
		DefaultLogger.Panic("%s", err)
	}
	// Finally, since we have just flushed the Page to the underlying current, we
	// can proceed with unsetting the dirty bit.
	pf.IsDirty = false
	return nil
}

// DeletePage removes the page from the buffer pool, and decrements the pin count on the
// frame potentially enabling the frame to be reused.
func (m *PageCache) DeletePage(pid PageID) error {
	// Check to see if the PageID is located in the pageTable.
	fid, found := m.pageTable[pid]
	if !found {
		// We have not located it, but we don't need to return any error
		return nil
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// check to see if it is currently pinned (indicating it is currently in use
	// elsewhere) and should therefore not be removed just yet.
	pf := m.pool[fid]
	if pf.PinCount > 0 {
		// Page must be in use elsewhere (or has otherwise not been properly
		// unpinned) so we'll return an error for now.
		return ErrPageInUse
	}
	// Now, we have our frame, and it is not currently in use, so first we will
	// remove it from the pageTable
	delete(m.pageTable, pid)
	// Next, we pin it, so it will not be marked as a potential victim--because we
	// are in the process of remove it altogether.
	m.replacer.Pin(fid)
	// After it is pinned, we will deallocate the Page on io (which will make
	// it free to use again in a pinch.)
	if err := m.io.DeallocatePage(pid); err != nil {
		// Ops, something went down io side, return error
		return err
	}
	// Finally, add the current FrameID back onto the free list and return nil
	m.addFrameID(fid)
	return nil
}

// getFrameID attempts to return a *FrameID. It first checks the freeList set to
// see if there are any available frames to pick from. If not, it will proceed to use
// the replacement policy to locate one.
func (m *PageCache) getFrameID() (*frameID, bool) {
	// Check the freeList first, and if it is not empty return one
	if len(m.freeList) > 0 {
		fid, newFreeList := m.freeList[0], m.freeList[1:]
		m.freeList = newFreeList
		return &fid, true // true == fromFreeList
	}
	// Otherwise, there is nothing for us in the free list, so it's time to use our
	// replacement policy
	return m.replacer.Victim(), false
}

// addFrameID takes a FrameID and adds it back onto our freeList for later use.
func (m *PageCache) addFrameID(fid frameID) {
	m.freeList = append(m.freeList, fid)
}

// getUsableFrameID attempts to return a usable FrameID. It is used in the event
// that the buffer pool is "full." It always checks the free list first, and then it
// will fall back to using the replacer.
func (m *PageCache) getUsableFrameID() (*frameID, error) {
	// First, we will check our freeList.
	fid, foundInFreeList := m.getFrameID()
	if fid == nil {
		return nil, ErrUsableFrameNotFound
	}
	// If we find ourselves here we have a FrameID, but we do not yet know if
	// it came from our free list, or from our replacement policy. If it came due to
	// our replacement policy, then we will need to check to make sure it has not
	// been marked dirty, otherwise we must flush the contents to io before reusing
	// the FrameID; so let us check on that.
	if !foundInFreeList {
		cf := m.pool[*fid]
		if &cf != nil {
			// We've located the correct Frame in the pool.
			if cf.IsDirty {
				// And it appears that it is in fact holding a dirty Page. We must
				// flush the dirty Page to io before recycling this Frame.
				err := m.io.WritePage(cf.PID, cf.Page)
				if err != nil {
					return nil, err
				}
			}
			// In either case, we will now be able to remove this pageTable mapping
			// because it is no longer valid, and the caller should be creating a new
			// entry soon regardless.
			delete(m.pageTable, cf.PID)
		}
	}
	// Finally, return our *FrameID, and a nil error
	return fid, nil
}

// flushAll attempts to flush any dirty Page data.
func (m *PageCache) flushAll() error {
	// We will range all the entries in the pageTable and call Flush on each one.
	for pid := range m.pageTable {
		err := m.FlushPage(pid)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close attempts to close the PageCache along with the underlying DiskManager
// and associated dependencies. Close makes sure to flush any dirty Page data
// before closing everything down.
func (m *PageCache) Close() error {
	// Make sure all dirty Page data is written
	err := m.flushAll()
	if err != nil {
		return err
	}
	// Close the DiskManager
	err = m.io.Close()
	if err != nil {
		return err
	}
	return nil
}

func (m *PageCache) JSON() string {
	return ""
}
