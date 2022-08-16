package dbms

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

func MakeFileNameFromID(index uint32) string {
	hexa := strconv.FormatInt(int64(index), 16)
	return fmt.Sprintf("%s%04s%s", segmentPrefix, hexa, segmentSuffix)
}

func GetIDFromFileName(name string) uint32 {
	hexa := name[len(segmentPrefix) : len(name)-len(segmentSuffix)]
	id, err := strconv.ParseUint(hexa, 16, 32)
	if err != nil {
		panic("GetIDFromFileName: " + err.Error())
	}
	return uint32(id)
}

// func GetSegmentIDs(dir string) ([]uint32, error) {
// 	files, err := os.ReadDir(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var sids []uint32
// 	for _, file := range files {
// 		if file.IsDir() || !strings.HasPrefix(file.Name(), segmentPrefix) {
// 			continue
// 		}
// 		if strings.HasPrefix(file.Name(), segmentPrefix) {
// 			if strings.Contains(file.Name(), currentSegment) {
// 				sids = append(sids, -1)
// 				continue
// 			}
// 			sids = append(sids, GetIDFromFileName(file.Name()))
// 		}
// 	}
// 	sort.Slice(
// 		sids, func(i, j int) bool {
// 			return sids[i] < sids[j]
// 		},
// 	)
// 	return sids, nil
// }

// PageInFile returns a boolean indicating true if the provided PageID is within the
// bounds of the provided segment ID, and false if they are outside the bounds.
func PageInFile(pid page.PageID, sid uint32) bool {
	return (pagesPerSegment*sid) <= pid && pid <= ((pagesPerSegment*sid)+pagesPerSegment-1)
}

// FileForPage takes a PageID and returns the ID of the segment where that page should
// be found.
func FileForPage(pid page.PageID) uint32 {
	return pid / pagesPerSegment
}

// PageRangeForFile takes a segment ID and returns the beginning and ending page ID's
// that the segment with the provided ID should contain.
func PageRangeForFile(sid uint32) (uint32, uint32) {
	return pagesPerSegment * sid, (pagesPerSegment * sid) + pagesPerSegment - 1
}

// FileSegment is an in memory index of a current segment.
type FileSegment struct {
	ID       uint32
	Name     string
	FirstPID page.PageID
	LastPID  page.PageID
	Size     int64
	Index    *BitsetIndex
	Cursor   int64
}

// LoadFileSegment opens the named FileSegment. If it can find a matching index current
// on disk the on disk index will be loaded directly from current. Otherwise, it will
// build the index by reading from data current segment directly. If the named data current
// segment does not  exist, an error will be returned.
func LoadFileSegment(path string) (*FileSegment, error) {
	// Check to make sure path exists before continuing
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}
	// Get the file segment size for later
	size := fi.Size()
	// Get the file segment id from the file name
	id := GetIDFromFileName(filepath.Base(path))
	// Find the PageID boundaries
	first, last := PageRangeForFile(id)
	// Create a new FileSegment instance.
	fs := &FileSegment{
		ID:       id,
		Name:     path,
		FirstPID: first,
		LastPID:  last,
		Size:     size,
		Index:    nil,
	}
	// load current segment index
	err = fs.LoadIndex()
	if err != nil {
		return nil, err
	}
	// we are finished
	return fs, nil
}

// LoadIndex checks to see if it can find a matching index current on disk, if it finds
// one, the index will be loaded from that index current. Otherwise, it will rebuild the
// index by reading from the data current segment directly.
func (fs *FileSegment) LoadIndex() error {
	// check for an index current
	indexName := strings.Replace(fs.Name, segmentSuffix, segmentIndexSuffix, 1)
	_, err := os.Stat(indexName)
	if os.IsExist(err) {
		// found a matching index current, load the index current directly
		err = fs.Index.ReadFile(indexName)
		if err != nil {
			return err
		}
		// all good, we can return
		return nil
	}
	// otherwise, we must manually rebuild the index so first attempt to open
	// the existing data segment current for reading
	fd, err := os.OpenFile(fs.Name, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	// defer current close
	defer func(fd *os.File) {
		_ = fd.Close()
	}(fd)
	// create a buffer to hold each page, and a page counter
	var pageNo int64
	pg := page.Page(make([]byte, page.PageSize))
	for {
		// read the page
		_, err = fd.Read(pg)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// check to see if the page is marked as used
		if pg.IsUsed() {
			// if the page is marked as used, then we must
			// set the bit for this page
			fs.Index.SetBit(uint(pageNo))
		}
		// increment to the next page number
		pageNo++
	}
	// we have our current index built and since we had to rebuild it, we
	// should write it out, so we do not have to rebuild it next time.
	err = fs.Index.WriteFile(indexName)
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileSegment) ReadIndex() error {
	indexName := strings.Replace(fs.Name, segmentSuffix, segmentIndexSuffix, 1)
	return fs.Index.ReadFile(indexName)
}

func (fs *FileSegment) WriteIndex() error {
	indexName := strings.Replace(fs.Name, segmentSuffix, segmentIndexSuffix, 1)
	return fs.Index.WriteFile(indexName)
}

// PageOffset returns the next page offset for writing. It will attempt to use any
// free space that has been deallocated first. Otherwise, sequential access is used.
func (fs *FileSegment) PageOffset() uint32 {
	// First, we will get the population count for the amount of pages that are
	// being used in this file segment.
	return 0
}
