package buffer

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/cagnosolutions/go-data/pkg/engine/logging"
	"github.com/cagnosolutions/go-data/pkg/engine/page"
	"github.com/cagnosolutions/go-data/pkg/engine/storage"
)

// BufferPoolManager is the access level structure wrapping up the bufferPool, and DiskStore,
// along with a page table, and replacement policy.
type BufferPoolManager struct {
	latch     sync.Mutex
	pool      []Frame                 // buffer pool page frames
	replacer  Replacer                // page replacement policy structure
	store     storage.Storer          // underlying current manager
	freeList  []FrameID               // list of frames that are free to use
	pageTable map[page.PageID]FrameID // table of the current page to frame mappings

	hits   uint64 // number of times a page was found in the buffer pool
	misses uint64 // number of times a page was not found (had to be paged in)
}

// New initializes and returns a new buffer cache manager instance.
func New(storage storage.Storer, size uint16) (*BufferPoolManager, error) {
	// Create a new replacer instance
	replacer := NewClockReplacer(size)
	// Create buffer manager instance
	bm := &BufferPoolManager{
		pool:      make([]Frame, size, size),
		replacer:  replacer,
		store:     storage,
		freeList:  make([]FrameID, size),
		pageTable: make(map[page.PageID]FrameID),
	}
	// initialize the pool in the buffer manager
	for i := uint16(0); i < size; i++ {
		bm.pool[i] = Frame{
			pid:      0,
			fid:      0,
			pinCount: 0,
			isDirty:  false,
			Page:     nil,
		}
		bm.freeList[i] = FrameID(i)
	}
	// return buffer manager
	return bm, nil
}

// AllocatePage simply returns the next sequential page id, but it does not
// initialize, return, or allocate any pages on the disk.
func (bm *BufferPoolManager) AllocatePage() page.PageID {
	// latch
	bm.latch.Lock()
	defer bm.latch.Unlock()
	// Call allocate
	return bm.store.AllocatePage()
}

// NewPage returns a fresh empty page from the pool.
// TODO: consider returning an error from here in the case where the pool is full
func (bm *BufferPoolManager) NewPage() (page.Page, error) {
	// latch
	bm.latch.Lock()
	defer bm.latch.Unlock()
	// First we must acquire a Frame in order to store our page. Calling
	// GetUsableFrame first checks our freeList and if we cannot find one in there
	// our replacement policy is used to locate a victimized one Frame.
	fid, err := bm.getUsableFrameID()
	if err != nil {
		// This can happen when the BufferPoolManager is full, so let's make sure that
		// it's something like that, and not something more sinister.
		// if len(bm.freeList) == 0 && bm.replacer.Size() == 0 {
		// 	return nil
		// }
		if err != ErrUsableFrameNotFound {
			// Something went terribly wrong if this happens.
			logging.DefaultLogger.Panic("new page: %s", err)
		}
		// Nope, it's something more sinister... shoot.
		// logging.DefaultLogger.Panic("{!!!} %s", err)
		return nil, err
	}
	// Allocate (get the next sequential PageID) so we can use it to initialize
	// the next page we will use.
	pid := bm.store.AllocatePage()
	// Create a new Frame initialized with our PageID and Page.
	pf := newFrame(pid, *fid, page.PageSize)
	pg := page.NewPage(pid, page.P_USED)
	copy(pf.Page, pg)
	// Add an entry to our pageTable
	bm.pageTable[pid] = *fid
	// And update the pool
	bm.pool[*fid] = pf
	// Finally, return our Page for use
	return pf.Page, nil
}

// FetchPage retrieves specific page from the pool, or storage medium by the page ID.
func (bm *BufferPoolManager) FetchPage(pid page.PageID) (page.Page, error) {
	// latch
	bm.latch.Lock()
	defer bm.latch.Unlock()
	// Check to see if the PageID is located in the pageTable.
	if fid, found := bm.pageTable[pid]; found {
		// We located it, so now we access the Frame and ensure that it will
		// not be a victim candidate by our replacement policy.
		pf := bm.pool[fid]
		pf.pinCount++
		bm.replacer.Pin(fid)
		// We have a page hit, so we can increase our hit counter
		bm.hits++
		// And now, we can safely return our Page.
		return pf.Page, nil
	}
	// A match was not found in our pageTable, so now we must swap the Page in
	// from disk. But first, we must get a Frame to hold our Page. We will
	// call on GetUsableFrame to check our freeList, and then potentially move on to
	// return a victimized Frame if we need to.
	fid, err := bm.getUsableFrameID()
	if err != nil {
		// TODO: think about a more graceful way of handling this whole situation
		// Check the EXACT error
		if err != ErrUsableFrameNotFound {
			// Something went terribly wrong if this happens.
			logging.DefaultLogger.Panic("fetch page: %s", err)
		}
		return nil, err
	}
	// Now, we will swap the Page in from the disk using the DiskStore.
	data := make([]byte, page.PageSize)
	err = bm.store.ReadPage(pid, data)
	if err != nil {
		// Something went terribly wrong if this happens.
		logging.DefaultLogger.Panic("%s", err)
	}
	// Create a new frame, so we can copy the page data we just swapped
	// in from off the disk and add the Frame to the pageTable.
	pf := newFrame(pid, *fid, page.PageSize)
	copy(pf.Page, data)
	// Add the entry to our pageTable
	bm.pageTable[pid] = *fid
	// And update the pool
	bm.pool[*fid] = pf
	// We had to swap a page in, so we can update our page miss counter
	bm.misses++
	// Finally, return our Page for use
	return pf.Page, nil
}

