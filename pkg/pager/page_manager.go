package pager

import (
	"sync"
)

// pageManager is our implementation of a pageFrameManager, which is
// also sometimes called a buffer pool manager in a dbms system.
type pageManager struct {
	lock     sync.RWMutex
	frames   []*pageFrame       // list of loaded page frames
	replacer *clockReplacer     // used to find an unpinned page for replacement
	manager  *tempDiskManager   // underlying storage manager
	free     []frameID          // used to find a page for replacement
	table    map[pageID]frameID // used to keep track of pages
}

// newPageManager initializes and returns a new instance of a pageManager.
func newPageManager(size int, disk *tempDiskManager) *pageManager {
	bm := &pageManager{
		frames:   make([]*pageFrame, size, size),
		replacer: newClockReplacer(size),
		manager:  disk,
		free:     make([]frameID, size),
		table:    make(map[pageID]frameID),
	}
	for i := 0; i < size; i++ {
		bm.frames[i] = nil
		bm.free[i] = frameID(i)
	}
	return bm
}

// fetchPage fetches the requested page pageFrame from our page manager.
func (b *pageManager) fetchPage(pid pageID) *pageFrame {
	// Check to see if the pageFrame is already loaded in the
	// page pageFrame table.
	if fid, found := b.table[pid]; found {
		// It appears to be, so we should get the matching
		// page pageFrame and return it.
		pf := b.frames[fid]
		// Don't forget to increment the pin count, and
		// also mark it pinned in the replacer.
		pf.pinCount++
		b.replacer.pin(fid)
		// Finally, return the page.
		return pf
	}
	// We did not find the page in the page pageFrame table. So
	// now we must ensure that we have an empty pageFrame to load
	// the page into once we locate it. So, we will do this...

	// First we need a frameID we can use to proceed. We will call
	// getFrame which will first check our free list and if we do
	// not find one in there getFrame will victimize a pageFrame and
	// return a frameID we can use.
	fid, err := b.getFrame()
	if err != nil {
		// Something went wrong!
		panic(err)
	}
	// Now that we have an empty pageFrame in the table to load a pageFrame
	// into, we will make an attempt to locate the requested page
	// using the storage manager and read it into a new page pageFrame.
	pf := newFrame(pid)
	err = b.manager.read(pid, pf.page)
	if err != nil {
		// Something went wrong!
		panic(err)
	}
	// Looks like we found it and read it successfully. Now we can
	// add it to our page table.
	b.table[pf.pid] = *fid
	// Finally, add it to our active pageFrame set.
	b.frames[*fid] = pf
	// And return our page pageFrame.
	return pf
}

// unpinPage unpins the target page pageFrame from the pageFrameManager
func (b *pageManager) unpinPage(pid pageID, isDirty bool) error {
	// First we attempt to locate the requested page pageFrame in the
	// page table using the supplied pageID.
	fid, found := b.table[pid]
	if !found {
		// We have not found it, return an error.
		return ErrPageNotFound
	}
	// Otherwise, we have located the matching frameID for the page
	// we are looking for. Next we should get the actual page pageFrame
	// from our page pageFrame set.
	pf := b.frames[fid]
	// Decrement the pin count and check if we are able to unpin it.
	pf.decrPinCount()
	if pf.pinCount == 0 {
		b.replacer.unpin(fid)
	}
	// Next, we must check the page pageFrame to see if the dirty bit
	// needs to be set.
	if pf.isDirty || isDirty {
		pf.isDirty = true
		return nil
	}
	// Looks like we can unset the dirty bit.
	pf.isDirty = false
	return nil
}

// flushPage flushes the target page to the storage manager
func (b *pageManager) flushPage(pid pageID) error {
	// First we attempt to locate the requested page pageFrame in the
	// page table using the supplied pageID.
	fid, found := b.table[pid]
	if !found {
		// We have not found it, return an error.
		return ErrPageNotFound
	}
	// Otherwise, we have located the matching frameID for the page
	// we are looking for. Next we should get the actual page pageFrame
	// from our page pageFrame set.
	pf := b.frames[fid]
	// Decrement the pin count and flush it.
	pf.decrPinCount()
	err := b.manager.write(pf.pid, pf.page)
	if err != nil {
		return err
	}
	// Finally, since we just flushed it, we can unset the dirty bit.
	pf.isDirty = false
	return nil
}

