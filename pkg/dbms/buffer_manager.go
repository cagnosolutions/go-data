package dbms

import (
	"errors"
	"sync"

	"github.com/cagnosolutions/go-data/pkg/dbms/frame"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

type BufferPool interface {

	// AddFreeFrame takes a frameID and adds it to the set of free frames list.
	// USES: freeList
	AddFreeFrame(fid frame.FrameID)

	// GetFrameID attempts to return a frameID. It first checks the free frame
	// list to see if there are any free frames in there. If one is found it
	// will return it along with a boolean indicating true. If none are found,
	// it will then go on to the replacer in search of one.
	// USES: freeList, Replacer
	GetFrameID() (*frame.FrameID, bool)
}

type Replacer interface {
	// Pin pins the frame matching the supplied frame ID, indicating that it should
	// not be victimized until it is unpinned.
	Pin(fid frame.FrameID)

	// Victim removes and returns the next "victim frame", as defined by the policy.
	Victim() *frame.FrameID

	// Unpin unpins the frame matching the supplied frame ID, indicating that it may
	// now be victimized.
	Unpin(fid frame.FrameID)
}

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
	pool      []frame.FrameID               // buffer pool page frames
	replacer  Replacer                      // page replacement policy structure
	io        FileManager                   // underlying file manager
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
	// open disk manager

	return nil, nil
}

// NewPage returns a fresh empty page from the pool.
func (m *BufferManager) NewPage() page.Page { return nil }

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

// GetUsableFrame attempts to return a usable frameID. It is used in the event that
// the buffer pool is "full." It always checks the free list first, and then it will
// fall back to using the replacer.
func (m *BufferManager) GetUsableFrame() (*frame.FrameID, bool) { return nil, false }

// Close closes the buffer manager.
func (m *BufferManager) Close() error { return nil }
