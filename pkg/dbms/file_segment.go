package dbms

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

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

// FileSegment is an in memory index of a file segment.
type FileSegment struct {
	ID       int
	Name     string
	FirstPID page.PageID
	LastPID  page.PageID
	Size     int64
	Index    *BitsetIndex
}

// NewFileSegment creates and returns a new *FileSegment struct
func NewFileSegment(path string) *FileSegment {
	// get the base filename
	base := filepath.Base(path)
	// get the id from the file name
	var id int
	if strings.Contains(base, currentSegment) {
		id = -1
	} else {
		id = GetIDFromFileName(base)
	}
	// get the page boundaries
	first, last := PageRangeForFile(id)
	// create and return new *FileSegment instance.
	return &FileSegment{
		ID:       id,
		Name:     path,
		FirstPID: page.PageID(first),
		LastPID:  page.PageID(last),
		Index:    NewBitsetIndex(),
	}
}

// LoadIndex checks to see if it can find a matching index file on disk, if it finds
// one, the index will be loaded from that index file. Otherwise, it will rebuild the
// index by reading from the data file segment directly.
func (fs *FileSegment) LoadIndex() error {
	// check for an index file
	indexName := strings.Replace(fs.Name, segmentSuffix, segmentIndexSuffix, 1)
	_, err := os.Stat(indexName)
	if os.IsExist(err) {
		// found a matching index file, load the index file directly
		err = fs.Index.ReadFile(indexName)
		if err != nil {
			return err
		}
		// all good, we can return
		return nil
	}
	// otherwise, we must manually rebuild the index so first attempt to open
	// the existing data segment file for reading
	fd, err := os.OpenFile(fs.Name, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	// defer file close
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
	// we have our file index built and since we had to rebuild it, we
	// should write it out, so we do not have to rebuild it next time.
	err = fs.Index.WriteFile(indexName)
	if err != nil {
		return err
	}
	return nil
}

// OpenFileSegment opens the named FileSegment. If it can find a matching index file
// on disk the on disk index will be loaded directly from file. Otherwise, it will
// build the index by reading from data file segment directly. If the named data file
// segment does not  exist, an error will be returned.
func OpenFileSegment(path string) (*FileSegment, error) {
	// check to make sure path exists before continuing
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}
	// create a new FileSegment instance.
	fs := NewFileSegment(path)
	// load file segment index
	err = fs.LoadIndex()
	if err != nil {
		return nil, err
	}
	// we are finished
	return fs, nil
}
