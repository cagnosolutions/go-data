package dbms

import (
	"errors"
	"fmt"
	"io"
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
	maxSegmentSize  = 16 << 20
	pagesPerSegment = maxSegmentSize / page.PageSize
	currentSegment  = "dat-current.seg"
	segmentPrefix   = "dat-"
	segmentSuffix   = ".seg"
	extentSize      = 64 << 10
)

type BitsetIndex [16]uint64

func NewBitsetIndex() *BitsetIndex {
	return new(BitsetIndex)
}

func (b *BitsetIndex) HasBit(i uint) bool {
	return ((*b)[i>>6] & (1 << (i & 63))) != 0
}

func (b *BitsetIndex) SetBit(i uint) {
	(*b)[i>>6] |= 1 << (i & 63)
}

func (b *BitsetIndex) GetBit(i uint) uint64 {
	return (*b)[i>>6] & (1 << (i & 63))
}

func (b *BitsetIndex) UnsetBit(i uint) {
	(*b)[i>>6] &^= 1 << (i & 63)
}

func (b *BitsetIndex) GetFree() int {
	for j, n := range b {
		if n < ^uint64(0) {
			for bit := uint(j * 64); bit < uint((j*64)+64); bit++ {
				if !b.HasBit(bit) {
					return int(bit)
				}
			}
		}
	}
	return -1
}

func (b *BitsetIndex) String() string {
	resstr := strconv.Itoa(64)
	return fmt.Sprintf("%."+resstr+"b (%d bits)", *b, 64*len(*b))
}

type FileIndex struct {
	ID          int
	Name        string
	FirstPageID page.PageID
	LastPageID  page.PageID
	Index       *BitsetIndex
}

// FileManager is a structure responsible for creating and managing files and segments
// on disk. A file manager instance is only responsible for one namespace at a time.
type FileManager struct {
	latch         sync.Mutex
	namespace     string
	file          *os.File
	nextPageID    page.PageID
	nextSegmentID int
	segments      []FileIndex
	size          int64
	maxSize       int64
}

func MakeFileNameFromID(index int) string {
	hexa := strconv.FormatInt(int64(index), 16)
	return fmt.Sprintf("%s%04s%s", segmentPrefix, hexa, segmentSuffix)
}

func GetIDFromFileName(name string) int {
	hexa := name[len(segmentPrefix) : len(name)-len(segmentSuffix)]
	id, err := strconv.ParseInt(hexa, 16, 32)
	if err != nil {
		panic("GetIDFromFileName: " + err.Error())
	}
	return int(id)
}

func IndexFile(path string) (*FileIndex, error) {
	// check to make sure path exists before continuing
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	// attempt to open existing segment file for reading
	fd, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	// defer file close
	defer func(fd *os.File) {
		_ = fd.Close()
	}(fd)
	// get the id from the file name
	id := GetIDFromFileName(fd.Name())
	// get the page boundaries
	first, last := PageRangeForFile(id)
	// create a new index to start filling out
	idx := &FileIndex{
		ID:          id,
		Name:        fd.Name(),
		FirstPageID: page.PageID(first),
		LastPageID:  page.PageID(last),
		Index:       NewBitsetIndex(),
	}
	// fill out the bitset index
	var pageNo int64
	status := make([]byte, 2)
	for {
		// read the page header into the buffer
		_, err := fd.ReadAt(status, (pageNo*page.PageSize)+16)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// decode the page status
		if page.InUse(status) {
			// if the page is marked as used, then we must
			// set the bit for this page
			idx.Index.SetBit(uint(pageNo))
		}
		// increment to the next page number
		pageNo++
	}
	// we have our file index built, so return it
	return idx, nil
}

func LoadSegmentIndexes(dir string) ([]FileIndex, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var idxs []FileIndex
	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), segmentPrefix) {
			continue
		}
		if strings.HasPrefix(file.Name(), segmentPrefix) {
			if strings.Contains(file.Name(), currentSegment) {
				idx, err := IndexFile(filepath.Join(dir, file.Name()))
				if err != nil {
					return nil, err
				}
				idxs = append(idxs, *idx)
				continue
			}
			idx, err := IndexFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			idxs = append(idxs, *idx)
		}
	}
	return idxs, nil
}

func GetSegmentIDs(dir string) ([]int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var sids []int
	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), segmentPrefix) {
			continue
		}
		if strings.HasPrefix(file.Name(), segmentPrefix) {
			if strings.Contains(file.Name(), currentSegment) {
				sids = append(sids, -1)
				continue
			}
			sids = append(sids, GetIDFromFileName(file.Name()))
		}
	}
	sort.Ints(sids)
	return sids, nil
}

