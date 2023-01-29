package pager

import (
	"fmt"
	"sync"
)

/*
	Notes:
			Minimum buffer size for
			each of the page sizes
			with the minimum page
			count in the buffer set
			to 64 pages.
		+-----------------+-----------+
		| MIN_BUFFER_SIZE | PAGE_SIZE |
		+-----------------|-----------+
		|       64KB      |     1KB   |
		+-----------------|-----------+
		|      128KB      |     2KB   |
		+-----------------|-----------+
		|      256KB      |     4KB   |
		+-----------------|-----------+
		|      512KB      |     8KB   |
		+-----------------|-----------+
		|       1MB       |    16KB   |
		+-----------------|-----------+
		|       2MB       |    32KB   |
		+-----------------|-----------+
		|       4MB       |    64KB   |
		+-----------------|-----------+
*/

// Defaults for buffer pool
const (
	DefaultBufferSize uint32 = 1 << 20  // 1MB
	MinBufferSize     uint32 = 64 << 10 // 64KB
	MaxBufferSize     uint32 = 8 << 20  // 8MB

	KB = 1 << 10
	MB = 1 << 20

	Size1KB  = 1 << 10
	Size2KB  = 2 << 10
	Size4KB  = 4 << 10
	Size8KB  = 8 << 10
	Size16KB = 16 << 10
	Size32KB = 32 << 10
	Size64KB = 64 << 10
	Size1MB  = 1 << 20
	Size2MB  = 1 << 20
	Size4MB  = 1 << 20
	Size8MB  = 1 << 20
	Size16MB = 1 << 20
)

// bufferPool is an implementation of a page buffer pool, which is also
// sometimes called a buffer pool page disk in a dbms system.
type bufferPool struct {
	bpLatch  sync.Mutex         // latch
	frames   []frame            // list of loaded page frames
	replacer *clockReplacer     // page replacement structure (used to find an unpinned page for replacement)
	disk     *diskManager       // underlying disk storage manager
	free     []frameID          // used to find a page for replacement
	table    map[pageID]frameID // used to keep track of pages
	pageSize uint16             // page size for this buffer pool
}

// newBufferPool initializes and returns a new instance of a bufferPool.
func newBufferPool(file string, pageSize, pageCount uint16) *bufferPool {
	// sanitize the page size
	pageSize = calcPageSize(pageSize)
	dm, err := newDiskManagerSize(file, pageSize, pageCount)
	if err != nil {
		panic(ErrOpeningDiskManager)
	}
	bm := &bufferPool{
		frames:   make([]frame, pageCount, pageCount),
		replacer: newClockReplacer(pageCount),
		disk:     dm,
		free:     make([]frameID, pageCount),
		table:    make(map[pageID]frameID),
		pageSize: pageSize,
	}
	for i := uint16(0); i < pageCount; i++ {
		bm.frames[i] = frame{
			pid:      0,
			fid:      0,
			pinCount: 0,
			isDirty:  false,
			page:     nil,
		}
		bm.free[i] = frameID(i)
	}
	return bm
}

// calcPageSize ensures the provided page size is a power
// of two according to the default page size. If the provided
// page is less than the minimum, then the page size is set
// to the minimum allowable size. If the provided page size
// is set greater than the maximum, then the page size is set
// to the maximum allowable size.
func calcPageSize(pageSize uint16) uint16 {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if 0 < pageSize && pageSize < MinPageSize {
		pageSize = MinPageSize
	}
	if MaxPageSize < pageSize {
		pageSize = MaxPageSize
	}
	if pageSize&(DefaultPageSize-1) != 0 {
		// must be a power of two
		var x uint16 = 1
		for x < pageSize {
			x *= 2
		}
		pageSize = x
	}
	return pageSize
}

// calcBufferSize ensures the provided buffer size is a power
// of two that works well with the minimum and maximum allowable
// buffer sizes. It also ensures that the page size works well
// with the provided buffer size--otherwise, it will be adjusted.
func calcBufferSize(pageSize uint16, bufferSize uint32) uint32 {
	pageSize = calcPageSize(pageSize)
	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}
	if 0 < bufferSize && bufferSize < MinBufferSize {
		bufferSize = MinBufferSize
	}
	if MaxBufferSize < bufferSize {
		bufferSize = MaxBufferSize
	}
	if bufferSize&(DefaultBufferSize-1) != 0 {
		// must be a power of two
		var x uint32 = 1
		for x < bufferSize {
			x *= 2
		}
		bufferSize = x
	}
	return bufferSize
}