// newPage allocates a new page in the pageFrameManager requesting it
// from the storage manager
func (b *pageManager) newPage() *pageFrame {
	// First we need a frameID we can use to proceed. We will call
	// getFrame which will first check our free list and if we do
	// not find one in there getFrame will victimize a pageFrame and
	// return a frameID we can use.
	fid, err := b.getFrame()
	if err != nil {
		// Something went wrong!
		// panic(err)
		return nil
	}
	// Now that we have an empty pageFrame in the table to load a pageFrame
	// into, we will ask the storage manager to allocate a new page
	// and return the pageID, so we can proceed.
	pid := b.manager.allocate()
	// Next, we should create a page pageFrame utilizing the new pageID.
	// We can also add it to our page table, and to our active pageFrame
	// set. *It should be noted that this newly allocated page will
	// not be persisted with the storage manager unless this page is
	// flushed at some point, and if this page is victimized before
	// it is flushed it will be lost.
	pf := newFrame(pid)
	b.table[pf.pid] = *fid
	b.frames[*fid] = pf
	// Finally, return our page pageFrame.
	return pf
}

// deletePage deletes a page from the pageFrameManager
func (b *pageManager) deletePage(pid pageID) error {
	// First, we should check to see if this page pageFrame is in the
	// page table or not. If it is not, then we will not need to
	// do anything.
	fid, found := b.table[pid]
	if !found {
		// Not in the page table, so we don't need to do anything.
		return nil
	}
	// If we get here, then we found it in the page table. So first
	// we should get the actual page pageFrame, and then check to see
	// if it is pinned.
	pf := b.frames[fid]
	if pf.pinCount > 0 {
		// Page must be in use (or has not properly been unpinned.)
		return ErrPageInUse
	}
	// If we are here, we have our page pageFrame, and it is not currently
	// in use. First remove it from the page table.
	delete(b.table, pf.pid)
	// Next, we will pin this pageFrame. Not exactly sure why we should
	// pin it. Maybe so this page cannot be used again? Or perhaps
	// it is useful later on if we pull this pageFrame from the free list
	// for some reason? Regardless, this is what the algorithm that I
	// am referring to instructs us to do, so we will do it.
	b.replacer.pin(fid)
	// Then we will instruct the storage manager to deallocate it the
	// page, and finally we will add the pageFrame back into the free list.
	b.manager.deallocate(pid)
	b.addFreeFrame(fid)
	// And lastly, return a nil error.
	return nil
}

func (b *pageManager) addFreeFrame(fid frameID) {
	b.free = append(b.free, fid)
}

func (b *pageManager) getFreeFrame() (*frameID, bool) {
	if len(b.free) > 0 {
		fid, newFree := b.free[0], b.free[1:]
		b.free = newFree
		return &fid, true
	}
	return nil, false
}

// getFrame attempts to fetch a free page pageFrame, otherwise it victimizes one
func (b *pageManager) getFrame() (*frameID, error) {
	// First we should check to see if we have any free page
	// frames in our free list.
	// fid, found := b.free.get()
	fid, found := b.getFreeFrame()
	// If we do not have any available frames in our free
	// list, we have no choice but to victimize one; so let
	// us see what we are dealing with.
	if !found {
		// We did not find a frameID in our free list, so
		// now we will victimize one.
		fid = (*b.replacer).victim()
		// Next we must remove the victimized pageFrame, but
		// in order to do that we also need to ensure it
		// does not need to be flushed.
		vf := b.frames[*fid]
		if vf != nil {
			if vf.isDirty {
				// Our page pageFrame is dirty; write it.
				err := b.manager.write(vf.pid, vf.page)
				if err != nil {
					return nil, err
				}
			}
			// And now we remove it from the table.
			delete(b.table, vf.pid)
		}
	}
	// Now we have an empty pageFrame we can utilize, so we
	// will return a pointer to our frameID, and a nil
	// error.
	return fid, nil
}

// flushAll flushes all the pinned pages to the storage manager
func (b *pageManager) flushAll() error {
	// Very simply, we will just range all the entries in the
	// page pageFrame table, and call flushPage for each of them.
	var err error
	for pid := range b.table {
		err = b.flushPage(pid)
		if err != nil {
			return err
		}
	}
	return nil
}