// PageInFile returns a boolean indicating true if the provided PageID is within the
// bounds of the provided segment ID, and false if they are outside the bounds.
func PageInFile(pid page.PageID, sid int) bool {
	return (pagesPerSegment*sid) <= int(pid) && int(pid) <= ((pagesPerSegment*sid)+pagesPerSegment-1)
}

// FileForPage takes a PageID and returns the ID of the segment where that page should
// be found.
func FileForPage(pid page.PageID) int {
	return int(pid) / pagesPerSegment
}

// PageRangeForFile takes a segment ID and returns the beginning and ending page ID's
// that the segment with the provided ID should contain.
func PageRangeForFile(sid int) (int, int) {
	return pagesPerSegment * sid, (pagesPerSegment * sid) + pagesPerSegment - 1
}

// OpenFileManager opens an existing file manager instance if one exists with the same
// namespace otherwise it creates a new instance and returns it along with any potential
// errors encountered.
func OpenFileManager(namespace string) (*FileManager, error) {
	// clean namespace path
	path, err := disk.PathCleanAndTrimSuffix(namespace)
	if err != nil {
		return nil, err
	}
	// open the current file segment
	fd, err := disk.FileOpenOrCreate(filepath.Join(path, currentSegment))
	if err != nil {
		return nil, err
	}
	// get the segment id for this namespace
	sids, err := GetSegmentIDs(path)
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
		segments:      make([]FileIndex, 0),
		size:          size,
		maxSize:       maxSegmentSize,
	}
	// m.segments = FileIndex{
	// 	ID:          m.nextSegmentID,
	// 	FirstPageID: page.PageID(pagesPerSegment * m.nextSegmentID),
	// 	LastPageID:  page.PageID((pagesPerSegment * m.nextSegmentID) + pagesPerSegment - 1),
	// }
	return m, nil
}

func (f *FileManager) LoadSegmentIndex(id int) error {
	return nil
}

// CheckSpace takes a segment ID and checks to make sure there is room in the matching
// segment to
// allocate a new page. If there is not enough room left in the current segment
// a new segment must be swapped in. CheckSpace returns nil as long as there is
// room in
// to the current file segment would cause it to be above the maximum size threshold, a
// new current segment will be created, and swapped in.
func (f *FileManager) CheckSpace(id int) error {
	// first, we check to make sure there is enough room in the current segment file to
	// allocate an additional extent.
	if f.size+extentSize >= maxSegmentSize {
		// there is not enough room in the current file segment to allocate another
		// extent, so first we close the current file segment
		err := f.file.Close()
		if err != nil {
			return err
		}
		// next, we rename the current file segment to be the nextSegmentID
		err = os.Rename(currentSegment, MakeFileNameFromID(f.nextSegmentID))
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
		// and we return
		return nil
	}
	// allocate an extent.
	err := f.file.Truncate(f.size + extentSize)
	if err != nil {
		return err
	}
	// finally, we can return
	return nil
}

// AllocateExtent adds a new extent to the current file segment. If adding a new extent
// to the current file segment would cause it to be above the maximum size threshold, a
// new current segment will be created, and swapped in.
func (f *FileManager) AllocateExtent() error {
	// first, we check to make sure there is enough room in the current segment file to
	// allocate an additional extent.
	if f.size+extentSize >= maxSegmentSize {
		// there is not enough room in the current file segment to allocate another
		// extent, so first we close the current file segment
		err := f.file.Close()
		if err != nil {
			return err
		}
		// next, we rename the current file segment to be the nextSegmentID
		err = os.Rename(currentSegment, MakeFileNameFromID(f.nextSegmentID))
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
		// and we return
		return nil
	}
	// allocate an extent.
	err := f.file.Truncate(f.size + extentSize)
	if err != nil {
		return err
	}
	// finally, we can return
	return nil
}

// AllocatePage simply returns the next logical page ID that is can be written to.
func (f *FileManager) AllocatePage() page.PageID {
	// increment and return the next PageID
	return atomic.SwapUint32(&f.nextPageID, f.nextPageID+1)
}

// DeallocatePage writes zeros to the page located at the logical address
// calculated using the page ID provided.
func (f *FileManager) DeallocatePage(pid page.PageID) error {
	// Locate the segment file for the supplied page ID
	// id := FileForPage(pid)
	// Check to ensure the page we want to deallocate is in the current file segment
	// if f.currentSegmentIndex.ID {
	//
	//	}
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

// Close closes the file manager instance
func (f *FileManager) Close() error {
	return f.file.Close()
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
