package engine

import (
	"sync"
)

// pageCache is the access level structure wrapping up the bufferPool, and DiskManager,
// along with a page table, and replacement policy.
type pageCache struct {
	latch     sync.Mutex
	pool      []frame            // buffer pool page frames
	replacer  *ClockReplacer     // page replacement policy structure
	io        *DiskManager       // underlying current manager
	freeList  []frameID          // list of frames that are free to use
	pageTable map[PageID]frameID // table of the current page to frame mappings
}

// openPageCache opens an existing storage manager instance if one exists with the same namespace
// otherwise it creates a new instance and returns it.
func openPageCache(base string, pageCount uint16) (*pageCache, error) {
	// open current manager
	fm, err := OpenDiskManager(base)
	if err != nil {
		return nil, err
	}
	// create buffer manager instance
	bm := &pageCache{
		pool:      make([]frame, pageCount, pageCount),
		replacer:  NewClockReplacer(pageCount),
		io:        fm,
		freeList:  make([]frameID, pageCount),
		pageTable: make(map[PageID]frameID),
	}
	// initialize the pool in the buffer manager
	for i := uint16(0); i < pageCount; i++ {
		bm.pool[i] = frame{
			pid:      0,
			fid:      0,
			pinCount: 0,
			isDirty:  false,
			page:     nil,
		}
		bm.freeList[i] = frameID(i)
	}
	// return buffer manager
	return bm, nil
}

// newPage returns a fresh empty page from the pool.
func (pc *pageCache) newPage() page {
	// First we must acquire a Frame in order to store our page. Calling
	// GetUsableFrame first checks our freeList and if we cannot find one in there
	// our replacement policy is used to locate a victimized one Frame.
	fid, err := pc.getUsableFrameID()
	if err != nil {
		// This can happen when the PageCache is full, so let's make sure that
		// it's something like that, and not something more sinister.
		if len(pc.freeList) == 0 && pc.replacer.size() == 0 {
			return nil
		}
		// Nope, it's something more sinister... shoot.
		DefaultLogger.Panic("%s", err)
	}
	// Allocate (get the next sequential PageID) so we can use it to initialize
	// the next page we will use.
	pid := pc.io.AllocatePage()
	// Create a new Frame initialized with our PageID and Page.
	pf := newFrame(pid, *fid, PageSize)
	pg := newPage(uint32(pid), P_USED)
	copy(pf.page, pg)
	// Add an entry to our pageTable
	pc.pageTable[pid] = *fid
	// And update the pool
	pc.pool[*fid] = pf
	// Finally, return our Page for use
	return pf.page
}

// fetchPage retrieves specific page from the pool, or storage medium by the page ID.
func (pc *pageCache) fetchPage(pid PageID) page {
	// Check to see if the PageID is located in the pageTable.
	if fid, found := pc.pageTable[pid]; found {
		// We located it, so now we access the Frame and ensure that it will
		// not be a victim candidate by our replacement policy.
		pf := pc.pool[fid]
		pf.pinCount++
		pc.replacer.Pin(fid)
		// And now, we can safely return our Page.
		return pf.page
	}
	// A match was not found in our pageTable, so now we must swap the Page in
	// from io. But first, we must get a Frame to hold our Page. We will
	// call on GetUsableFrame to check our freeList, and then potentially move on to
	// return a victimized Frame if we need to.
	fid, err := pc.getUsableFrameID()
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
	err = pc.io.ReadPage(pid, data)
	if err != nil {
		// Something went terribly wrong if this happens.
		DefaultLogger.Panic("%s", err)
	}
	// Create a new frame, so we can copy the page data we just swapped
	// in from off the io and add the Frame to the pageTable.
	pf := newFrame(pid, *fid, PageSize)
	copy(pf.page, data)
	// Add the entry to our pageTable
	pc.pageTable[pid] = *fid
	// And update the pool
	pc.pool[*fid] = pf
	// Finally, return our Page for use
	return pf.page
}

