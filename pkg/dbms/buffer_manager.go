package dbms

import (
	"errors"
	"sync"

	"github.com/cagnosolutions/go-data/pkg/dbms/errs"
	"github.com/cagnosolutions/go-data/pkg/dbms/frame"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
	"github.com/cagnosolutions/go-data/pkg/logging"
)

var out = logging.NewDefaultLogger()

const (
	maxSegmentSize     = 16 << 20
	pagesPerSegment    = maxSegmentSize / page.PageSize
	currentSegment     = "dat-current.seg"
	segmentPrefix      = "dat-"
	segmentSuffix      = ".seg"
	segmentIndexSuffix = ".idx"
	extentSize         = 64 << 10
)

var (
	PageSizes = []uint16{
		4096,  // 4KB
		8192,  // 8KB
		16384, // 16KB
	}
	PageCounts = []uint16{64, 128, 256, 512}
)

var (
	ErrBadPageCount = errors.New("bad page count, must be a multiple of 64 between 64 and 1024")
)

// BufferManager is the access level structure wrapping up the bufferPool, and FileManager,
// along with a page table, and replacement policy.
type BufferManager struct {
	latch     sync.Mutex
	pool      []frame.Frame                 // buffer pool page frames
	replacer  *ClockReplacer                // page replacement policy structure
	disk      *FileManager                  // underlying file manager
	freeList  []frame.FrameID               // list of frames that are free to use
	pageTable map[page.PageID]frame.FrameID // table of the current page to frame mappings
}

// OpenBufferManager opens an existing storage manager instance if one exists with the same namespace
// otherwise it creates a new instance and returns it.
func OpenBufferManager(base string, pageCount uint16) (*BufferManager, error) {
	// validate page count
	if pageCount%64 != 0 || pageCount/64 > 16 {
		return nil, ErrBadPageCount
	}
	// open file manager
	fm, err := OpenFileManager(base)
	if err != nil {
		return nil, err
	}
	// create buffer manager instance
	bm := &BufferManager{
		pool:      make([]frame.Frame, pageCount, pageCount),
		replacer:  NewClockReplacer(pageCount),
		disk:      fm,
		freeList:  make([]frame.FrameID, pageCount),
		pageTable: make(map[page.PageID]frame.FrameID),
	}
	// initialize the pool in the buffer manager
	for i := uint16(0); i < pageCount; i++ {
		bm.pool[i] = frame.Frame{
			PID:      0,
			FID:      0,
			PinCount: 0,
			IsDirty:  false,
			Page:     nil,
		}
		bm.freeList[i] = frame.FrameID(i)
	}
	// return buffer manager
	return bm, nil
}

// NewPage returns a fresh empty page from the pool.
func (m *BufferManager) NewPage() page.Page {
	// First we must acquire a frame.Frame in order to store our page. Calling
	// GetUsableFrame first checks our freeList and if we cannot find one in there
	// our replacement policy is used to locate a victimized one frame.Frame.
	fid, err := m.getUsableFrameID()
	if err != nil {
		// Something went terribly wrong if this happens.
		out.Panic("%s", err)
	}
	// Allocate (get the next sequential page.PageID) so we can use it to initialize
	// the next page we will use.
	pid := m.disk.AllocatePage()
	// Create a new frame.Frame initialized with our page.PageID and page.Page.
	pf := frame.NewFrame(pid, *fid, page.PageSize)
	pg := page.NewPage(pid)
	copy(pf.Page, pg)
	// Add an entry to our pageTable
	m.pageTable[pid] = *fid
	// And update the pool
	m.pool[*fid] = pf
	// Finally, return our page.Page for use
	return pf.Page
}

// FetchPage retrieves specific page from the pool, or storage medium by the page ID.
func (m *BufferManager) FetchPage(pid page.PageID) page.Page {
	// Check to see if the page.PageID is located in the pageTable.
	if fid, found := m.pageTable[pid]; found {
		// We located it, so now we access the frame.Frame and ensure that it will
		// not be a victim candidate by our replacement policy.
		pf := m.pool[fid]
		pf.PinCount++
		m.replacer.Pin(fid)
		// And now, we can safely return our page.Page.
		return pf.Page
	}
	// A match was not found in our pageTable, so now we must swap the page.Page in
	// from disk. But first, we must get a frame.Frame to hold our page.Page. We will
	// call on GetUsableFrame to check our freeList, and then potentially move on to
	// return a victimized frame.Frame if we need to.
	fid, err := m.getUsableFrameID()
	if err != nil {
		// Something went terribly wrong if this happens.
		out.Panic("%s", err)
	}
	// Now, we will swap the page.Page in from the disk using the FileManager.
	data := make([]byte, page.PageSize)
	err = m.disk.ReadPage(pid, data)
	if err != nil {
		// Something went terribly wrong if this happens.
		out.Panic("%s", err)
	}
	// Create a new frame.Frame so we can copy the page.Page data we just swapped
	// in from off the disk and add the frame.Frame to the pageTable.
	pf := frame.NewFrame(pid, *fid, page.PageSize)
	copy(pf.Page, data)
	// Add the entry to our pageTable
	m.pageTable[pid] = *fid
	// And update the pool
	m.pool[*fid] = pf
	// Finally, return our page.Page for use
	return pf.Page
}

