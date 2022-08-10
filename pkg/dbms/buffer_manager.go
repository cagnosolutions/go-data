package dbms

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cagnosolutions/go-data/pkg/dbms/disk"
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
	maxSegmentSize = 16 << 20
	currentSegment = "seg-current.db"
	extentSize     = 64 << 10
)

// seg-000001-.db
// seg-current.db

// FileManager is a structure responsible for creating and managing files and segments
// on disk. A file manager instance is only responsible for one namespace at a time.
type FileManager struct {
	namespace     string
	file          *os.File
	nextPageID    page.PageID
	nextSegmentID int
	size          int64
	maxSize       int64
}

// CreateSegmentFileName takes a namespace, and an integer as an ID, and returns a segment file name.
func CreateSegmentFileName(namespace string, id int) string {
	return fmt.Sprintf("seg-%6d-%4s.db", id, namespace)
}

func GetSegmentIDs(dir string) ([]int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var sids []int
	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "seg-") {
			continue
		}
		if strings.HasPrefix(file.Name(), "seg-") {
			if strings.Contains(file.Name(), currentSegment) {
				sids = append(sids, -1)
				continue
			}
			sid, err := strconv.ParseInt(file.Name(), 16, 64)
			if err != nil {
				return nil, err
			}
			sids = append(sids, int(sid))
		}
	}
	sort.Ints(sids)
	return sids, nil
}

// OpenFileManager opens an existing file manager instance if one exists with the same
// namespace otherwise it creates a new instance and returns it along with any potential
// errors encountered.
func OpenFileManager(namespace string) (*FileManager, error) {
	// clean namespace path
	path, err := disk.PathCleanAndTrimSUffix(namespace)
	if err != nil {
		return nil, err
	}
	// get the current segment id's
	sids, err := GetSegmentIDs(path)
	if err != nil {
		return nil, err
	}
	// open the current file segment
	fd, err := disk.FileOpenOrCreate(filepath.Join(path, currentSegment))
	if err != nil {
		return nil, err
	}
	// get the size of the current file segment
	fi, err := os.Stat(fd.Name())
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	// fill and return a new FileManager instance
	m := &FileManager{
		namespace:     path,
		file:          fd,
		nextPageID:    page.PageID(size / page.PageSize),
		nextSegmentID: len(sids),
		size:          size,
		maxSize:       maxSegmentSize,
	}
	return m, nil
}

func (f *FileManager) LoadSegment(id int) error {
	return nil
}

// AllocateExtent adds a new extent to the current file segment. If adding a new extent
// to the current file segment would cause it to be above the maximum size threshold, a
// new current segment will be created, and swapped in.
func (f *FileManager) AllocateExtent() error {
	// first, we check to make sure there is enough room in the current segment file to
	// allocate an additional extent.
	if f.size+extentSize < maxSegmentSize {
		// we have room to add an extent
		err := f.file.Truncate(extentSize)
		if err != nil {
			return err
		}
	} else {
		// otherwise, there is not enough room in the current file segment to allocate another
		// extent, so first we close the current file segment
		err := f.file.Close()
		if err != nil {
			return err
		}
		// next, we rename the current file segment to be the nextSegmentID
		err = os.Rename(currentSegment, CreateSegmentFileName(f.namespace, f.nextSegmentID))
		if err != nil {
			return err
		}
		// then, we increment the nextSegmentID
		f.nextSegmentID++
		// and, finally, we create and open a new current segment file
		f.file, err = disk.FileOpenOrCreate(filepath.Join(f.namespace, currentSegment))
		if err != nil {
			return err
		}
	}
	// finally, we can return
	return nil
}

// AllocatePage makes sure there is space for a new page, allocating space if
// necessary, and returning the next logical page ID.
func (f *FileManager) AllocatePage() page.PageID {
	// increment and return the next PageID
	return atomic.SwapUint32(&f.nextPageID, f.nextPageID+1)
}

// DeallocatePage writes zeros to the page located at the logical address
// calculated using the page ID provided.
func (f *FileManager) DeallocatePage(pid page.PageID) error {
	return nil
}

// ReadPage reads the page located at the logical address calculated using the
// page ID provided.
func (f *FileManager) ReadPage(pid page.PageID, page page.Page) error {
	return nil
}

// WritePage writes the page located at the logical address calculated using the
// page ID provided.
func (f *FileManager) WritePage(pid page.PageID, page page.Page) error {
	return nil
}

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

// Open opens an existing storage manager instance if one exists with the same namespace
// otherwise it creates a new instance and returns it.
func Open(base string, pageCount uint16) (*BufferManager, error) {
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
