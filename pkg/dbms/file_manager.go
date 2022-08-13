package dbms

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cagnosolutions/go-data/pkg/dbms/disk"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

// FileManager is a structure responsible for creating and managing files and segments
// on disk. A file manager instance is only responsible for one namespace at a time.
type FileManager struct {
	latch         sync.Mutex
	namespace     string
	file          *os.File
	nextPageID    page.PageID
	nextSegmentID int
	segments      []FileSegment
	size          int64
	maxSize       int64
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
	// current, err := OpenFileSegment(filepath.Join(path, currentSegment))
	// if err != nil {
	//	return nil, err
	// }
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
	// initialize a new *FileManager instance
	m := &FileManager{
		namespace:     path,
		file:          fd,
		nextPageID:    page.PageID(size / page.PageSize),
		nextSegmentID: len(sids),
		segments:      make([]FileSegment, 0),
		size:          size,
		maxSize:       maxSegmentSize,
	}
	// populate the FileSegment set
	err = m.LoadFileSegments()
	if err != nil {
		return nil, err
	}
	// finished, return file manager instance
	return m, nil
}

// LoadFileSegments walks the base namespace directory and attempts to populate
// the FileManager's segment file set with all the necessary file segment and
// file segment index data.
func (f *FileManager) LoadFileSegments() error {
	// read the directory
	files, err := os.ReadDir(f.namespace)
	if err != nil {
		return err
	}
	// range the files in the base namespace directory
	for _, file := range files {
		// skip any directories or non-segment files
		if file.IsDir() || !strings.HasSuffix(file.Name(), segmentSuffix) {
			continue
		}
		// handle loading a segment
		if strings.HasSuffix(file.Name(), segmentSuffix) {
			fs, err := OpenFileSegment(filepath.Join(f.namespace, file.Name()))
			if err != nil {
				return err
			}
			f.segments = append(f.segments, *fs)
		}
	}
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
