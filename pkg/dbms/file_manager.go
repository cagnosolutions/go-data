package dbms

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cagnosolutions/go-data/pkg/dbms/disk"
	"github.com/cagnosolutions/go-data/pkg/dbms/errs"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

type FileSegments []FileSegment

func (fs FileSegments) Len() int {
	return len(fs)
}

func (fs FileSegments) Less(i, j int) bool {
	return fs[i].ID < fs[j].ID
}

func (fs FileSegments) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

// FileManager is a structure responsible for creating and managing files and segments
// on disk. A current manager instance is only responsible for one namespace at a time.
type FileManager struct {
	latch      sync.Mutex
	namespace  string
	file       *os.File
	nextPageID page.PageID
	segments   FileSegments
	current    *FileSegment
	size       int64
	maxSize    int64
}

// OpenFileManager opens an existing current manager instance if one exists with the same
// namespace otherwise it creates a new instance and returns it along with any potential
// errors encountered.
func OpenFileManager(namespace string) (*FileManager, error) {
	// clean namespace path
	path, err := disk.PathCleanAndTrimSuffix(namespace)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(path, 0644)
	if err != nil {
		return nil, err
	}
	// // open the current file segment
	// fd, err := disk.FileOpenOrCreate(filepath.Join(path, currentSegment))
	// if err != nil {
	// 	return nil, err
	// }
	// // get the size of the current file segment
	// fi, err := os.Stat(fd.Name())
	// if err != nil {
	// 	return nil, err
	// }
	// size := fi.Size()
	// initialize a new *FileManager instance
	m := &FileManager{
		namespace: path,
		// file:       fd,
		// nextPageID: page.PageID(size / page.PageSize),
		segments: make(FileSegments, 0),
		current:  nil,
		// size:       size,
		maxSize: maxSegmentSize,
	}
	// populate the FileSegment set
	err = m.LoadFileSegments()
	if err != nil {
		return nil, err
	}
	m.nextPageID = page.PageID(m.segments[len(m.segments)-1].Index.GetFree())
	// finished, return current manager instance
	return m, nil
}

