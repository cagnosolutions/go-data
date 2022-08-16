package dbms

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cagnosolutions/go-data/pkg/dbms/disk"
	"github.com/cagnosolutions/go-data/pkg/dbms/errs"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

type SegmentID uint32

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
	latch     sync.Mutex
	namespace string
	file      *os.File
	nextPID   uint32
	nextSID   uint32
	segments  FileSegments
	current   *FileSegment
	size      int64
	maxSize   int64
}

// OpenFileManager opens an existing current manager instance if one exists with the same
// namespace otherwise it creates a new instance and returns it along with any potential
// errors encountered.
func OpenFileManager(namespace string) (*FileManager, error) {
	// Clean namespace path
	path, err := disk.PathCleanAndTrimSuffix(namespace)
	if err != nil {
		return nil, err
	}
	// Create any directories that may not exist
	err = os.MkdirAll(path, 0644)
	if err != nil {
		return nil, err
	}
	// Initialize a new *FileManager instance
	f := &FileManager{
		namespace: path,
		nextPID:   0,
		nextSID:   1,
		segments:  make(FileSegments, 0),
		current:   nil,
		maxSize:   maxSegmentSize,
	}
	// Load the segments list along with their indexes
	err = f.LoadFileSegments()
	if err != nil {
		return nil, err
	}
	// Finished, return *FileManager instance
	return f, nil
}

// LoadFileSegments walks the base namespace directory and attempts to populate
// the FileManager's segment current set with all the necessary current segment and
// current segment index data.
func (f *FileManager) LoadFileSegments() error {
	// Enable the latch, while we load
	f.latch.Lock()
	defer f.latch.Unlock()
	// Read the file listing in the namespace directory
	files, err := os.ReadDir(f.namespace)
	if err != nil {
		return err
	}
	// Iterate over the directory listing, and only act on the segment files
	for _, file := range files {
		// Skip non segment files
		if file.IsDir() || !strings.HasSuffix(file.Name(), segmentSuffix) {
			continue
		}
		// Attempt to load the segment file
		segment, err := LoadFileSegment(filepath.Join(f.namespace, file.Name()))
		if err != nil {
			return err
		}
		// Add file segment to our segments list
		f.segments = append(f.segments, *segment)
	}
	// Check the segment file list; if none were found, initialize a new one.
	if len(f.segments) == 0 {
		// No segments were found; initialize a new one.
		segment, err := f.MakeFileSegment(f.nextSID)
		if err != nil {
			return err
		}
		// New file segment has been created successfully; add to our segments list.
		f.segments = append(f.segments, *segment)
	}
	// Now, we update the current segment pointer to point to the tail segment.
	f.current = &f.segments[len(f.segments)-1]
	// Next, we must update the nextPID, and nextSID
	f.nextPID = f.current.PageOffset()
	f.nextSID = f.current.ID
	return nil
}

func (f *FileManager) LoadFileSegment(path string) (*FileSegment, error) {
	return nil, nil
}

// MakeFileSegment takes an ID and uses it to create and return a new *FileSegment
func (f *FileManager) MakeFileSegment(id uint32) (*FileSegment, error) {
	// Create the filename
	name := filepath.Join(f.namespace, MakeFileNameFromID(id))
	// Create a new file for this file segment
	fd, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	// Don't forget to close!
	err = fd.Close()
	if err != nil {
		return nil, err
	}
	// Get the page boundaries
	first, last := PageRangeForFile(id)
	// Create and return new *FileSegment
	fs := &FileSegment{
		ID:       id,
		Name:     name,
		FirstPID: first,
		LastPID:  last,
		Index:    NewBitsetIndex(),
	}
	return fs, nil
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
	return atomic.SwapUint32(&f.nextPID, f.current.PageOffset())
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
		if len(f.segments) > int(id) {
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
		if len(f.segments) > int(id) {
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
		if len(f.segments) > int(id) {
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