// UnpinPage allows for manual unpinning of a specific page from the pool by the page ID.
func (m *BufferManager) UnpinPage(pid page.PageID, isDirty bool) error {
	// Check to see if the page.PageID is located in the pageTable.
	fid, found := m.pageTable[pid]
	if !found {
		// We have not located it, we will return an error.
		return errs.ErrPageNotFound
	}
	// Otherwise, we located it in the pageTable. Now we access the frame.Frame and
	// ensure that it can be used as a victim candidate by our replacement policy.
	pf := m.pool[fid]
	pf.DecrPinCount()
	if pf.PinCount == 0 {
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
func (m *BufferManager) FlushPage(pid page.PageID) error {
	// Check to see if the page.PageID is located in the pageTable.
	fid, found := m.pageTable[pid]
	if !found {
		// We have not located it, we will return an error.
		return errs.ErrPageNotFound
	}
	// Otherwise, we located it in the pageTable. Now we access the frame.Frame and
	// ensure that it can be used as a victim candidate by our replacement policy.
	pf := m.pool[fid]
	pf.DecrPinCount()
	// Now, we can make sure we flush it to the disk using the FileManager.
	err := m.disk.WritePage(pf.PID, pf.Page)
	if err != nil {
		// Something went terribly wrong if this happens.
		out.Panic("%s", err)
	}
	// Finally, since we have just flushed the page.Page to the underlying file, we
	// can proceed with unsetting the dirty bit.
	pf.IsDirty = false
	return nil
}

// DeletePage removes the page from the buffer pool, and decrements the pin count on the
// frame potentially enabling the frame to be reused.
func (m *BufferManager) DeletePage(pid page.PageID) error {
	// Check to see if the page.PageID is located in the pageTable.
	fid, found := m.pageTable[pid]
	if !found {
		// We have not located it, but we don't need to return any error
		return nil
	}
	// Otherwise, we located it in the pageTable. Now we access the frame.Frame and
	// check to see if it is currently pinned (indicating it is currently in use
	// elsewhere) and should therefore not be removed just yet.
	pf := m.pool[fid]
	if pf.PinCount > 0 {
		// page.Page must be in use elsewhere (or has otherwise not been properly
		// unpinned) so we'll return an error for now.
		return errs.ErrPageInUse
	}
	// Now, we have our frame.Frame and it is not currently in use, so first we will
	// remove it from the pageTable
	delete(m.pageTable, pid)
	// Next, we pin it, so it will not be marked as a potential victim--because we
	// are in the process of remove it altogether.
	m.replacer.Pin(fid)
	// After it is pinned, we will deallocate the page.Page on disk (which will make
	// it free to use again in a pinch.)
	if err := m.disk.DeallocatePage(pid); err != nil {
		// Ooops, something went down disk side, return error
		return err
	}
	// Finally, add the current frame.FrameID back onto the free list and return nil
	m.addFrameID(fid)
	return nil
}

// getFrameID attempts to return a *frame.FrameID. It first checks the freeList set to
// see if there are any availble frames to pick from. If not, it will proceed to use
// the replacement policy to locate one.
func (m *BufferManager) getFrameID() (*frame.FrameID, bool) {
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

// addFrameID takes a frame.FrameID and adds it back onto our freeList for later use.
func (m *BufferManager) addFrameID(fid frame.FrameID) {
	m.freeList = append(m.freeList, fid)
}

// getUsableFrameID attempts to return a usable frame.FrameID. It is used in the event
// that the buffer pool is "full." It always checks the free list first, and then it
// will fall back to using the replacer.
func (m *BufferManager) getUsableFrameID() (*frame.FrameID, error) {
	// First, we will check our freeList.
	fid, foundInFreeList := m.getFrameID()
	if fid == nil {
		return nil, errs.ErrUsableFrameNotFound
	}
	// If we find ourselves here we have a frame.FrameID, but we do not yet know if
	// it came from our free list, or from our replacement policy. If it came due to
	// our replacement policy, then we will need to check to make sure it has not
	// been marked dirty, otherwise we must flush the contents to disk before reusing
	// the frame.FrameID; so let us check on that.
	if !foundInFreeList {
		cf := m.pool[*fid]
		if &cf != nil {
			// We've located the correct frame.Frame in the pool.
			if cf.IsDirty {
				// And it appears that it is in fact holding a dirty page.Page. We must
				// flush the dirty page.Page to disk before recycling this frame.Frame.
				err := m.disk.WritePage(cf.PID, cf.Page)
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
	// Finally, return our *frame.FrameID, and a nil error
	return fid, nil
}

// flushAll attempts to flush any dirty page.Page data.
func (m *BufferManager) flushAll() error {
	// We will range all the entries in the pageTable and call Flush on each one.
	for pid := range m.pageTable {
		err := m.FlushPage(pid)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close attempts to close the BufferManager along with the underlying FileManager
// and associated dependencies. Close makes sure to flush any dirty page.Page data
// before closing everything down.
func (m *BufferManager) Close() error {
	// Make sure all dirty page.Page data is written
	err := m.flushAll()
	if err != nil {
		return err
	}
	// Close the FileManager
	err = m.disk.Close()
	if err != nil {
		return err
	}
	return nil
}
