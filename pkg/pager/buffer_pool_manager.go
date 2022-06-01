package pager

import (
	"log"
	"sync"
)

// bufferPoolManager is our implementation of a pageFrameManager, which is
// also sometimes called a buffer pool manager in a dbms system.
type bufferPoolManager struct {
	lock     sync.RWMutex
	frames   []*frame                      // list of loaded page frames
	replacer *clockReplacer[frameID, bool] // used to find an unpinned page for replacement
	manager  *tempDiskManager              // underlying storage manager
	free     []frameID                     // used to find a page for replacement
	table    *pageTable                    // used to keep track of pages
}

// newPageManager initializes and returns a new instance of a bufferPoolManager.
func newPageManager(size int, disk *tempDiskManager) *bufferPoolManager {
	bm := &bufferPoolManager{
		frames:   make([]*frame, size, size),
		replacer: newClockReplacer[frameID, bool](size),
		manager:  disk,
		free:     make([]frameID, size),
		table:    newPageTable(size),
	}
	for i := 0; i < size; i++ {
		bm.frames[i] = nil
		bm.free[i] = frameID(i)
	}
	return bm
}

// fetchPage attempts to locate and return a page from the bufferPoolManager.
// It undertakes the following steps in order to accomplish this task. Steps:
//
//	1.0 - Search the page table for the requested page (P).
//		1.1 - If P exists, pin it and return it immediately.
//	 	1.2 - If P does not exist, find a replacement page (R).
//			1.2.0 - First check the free list for R. (always check free list first)
// 			1.2.1 - If R can not be found in the free list, use the replacer.
//	2.0 - If R is dirty, write it back to the disk.
//	3.0 - Delete R from the page table and insert P.
// 	4.0 - Update P's metadata. Read in page from disk, and return a pointer to P.
//
func (b *bufferPoolManager) fetchPage(pid pageID) *page {
	// Check to see if the pageFrame is already loaded in the
	// page pageFrame table.
	if f, found := b.table.getFrame(pid); found {
		// It appears to be, so we should get the matching page.
		pf := b.frames[f.fid]
		// Don't forget to increment the pin count, and also mark it
		// pinned in the replacer.
		pf.pinCount++
		b.replacer.pin(f.fid)
		// Finally, return the page.
		return &pf.page
	}
	// We did not find the page in the page pageTable. So now we must
	// ensure that we have an empty frame to load the page into once
	// we locate it. So, we will do this...

	// First we need an empty frame we can use to proceed. We will call
	// getUsableFrame which will:
	// 1) Check our free list first. and if we do
	// 2) If
	// not find one in there getUsableFrame will victimize a pageFrame and
	// return a frameID we can use.
	f, err := b.getUsableFrame()
	if err != nil {
		// Something went wrong!
		panic(err)
	}
	// Now that we have an empty pageFrame in the table to load a pageFrame
	// into, we will make an attempt to locate the requested page
	// using the storage manager and read it into a new page pageFrame.
	err = b.manager.read(pid, b.frames[f.pid].page)
	if err != nil {
		// Something went wrong!
		panic(err)
	}
	// Looks like we found it and read it successfully. Now we can
	// add it to our page table.
	b.table.addFrame(f)
	// Finally, add it to our active pageFrame set.
	// b.frames[f.fid] = pf
	// And return our page pageFrame.
	return &(b.frames)[f.pid].page
}

// unpinPage unpins the target page pageFrame from the pageFrameManager
func (b *bufferPoolManager) unpinPage(pid pageID, isDirty bool) error {
	// First we attempt to locate the requested page pageFrame in the
	// page table using the supplied pageID.
	f, found := b.table.getFrame(pid)
	if !found {
		// We have not found it, return an error.
		return ErrPageNotFound
	}
	// Otherwise, we have located the matching frameID for the page
	// we are looking for. Next we should get the actual page pageFrame
	// from our page pageFrame set.
	// p := b.frames[f.fid]
	// Decrement the pin count and check if we are able to unpin it.
	f.decrPinCount()
	if f.pinCount == 0 {
		b.replacer.unpin(f.fid, true)
	}
	// Next, we must check the page pageFrame to see if the dirty bit
	// needs to be set.
	if f.isDirty || isDirty {
		f.isDirty = true
		return nil
	}
	// Looks like we can unset the dirty bit.
	f.isDirty = false
	return nil
}

// flushPage flushes the target page to the storage manager
func (b *bufferPoolManager) flushPage(pid pageID) error {
	// First we attempt to locate the requested page pageFrame in the
	// page table using the supplied pageID.
	f, found := b.table.getFrame(pid)
	if !found {
		// We have not found it, return an error.
		return ErrPageNotFound
	}
	// Otherwise, we have located the matching frameID for the page
	// we are looking for. Next we should get the actual page pageFrame
	// from our page pageFrame set.
	// pf := b.frames[fid]
	// Decrement the pin count and flush it.
	f.decrPinCount()
	err := b.manager.write(f.pid, b.frames[f.fid].page)
	if err != nil {
		return err
	}
	// Finally, since we just flushed it, we can unset the dirty bit.
	f.isDirty = false
	return nil
}