// unpinPage allows for manual unpinning of a specific page from the pool by the page ID.
func (pc *pageCache) unpinPage(pid PageID, isDirty bool) error {
	// Check to see if the PageID is located in the pageTable.
	fid, found := pc.pageTable[pid]
	if !found {
		// We have not located it, we will return an error.
		return ErrPageNotFound
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// ensure that it can be used as a victim candidate by our replacement policy.
	pf := pc.pool[fid]
	pf.decrPinCount()
	if pf.pinCount <= 0 {
		// After we decrement the pin count, check to see if it is low enough to
		// completely unpin it, and if so, unpin it.
		pc.replacer.Unpin(fid)
	}
	// Now, check to see if the dirty bit needs to be set.
	if pf.isDirty || isDirty {
		pf.isDirty = true
		return nil
	}
	// If not, we can make sure to unset the dirty bit.
	pf.isDirty = false
	return nil
}

// flushPage forces a page to be written onto the storage medium, and decrements the
// pin count on the frame potentially enabling the frame to be reused.
func (pc *pageCache) flushPage(pid PageID) error {
	// Check to see if the PageID is located in the pageTable.
	fid, found := pc.pageTable[pid]
	if !found {
		// We have not located it, we will return an error.
		return ErrPageNotFound
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// ensure that it can be used as a victim candidate by our replacement policy.
	pf := pc.pool[fid]
	pf.decrPinCount()
	// Now, we can make sure we flush it to the io using the DiskManager.
	err := pc.io.WritePage(pf.pid, pf.page)
	if err != nil {
		// Something went terribly wrong if this happens.
		DefaultLogger.Panic("%s", err)
	}
	// Finally, since we have just flushed the Page to the underlying current, we
	// can proceed with unsetting the dirty bit.
	pf.isDirty = false
	return nil
}

// deletePage removes the page from the buffer pool, and decrements the pin count on the
// frame potentially enabling the frame to be reused.
func (pc *pageCache) deletePage(pid PageID) error {
	// Check to see if the PageID is located in the pageTable.
	fid, found := pc.pageTable[pid]
	if !found {
		// We have not located it, but we don't need to return any error
		return nil
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// check to see if it is currently pinned (indicating it is currently in use
	// elsewhere) and should therefore not be removed just yet.
	pf := pc.pool[fid]
	if pf.pinCount > 0 {
		// Page must be in use elsewhere (or has otherwise not been properly
		// unpinned) so we'll return an error for now.
		return ErrPageInUse
	}
	// Now, we have our frame, and it is not currently in use, so first we will
	// remove it from the pageTable
	delete(pc.pageTable, pid)
	// Next, we pin it, so it will not be marked as a potential victim--because we
	// are in the process of remove it altogether.
	pc.replacer.Pin(fid)
	// After it is pinned, we will deallocate the Page on io (which will make
	// it free to use again in a pinch.)
	if err := pc.io.DeallocatePage(pid); err != nil {
		// Ops, something went down io side, return error
		return err
	}
	// Finally, add the current FrameID back onto the free list and return nil
	pc.addFrameID(fid)
	return nil
}

// getFrameID attempts to return a *FrameID. It first checks the freeList set to
// see if there are any available frames to pick from. If not, it will proceed to use
// the replacement policy to locate one.
func (pc *pageCache) getFrameID() (*frameID, bool) {
	// Check the freeList first, and if it is not empty return one
	if len(pc.freeList) > 0 {
		fid, newFreeList := pc.freeList[0], pc.freeList[1:]
		pc.freeList = newFreeList
		return &fid, true // true == fromFreeList
	}
	// Otherwise, there is nothing for us in the free list, so it's time to use our
	// replacement policy
	return pc.replacer.Victim(), false
}

// addFrameID takes a FrameID and adds it back onto our freeList for later use.
func (pc *pageCache) addFrameID(fid frameID) {
	pc.freeList = append(pc.freeList, fid)
}

// getUsableFrameID attempts to return a usable FrameID. It is used in the event
// that the buffer pool is "full." It always checks the free list first, and then it
// will fall back to using the replacer.
func (pc *pageCache) getUsableFrameID() (*frameID, error) {
	// First, we will check our freeList.
	fid, foundInFreeList := pc.getFrameID()
	if fid == nil {
		return nil, ErrUsableFrameNotFound
	}
	// If we find ourselves here we have a FrameID, but we do not yet know if
	// it came from our free list, or from our replacement policy. If it came due to
	// our replacement policy, then we will need to check to make sure it has not
	// been marked dirty, otherwise we must flush the contents to io before reusing
	// the FrameID; so let us check on that.
	if !foundInFreeList {
		cf := pc.pool[*fid]
		if &cf != nil {
			// We've located the correct Frame in the pool.
			if cf.isDirty {
				// And it appears that it is in fact holding a dirty Page. We must
				// flush the dirty Page to io before recycling this Frame.
				err := pc.io.WritePage(cf.pid, cf.page)
				if err != nil {
					return nil, err
				}
			}
			// In either case, we will now be able to remove this pageTable mapping
			// because it is no longer valid, and the caller should be creating a new
			// entry soon regardless.
			delete(pc.pageTable, cf.pid)
		}
	}
	// Finally, return our *FrameID, and a nil error
	return fid, nil
}

// flushAll attempts to flush any dirty page data.
func (pc *pageCache) flushAll() error {
	// We will range all the entries in the pageTable and call Flush on each one.
	for pid := range pc.pageTable {
		err := pc.flushPage(pid)
		if err != nil {
			return err
		}
	}
	return nil
}

// close attempts to close the pageCache along with the underlying DiskManager
// and associated dependencies. close makes sure to flush any dirty page data
// before closing everything down.
func (pc *pageCache) close() error {
	// Make sure all dirty Page data is written
	err := pc.flushAll()
	if err != nil {
		return err
	}
	// close the DiskManager
	err = pc.io.Close()
	if err != nil {
		return err
	}
	return nil
}

func (pc *pageCache) JSON() string {
	return ""
}