// newPage attempts to allocate and return a new page in the bufferPool.
//
// It undertakes the following steps in order to accomplish this task. Steps:
//
//	1.0 - If the buffer pool is full and all pages are pinned return nil.
//	2.0 - Pick a victim page P
//		2.1 - First look in the free list for P
//		2.2 - If P cannot be found in the free list, use the replacer.
//	3.0 - Update P's metadata. Zero out memory ad add P to the page table.
//	4.0 - Return a pointer to P.
func (b *bufferPool) newPage() page {
	// First we need a frameID we can use to proceed. We will call
	// getUsableFrame which will first check our free list and if we do
	// not find one in there getUsableFrame will victimize a pageFrame and
	// return a frameID we can use.
	fid, err := b.getUsableFrame()
	if err != nil {
		// Something went wrong!
		return nil
	}
	// Now that we have an empty pageFrame in the table to load a pageFrame
	// into, we will ask the storage storageManager to allocate a new page
	// and return the pageID, so we can proceed.
	pid := b.disk.allocatePage()
	// Next, we should create a page frame utilizing the new pageID.
	pf := newFrame(pid, *fid, b.pageSize)
	pg := newPageSize(pid, b.pageSize)
	copy(pf.page, pg)
	// Add the entry to our page table.
	b.table[pid] = *fid
	// And update the page frame
	b.frames[*fid] = pf
	// Finally, return our page pageFrame.
	return pf.page
}