// newPage attempts to allocate and return a new page in the bufferPoolManager.
//
// It undertakes the following steps in order to accomplish this task. Steps:
// 	1.0 - If the buffer pool is full and all pages are pinned return nil.
// 	2.0 - Pick a victim page P
// 		2.1 - First look in the free list for P
// 		2.2 - If P cannot be found in the free list, use the replacer.
// 	3.0 - Update P's metadata. Zero out memory ad add P to the page table.
// 	4.0 - Return a pointer to P.
//
func (b *bufferPoolManager) newPage() *page {
	// First we need a frameID we can use to proceed. We will call
	// getUsableFrame which will first check our free list and if we do
	// not find one in there getUsableFrame will victimize a pageFrame and
	// return a frameID we can use.
	f, err := b.getUsableFrame()
	if err != nil {
		// Something went wrong!
		// panic(err)
		log.Println("could not get a usable frame for some reason!")
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
	// pf := newFrame(pid)
	f.pid = pid
	b.table.addFrame(f)
	// b.table[pf.pid] = *fid
	p := newPage(uint32(pid))
	b.frames[f.fid].page = p
	// Finally, return our page pageFrame.
	return &f.page
}

// deletePage deletes a page from the pageFrameManager
func (b *bufferPoolManager) deletePage(pid pageID) error {
	// First, we should check to see if this page pageFrame is in the
	// page table or not. If it is not, then we will not need to
	// do anything.
	f, found := b.table.getFrame(pid)
	if !found {
		// Not in the page table, so we don't need to do anything.
		return nil
	}
	// If we get here, then we found it in the page table. So first
	// we should get the actual page pageFrame, and then check to see
	// if it is pinned.
	// p := b.frames[f.fid]
	if f.pinCount > 0 {
		// Page must be in use (or has not properly been unpinned.)
		return ErrPageInUse
	}
	// If we are here, we have our page pageFrame, and it is not currently
	// in use. First remove it from the page table.
	b.table.delFrame(f.pid)
	// delete(b.table, pf.pid)
	// Next, we will pin this pageFrame. Not exactly sure why we should
	// pin it. Maybe so this page cannot be used again? Or perhaps
	// it is useful later on if we pull this pageFrame from the free list
	// for some reason? Regardless, this is what the algorithm that I
	// am referring to instructs us to do, so we will do it.
	b.replacer.pin(f.fid)
	// Then we will instruct the storage manager to deallocate it the
	// page, and finally we will add the pageFrame back into the free list.
	b.manager.deallocate(pid)
	b.addFreeFrame(f.fid)
	// And lastly, return a nil error.
	return nil
}

func (b *bufferPoolManager) addFreeFrame(f frameID) {
	b.free = append(b.free, f)
}

func (b *bufferPoolManager) getFreeFrame() (*frameID, bool) {
	if len(b.free) > 0 {
		f, newFree := b.free[0], b.free[1:]
		b.free = newFree
		return &f, true
	}
	return nil, false
}

// getUsableFrame attempts to return a usable frame. It is used in the event that
// the buffer pool is full. It always checks the free list first, and in the
// event that one cannot be obtained from our free list, the replacer is
// called upon to locate the correct frame. If a frame is found, and indicating
// that the page is dirty, it flushes the page to disk before victimizing and
// returning the frame.
func (b *bufferPoolManager) getUsableFrame() (*frame, error) {
	// First we should check to see if we have any free frames in our free list.
	// fid, found := b.free.get()
	f, found := b.getFreeFrame()
	// If we do not have any available frames in our free
	// list, we have no choice but to victimize one; so let
	// us see what we are dealing with.
	if !found {
		// We did not find a free frame in our free list, so now we victimize one.
		*f, _ = (*b.replacer).victim()
		// Next we must remove the victimized pageFrame, but
		// in order to do that we also need to ensure it
		// does not need to be flushed.
		pf := b.frames[*f]
		if pf != nil {
			if pf.isDirty {
				// Our page pageFrame is dirty; write it.
				err := b.manager.write(pf.pid, pf.page)
				if err != nil {
					return nil, err
				}
			}
			// And now we remove it from the table.
			b.table.delFrame(pf.pid)
			// delete(b.table, f.pid)
		}
	}
	// Now we have an empty pageFrame we can utilize, so we
	// will return a pointer to our frameID, and a nil
	// error.
	return b.frames[*f], nil
}

// flushAll flushes all the pinned pages to the storage manager
func (b *bufferPoolManager) flushAll() error {
	// Very simply, we will just range all the entries in the
	// page pageFrame table, and call flushPage for each of them.
	var err error
	for pid, _ := range b.table.entries {
		err = b.flushPage(pid)
		if err != nil {
			return err
		}
	}
	return nil
}