// UnpinPage allows for manual unpinning of a specific page from the pool by the page ID.
func (bm *BufferPoolManager) UnpinPage(pid page.PageID, isDirty bool) error {
	// latch
	bm.latch.Lock()
	defer bm.latch.Unlock()
	// Check to see if the PageID is located in the pageTable.
	fid, found := bm.pageTable[pid]
	if !found {
		// We have not located it, we will return an error.
		return page.ErrPageNotFound
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// ensure that it can be used as a victim candidate by our replacement policy.
	pf := bm.pool[fid]
	pf.decrPinCount()
	if pf.pinCount <= 0 {
		// After we decrement the pin count, check to see if it is low enough to
		// completely unpin it, and if so, unpin it.
		bm.replacer.Unpin(fid)
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

// FlushPage forces a page to be written onto the storage medium, and decrements the
// pin count on the frame potentially enabling the frame to be reused.
func (bm *BufferPoolManager) FlushPage(pid page.PageID) error {
	// latch
	bm.latch.Lock()
	defer bm.latch.Unlock()
	// Check to see if the PageID is located in the pageTable.
	fid, found := bm.pageTable[pid]
	if !found {
		// We have not located it, we will return an error.
		return page.ErrPageNotFound
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// ensure that it can be used as a victim candidate by our replacement policy.
	pf := bm.pool[fid]
	pf.decrPinCount()
	// Now, we can make sure we flush it to the disk using the DiskStore.
	err := bm.store.WritePage(pf.pid, pf.Page)
	if err != nil {
		// Something went terribly wrong if this happens.
		logging.DefaultLogger.Panic("%s", err)
	}
	// Finally, since we have just flushed the Page to the underlying current, we
	// can proceed with unsetting the dirty bit.
	pf.isDirty = false
	return nil
}

// DeletePage removes the page from the buffer pool, and decrements the pin count on the
// frame potentially enabling the frame to be reused.
func (bm *BufferPoolManager) DeletePage(pid page.PageID) error {
	// latch
	bm.latch.Lock()
	defer bm.latch.Unlock()
	// Check to see if the PageID is located in the pageTable.
	fid, found := bm.pageTable[pid]
	if !found {
		// We have not located it, but we don't need to return any error
		return nil
	}
	// Otherwise, we located it in the pageTable. Now we access the Frame and
	// check to see if it is currently pinned (indicating it is currently in use
	// elsewhere) and should therefore not be removed just yet.
	pf := bm.pool[fid]
	if pf.pinCount > 0 {
		// Page must be in use elsewhere (or has otherwise not been properly
		// unpinned) so we'll return an error for now.
		return page.ErrPageInUse
	}
	// Now, we have our frame, and it is not currently in use, so first we will
	// remove it from the pageTable
	delete(bm.pageTable, pid)
	// Next, we pin it, so it will not be marked as a potential victim--because we
	// are in the process of remove it altogether.
	bm.replacer.Pin(fid)
	// After it is pinned, we will deallocate the Page on disk (which will make
	// it free to use again in a pinch.)
	if err := bm.store.DeallocatePage(pid); err != nil {
		// Ops, something went down disk side, return error
		return err
	}
	// Finally, add the current FrameID back onto the free list and return nil
	bm.addFrameID(fid)
	return nil
}

// getFrameID attempts to return a *FrameID. It first checks the freeList set to
// see if there are any available frames to pick from. If not, it will proceed to use
// the replacement policy to locate one.
func (bm *BufferPoolManager) getFrameID() (*FrameID, bool) {
	// Check the freeList first, and if it is not empty return one
	if len(bm.freeList) > 0 {
		fid, newFreeList := bm.freeList[0], bm.freeList[1:]
		bm.freeList = newFreeList
		return &fid, true // true == fromFreeList
	}
	// Otherwise, there is nothing for us in the free list, so it's time to use our
	// replacement policy
	return bm.replacer.Victim(), false
}

// addFrameID takes a FrameID and adds it back onto our freeList for later use.
func (bm *BufferPoolManager) addFrameID(fid FrameID) {
	bm.freeList = append(bm.freeList, fid)
}

// getUsableFrameID attempts to return a usable FrameID. It is used in the event
// that the buffer pool is "full." It always checks the free list first, and then it
// will fall back to using the replacer.
func (bm *BufferPoolManager) getUsableFrameID() (*FrameID, error) {
	// First, we will check our freeList.
	fid, foundInFreeList := bm.getFrameID()
	if fid == nil {
		return nil, ErrUsableFrameNotFound
	}
	// If we find ourselves here we have a FrameID, but we do not yet know if
	// it came from our free list, or from our replacement policy. If it came due to
	// our replacement policy, then we will need to check to make sure it has not
	// been marked dirty, otherwise we must flush the contents to disk before reusing
	// the FrameID; so let us check on that.
	if !foundInFreeList {
		cf := bm.pool[*fid]
		if &cf != nil {
			// We've located the correct Frame in the pool.
			if cf.isDirty {
				// And it appears that it is in fact holding a dirty Page. We must
				// flush the dirty Page to disk before recycling this Frame.
				err := bm.store.WritePage(cf.pid, cf.Page)
				if err != nil {
					return nil, err
				}
			}
			// In either case, we will now be able to remove this pageTable mapping
			// because it is no longer valid, and the caller should be creating a new
			// entry soon regardless.
			delete(bm.pageTable, cf.pid)
		}
	}
	// Finally, return our *FrameID, and a nil error
	return fid, nil
}

// FlushAll attempts to flush any dirty page data.
func (bm *BufferPoolManager) FlushAll() error {
	// We will range all the entries in the pageTable and call Flush on each one.
	for pid := range bm.pageTable {
		err := bm.FlushPage(pid)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close attempts to Close the BufferPoolManager along with the underlying DiskStore
// and associated dependencies. Close makes sure to flush any dirty page data
// before closing everything down.
func (bm *BufferPoolManager) Close() error {
	// Make sure all dirty Page data is written
	err := bm.FlushAll()
	if err != nil {
		return err
	}
	// close the DiskStore
	err = bm.store.Close()
	if err != nil {
		return err
	}
	return nil
}

func (bm *BufferPoolManager) monitor() {
	for {
		time.Sleep(500 * time.Millisecond)
		// Check to see if the hit rate is below 80%
		bm.latch.Lock()
		hitRate := float64(bm.hits) / float64(bm.misses)
		bm.latch.Unlock()
		if hitRate < 0.8 {
			// We should consider increasing the pool size
			log.Printf("page cache: hit rate is at %f%%, consider increasing pool size.\n", hitRate)
		}
	}
}

func (bm *BufferPoolManager) JSON() string {
	info := struct {
		UsedFrames     []Frame                 `json:"used_frames"`
		FreeFrames     []FrameID               `json:"free_frames"`
		StorageManager storage.Storer          `json:"disk_manager"`
		PageTable      map[page.PageID]FrameID `json:"page_table"`
		Hits           uint64                  `json:"hits"`
		Misses         uint64                  `json:"misses"`
	}{
		UsedFrames:     bm.pool,
		FreeFrames:     bm.freeList,
		StorageManager: bm.store,
		PageTable:      bm.pageTable,
		Hits:           bm.hits,
		Misses:         bm.misses,
	}

	b, err := json.MarshalIndent(&info, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (bm *BufferPoolManager) String() string {
	// var sb strings.Builder
	//
	// // Frame info
	// sb.WriteString(heading("Frame info"))
	// var used int
	// var notused int
	// for i := range bm.pool {
	// 	if bm.pool[i].pid > 0 && bm.pool[i].pinCount > 0 || len(bm.pool[i].Page) > 0 {
	// 		used++
	// 	} else {
	// 		notused++
	// 	}
	// }
	// sb.WriteString("used: ")
	// sb.WriteString(strconv.Itoa(used))
	// sb.WriteString("\n")
	// sb.WriteString("free: ")
	// sb.WriteString(strconv.Itoa(notused))
	// sb.WriteString("\n")
	//
	// // Hits and misses info
	// sb.WriteString(heading("Cache info"))
	// sb.WriteString("hits: ")
	// sb.WriteString(strconv.Itoa(int(bm.hits)))
	// sb.WriteString("\n")
	// sb.WriteString("misses: ")
	// sb.WriteString(strconv.Itoa(int(bm.misses)))
	// sb.WriteString("\n")
	//
	// return sb.String()
	return bm.JSON()
}

func heading(s string) string {
	return "\n" + s + "\n-----\n"
}

// pool      []Frame                 // buffer pool page frames
// replacer  *ClockReplacer          // page replacement policy structure
// disk      *disk.DiskStore       // underlying current manager
// freeList  []FrameID               // list of frames that are free to use
// pageTable map[page.PageID]FrameID // table of the current page to frame mappings
//
// hits   uint64 // number of times a page was found in the buffer pool
// misses uint64 // number of times a page was not found (had to be paged in)
