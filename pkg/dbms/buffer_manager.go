package dbms

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cagnosolutions/go-data/pkg/dbms/frame"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

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

// BufferManager is the access level structure wrapping up the BufferPool, and FileManager,
// along with a page table, and replacement policy.
type BufferManager struct {
	latch     sync.Mutex
	pool      []frame.Frame                 // buffer pool page frames
	replacer  *ClockReplacer                // page replacement policy structure
	io        *FileManager                  // underlying file manager
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
		io:        fm,
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
//
// 1. If the pool is full and all pages are pinned, return nil
// 2. Pick a victim page P
// 		a. First look in the free list for P
// 		b. If P cannot be found in the free list, use the replacement policy
// 3. Update P's metadata. Zero out memory and add P to the page table.
// 4. Return a pointer to P
//
func (m *BufferManager) NewPage() page.Page {
	fid, err := m.GetUsableFrame()
	if err != nil {
		return nil
	}
	pid := m.io.AllocatePage()
	pf := frame.NewFrame(pid, *fid, page.PageSize)
	pg := page.NewPage(pid)
	copy(pf.Page, pg)
	m.pageTable[pid] = *fid
	m.pool[*fid] = pf
	return pf.Page
}

// FetchPage retrieves specific page from the pool, or storage medium by the page ID.
func (m *BufferManager) FetchPage(pid page.PageID) page.Page { return nil }

// UnpinPage allows for manual unpinning of a specific page from the pool by the page ID.
func (m *BufferManager) UnpinPage(pid page.PageID, isDirty bool) error { return nil }

// FlushPage forces a page to be written onto the storage medium, and decrements the
// pin count on the frame potentially enabling the frame to be reused.
func (m *BufferManager) FlushPage(pid page.PageID) error { return nil }

// DeletePage removes the page from the buffer pool, and decrements the pin count on the
// frame potentially enabling the frame to be reused.
func (m *BufferManager) DeletePage(pid page.PageID) error { return nil }

// GetFrameID attempts to return a *frame.FrameID. It first checks the freeList set to
// see if there are any availble frames to pick from. If not, it will proceed to use
// the replacement policy to locate one.
func (m *BufferManager) GetFrameID() (*frame.FrameID, bool) {
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

// GetUsableFrame attempts to return a usable frameID. It is used in the event that
// the buffer pool is "full." It always checks the free list first, and then it will
// fall back to using the replacer.
func (m *BufferManager) GetUsableFrame() (*frame.FrameID, error) {
	fid, foundInFreeList := m.GetFrameID()
	if fid == nil {
		return nil, fmt.Errorf("usable frame not found")
	}
	if !foundInFreeList {
		cf := m.pool[*fid]
		if &cf != nil {
			if cf.IsDirty {
				err := m.io.WritePage(cf.PID, cf.Page)
				if err != nil {
					return nil, err
				}
			}
			delete(m.pageTable, cf.PID)
		}
	}
	return fid, nil
}