// LoadFileSegments walks the base namespace directory and attempts to populate
// the FileManager's segment current set with all the necessary current segment and
// current segment index data.
func (f *FileManager) LoadFileSegments() error {
	// read the directory
	files, err := os.ReadDir(f.namespace)
	if err != nil {
		return err
	}
	// handle case where we are just starting out
	if len(files) == 0 {
		// create our initial file
		fd, err := disk.FileOpenOrCreate(filepath.Join(f.namespace, MakeFileNameFromID(0)))
		if err != nil {
			return err
		}
		// don't forget to close
		defer func(fd *os.File) {
			err := fd.Close()
			if err != nil {
				panic(err)
			}
		}(fd)
		// create a new current segment, and return
		f.current = NewFileSegment(f.namespace, 0)
		return nil
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
	// check to see if we need to add any
	if len(f.segments) == 0 {
		// create a new current segment, and return
		f.current = NewFileSegment(f.namespace, 0)
		return nil
	}
	// sort them by id
	sort.Stable(f.segments)
	// set the current file
	f.current = &f.segments[len(f.segments)-1]
	f.file, err = disk.FileOpenOrCreate(f.current.Name)
	if err != nil {
		return err
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
// func (f *FileManager) CheckSpace(id int) error {
// 	// first, we check to make sure there is enough room in the current segment current to
// 	// allocate an additional extent.
// 	if f.size+extentSize >= maxSegmentSize {
// 		// there is not enough room in the current file segment to allocate another
// 		// extent, so first we close the current file segment
// 		err := f.current.Close()
// 		if err != nil {
// 			return err
// 		}
// 		sid := f.segments[len(f.segments)-1].ID
// 		// next, we rename the current file segment to be the nextSegmentID
// 		err = os.Rename(currentSegment, MakeFileNameFromID(sid))
// 		if err != nil {
// 			return err
// 		}
// 		// and, then we create and open a new current segment current
// 		f.current, err = disk.FileOpenOrCreate(filepath.Join(f.namespace, currentSegment))
// 		if err != nil {
// 			return err
// 		}
// 		// then, we add a new segment to the segments set
// 		f.segments = append(f.segments, *NewFileSegment(f.namespace))
// 		// and we return
// 		return nil
// 	}
// 	// allocate an extent.
// 	err := f.current.Truncate(f.size + extentSize)
// 	if err != nil {
// 		return err
// 	}
// 	// finally, we can return
// 	return nil
// }

// AllocatePage simply returns the next logical page ID that is can be written to.
func (f *FileManager) AllocatePage() page.PageID {
	// increment and return the next PageID
	return atomic.SwapUint32(&f.nextPageID, f.nextPageID+1)
}

// DeallocatePage writes zeros to the page located at the logical address
// calculated using the page ID provided.
func (f *FileManager) DeallocatePage(pid page.PageID) error {
	// Locate the segment file for the supplied page ID
	id := FileForPage(pid)
	// Check to ensure the page we want to deallocate is in the current file segment
	if f.current.ID != id {
		// The provided pid does not match the current file segment, so we must try
		// to locate and load the correct one.
		if len(f.segments) > id {
			// The file segment that would match up with the proper ID does not exist,
			// so we should now return an error
			return errs.ErrSegmentNotFound
		}
		// Otherwise, the segment does exist, and we should just change the current
		// file segment to the correct one, and open that file descriptor.
		f.current = &f.segments[id]
		// Close the current file descriptor
		err := f.file.Close()
		if err != nil {
			return err
		}
		// Open the new segment file
		f.file, err = disk.FileOpenOrCreate(f.current.Name)
		if err != nil {
			return err
		}
	}
	// The provided pid does indeed match fall within the range that should be in our
	// current file segment, so we should now locate the logical page offset.
	off := int64(pid * page.PageSize)
	// Next, we will create an empty page
	ep := page.NewEmptyPage(pid)
	// Then, we can attempt to write the contents of the empty page data directly to
	// the calculated offset
	_, err := f.file.WriteAt(ep, off)
	if err != nil {
		return err
	}
	// Don't forget to sync it up
	err = f.file.Sync()
	if err != nil {
		return err
	}
	// Finally, update the segment index, and return a nil error
	f.current.Index.UnsetBit(uint(pid))
	err = f.current.WriteIndex()
	if err != nil {
		return err
	}
	return nil
}

// ReadPage reads the page located at the logical address calculated using the
// page ID provided.
func (f *FileManager) ReadPage(pid page.PageID, p page.Page) error {
	// Locate the segment file for the supplied page ID
	id := FileForPage(pid)
	// Check to ensure the page we want to deallocate is in the current file segment
	if f.current.ID != id {
		// The provided pid does not match the current file segment, so we must try
		// to locate and load the correct one.
		if len(f.segments) > id {
			// The file segment that would match up with the proper ID does not exist,
			// so we should now return an error
			return errs.ErrSegmentNotFound
		}
		// Otherwise, the segment does exist, and we should just change the current
		// file segment to the correct one, and open that file descriptor.
		f.current = &f.segments[id]
		// Close the current file descriptor
		err := f.file.Close()
		if err != nil {
			return err
		}
		// Open the new segment file
		f.file, err = disk.FileOpenOrCreate(f.current.Name)
		if err != nil {
			return err
		}
	}
	// The provided pid does indeed match fall within the range that should be in our
	// current file segment, so we should now locate the logical page offset.
	off := int64(pid * page.PageSize)
	// Then, we can attempt to read the contents of the supplied page data directly
	// from the calculated offset
	_, err := f.file.ReadAt(p, off)
	if err != nil {
		return err
	}
	// Finally, return a nil error
	return nil
}

// WritePage writes the page located at the logical address calculated using the
// page ID provided.
func (f *FileManager) WritePage(pid page.PageID, p page.Page) error {
	// Locate the segment current for the supplied page ID
	id := FileForPage(pid)
	// Check to ensure the page we want to deallocate is in the current file segment
	if f.current.ID != id {
		// The provided pid does not match the current file segment, so we must try
		// to locate and load the correct one.
		if len(f.segments) > id {
			// The file segment that would match up with the proper ID does not exist,
			// so we should now return an error
			return errs.ErrSegmentNotFound
		}
		// Otherwise, the segment does exist, and we should just change the current
		// file segment to the correct one, and open that file descriptor.
		f.current = &f.segments[id]
		// Close the current file descriptor
		err := f.file.Close()
		if err != nil {
			return err
		}
		// Open the new segment file
		f.file, err = disk.FileOpenOrCreate(f.current.Name)
		if err != nil {
			return err
		}
	}
	// The provided pid does indeed match fall within the range that should be in our
	// current file segment, so we should now locate the logical page offset.
	off := int64(pid * page.PageSize)
	// Then, we can attempt to write the contents of the supplied page data directly to
	// the calculated offset
	_, err := f.file.WriteAt(p, off)
	if err != nil {
		return err
	}
	// Don't forget to sync it up
	err = f.file.Sync()
	if err != nil {
		return err
	}
	// Finally, update the segment index, and return a nil error
	f.current.Index.SetBit(uint(pid))
	err = f.current.WriteIndex()
	if err != nil {
		return err
	}
	return nil
}

// Close closes the current manager instance
func (f *FileManager) Close() error {
	return f.file.Close()
}