// unpinPage unpins the target page pageFrame from the pageFrameManager
func (b *bufferPool) unpinPage(pid pageID, isDirty bool) error {
	// First we attempt to locate the requested page pageFrame in the page table using
	// the supplied pageID.
	fid, found := b.table[pid]
	if !found {
		// We have not found it, return an error.
		return ErrPageNotFound
	}
	// We have located the matching frameID for the page we are looking for. Next, grab
	// the actual page frame from our set.
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

// fetchPage attempts to locate and return a page from the bufferPool.
// It undertakes the following steps in order to accomplish this task. Steps:
//
//	1.0 - Search the page table for the requested page (P).
//		1.1 - If P exists, pin it and return it immediately.
//	 	1.2 - If P does not exist, find a replacement page (R).
//			1.2.0 - First check the free list for R. (always check free list first)
//			1.2.1 - If R can not be found in the free list, use the replacer.
//	2.0 - If R is dirty, write it back to the disk.
//	3.0 - Delete R from the page table and insert P.
//	4.0 - Update P's metadata. Read in page from disk, and return a pointer to P.
func (b *bufferPool) fetchPage(pid pageID) page {
	// Check to see if the frame ID is in the page table.
	if fid, found := b.table[pid]; found {
		// It appears to be, so we should get the matching page.
		pf := b.frames[fid]
		// Don't forget to increment the pin count, and also pin it (make it
		// unusable as a victim) in the replacer.
		pf.pinCount++
		b.replacer.pin(fid)
		// Finally, return the page.
		return pf.page
	}
	// We did not find the entry in the page table, so now we must get a usable
	// frame ID from our free list, or the replacer.
	fid, err := b.getUsableFrame()
	if err != nil {
		// Something went wrong!
		// log.Printf("ERROR FETCHING PAGE, GETTING UN-USABLE FRAME: %s\n", err)
		return nil
	}
	// Read in the page data using the disk storageManager.
	data := make([]byte, b.pageSize)
	err = b.disk.readPage(pid, data)
	if err != nil {
		// Something went wrong!
		// log.Printf("FETCHING PAGE, READING DATA OFF THE DISK: %s\n", err)
		return nil
	}
	// Create a new frame instance and copy the data we read into the frame page, because
	// there is not currently an instance of this page frame in the page table since we had
	// to victimize one.
	pf := newFrame(pid, *fid, b.pageSize)
	copy(pf.page, data)
	// Add the entry to our page table.
	b.table[pid] = *fid
	// Update our frame with the new page
	b.frames[*fid] = pf
	// Fill out any other metadata...
	//
	// Lastly, we return our page.
	return pf.page
}

// flushPage flushes the target page to the storage storageManager
func (b *bufferPool) flushPage(pid pageID) error {
	// First we attempt to locate the matching frame ID using our page table.
	fid, found := b.table[pid]
	if !found {
		// We have not found it, return an error.
		return ErrPageNotFound
	}
	// Get our the page frame from our frame set using our frame ID.
	pf := b.frames[fid]
	// Decrement the pin count and flush it.
	pf.decrPinCount()
	err := b.disk.writePage(pf.pid, pf.page)
	if err != nil {
		return err
	}
	// Finally, since we just flushed it, we can unset the dirty bit.
	pf.isDirty = false
	return nil
}

// deletePage deletes a page from the pageFrameManager
func (b *bufferPool) deletePage(pid pageID) error {
	// First, we should check to see if this page frame is in the page table.
	// If it is not, then we will not need to do anything.
	fid, found := b.table[pid]
	if !found {
		// Not in the page table, so we don't need to do anything.
		return nil
	}
	// If we get here, then we found it in the page table. So first we should
	// get the actual page frame, and then check to see if it is pinned.
	pf := b.frames[fid]
	if pf.pinCount > 0 {
		// Page must be in use (or has not properly been unpinned.)
		return ErrPageInUse
	}
	// If we are here, we have our page pageFrame, and it is not currently in use.
	// First remove it from the page table.
	delete(b.table, pid)
	// Next, we will pin this page frame, so it cannot be returned as a victimized
	// frame (because we are trying to delete it.)
	b.replacer.pin(fid)
	// Then we will instruct the storage storageManager to deallocate it the page, and
	// finally we will add the pageFrame back into the free list.
	if err := b.disk.deallocatePage(pid); err != nil {
		// Oops, something happened on the disk.
		return err
	}
	b.addFreeFrame(fid)
	// And lastly, return a nil error.
	return nil
}

// addFreeFrame takes a frameID and adds it to the set of free frames.
func (b *bufferPool) addFreeFrame(f frameID) {
	b.free = append(b.free, f)
}

// getFrameID attempts to return a frameID. It first checks the free list
// to see if there are any free frames in there. If none are found, it will
// then go on to the replacer in search of one.
func (b *bufferPool) getFrameID() (*frameID, bool) {
	// check free list first
	if len(b.free) > 0 {
		// return one from our free list
		fid, newFree := b.free[0], b.free[1:]
		b.free = newFree
		return &fid, true // bool: fromFreeList == true
	}
	// nothing for us in the free list, time to use the replacer
	return b.replacer.victim(), false // bool: fromFreeList == false
}

// getUsableFrame attempts to return a usable frameID. It is used in the event that
// the buffer pool is "full." It always checks the free list first, and then it will
// fall back to using the replacer. In either case, if a frameID is not found an error
// of type ErrUsableFrameNotFound. Otherwise, a frameID is located and returned along
// with a nil error.
func (b *bufferPool) getUsableFrame() (*frameID, error) {
	// First check the free list, then the replacer. Try to obtain a usable frameID.
	fid, foundInFreeList := b.getFrameID()
	if fid == nil {
		// Make sure nothing weird happened.
		return nil, ErrUsableFrameNotFound
	}
	// If we are here, we have a frameID, but we don't yet know if it came out of our
	// free list, or from the allocator. If it is from the allocator then we need to
	// check the frame and possible flush it to disk before using the frame; so let's
	// see if it was returned using the allocator and go from there.
	if !foundInFreeList {
		// Get the current frame.
		cf := b.frames[*fid]
		// Check and handle empty or nil frame.
		if &cf != nil {
			if cf.isDirty {
				// Our page pageFrame is dirty; write it.
				err := b.disk.writePage(cf.pid, cf.page)
				if err != nil {
					return nil, err // error on write
				}
			}
			// And either way, in the event a frame from the free list has not been
			// returned, we must remove this frame entry from the page table because
			// it is no longer valid, and we will be creating a new entry soon.
			delete(b.table, cf.pid)
		}
	}
	// Now we return a pointer to our frameID, and a nil error.
	return fid, nil
}

// flushAll flushes all the pinned pages to the storage storageManager
func (b *bufferPool) flushAll() error {
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

// close attempts to close the buffer pool
func (b *bufferPool) close() error {
	err := b.flushAll()
	if err != nil {
		return err
	}
	err = b.disk.close()
	if err != nil {
		return err
	}
	return nil
}

// String is the string method for this type
func (b *bufferPool) String() string {
	ss := fmt.Sprintf("Buffer Pool Manager\n")
	ss += fmt.Sprintf("\tframes:\n")
	for i := range b.frames {
		ss += fmt.Sprintf("\t\tframe %d = %v\n", i, b.frames[i])
	}
	ss += fmt.Sprintf("\tfree:\n")
	ss += fmt.Sprintf("\t\tfree frames = %v\n", b.free)
	ss += fmt.Sprintf("\ttable(pid,fid):\n")
	if len(b.table) < 1 {
		ss += fmt.Sprintf("\t\tnil\n")
	} else {
		for pid, fid := range b.table {
			ss += fmt.Sprintf("\t\t%d -> %d\n", pid, fid)
		}
	}
	ss += fmt.Sprintf("\tclock replacer:\n")
	ss += fmt.Sprintf("\t\t%v\n", &b.replacer)
	return ss
}
