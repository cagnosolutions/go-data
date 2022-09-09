package engine

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

const dataFilePerm = 1466

// DiskManager is a structure responsible for creating and managing access with
// the actual files stored on disk. The current disk manager instance is only
// responsible for dealing with one file at a time.
type DiskManager struct {
	sync.RWMutex
	file    *os.File
	nextPID PageID
	size    int64
}

// OpenDiskManager opens an existing disk manager instance if one exists with the same
// name, otherwise it creates a new instance and returns it along with any potential
// errors encountered.
func OpenDiskManager(path string) (*DiskManager, error) {
	// Clean path
	path, err := filepath.Abs(filepath.ToSlash(path))
	if err != nil {
		return nil, err
	}
	var fp *os.File
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// Create a new instance
		err = os.MkdirAll(filepath.Dir(path), os.ModeDir|dataFilePerm)
		if err != nil {
			return nil, err
		}
		fp, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC, dataFilePerm)
		if err != nil {
			return nil, err
		}
		err = fp.Close()
		if err != nil {
			return nil, err
		}
	}
	// Open file at the fully cleaned path
	fp, err = os.OpenFile(path, os.O_RDWR|os.O_SYNC, dataFilePerm)
	if err != nil {
		return nil, err
	}
	// Get the current file size
	fi, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	nextPageID := PageID(0)
	if size > 0 {
		nextPageID = PageID(size / PageSize)
	}
	// Initialize a new DiskManager instance
	fm := &DiskManager{
		file:    fp,
		nextPID: nextPageID,
		size:    size,
	}
	// Load the meta info for the DiskManager instance
	err = fm.load()
	if err != nil {
		return nil, err
	}
	// Return our instance
	return fm, nil
}

// load attempts to populate our DiskManager instance with metadata about the file.
func (f *DiskManager) load() error {
	return nil
}

// logicalOffset checks for any out of bounds errors, and returns an error if there
// is one. Otherwise, it takes a page ID and returns a logical page offset.
func (f *DiskManager) logicalOffset(pid PageID) (int64, error) {
	// Check to see if the requested pid falls within the set that has been distributed
	if pid > f.nextPID {
		return -1, ErrPageIDHasNotBeenAllocated(pid)
	}
	// We are good, so we will calculate the logical page offset.
	return int64(pid * PageSize), nil
}

// AllocatePage simply returns the next logical page ID that is can be written to.
func (f *DiskManager) AllocatePage() PageID {
	// increment and return the nextpage.PageID
	return PageID(atomic.SwapUint32((*uint32)(&f.nextPID), uint32(f.nextPID+1)))
}

// DeallocatePage writes zeros to the page located at the logical address
// calculated using the page ID provided.
func (f *DiskManager) DeallocatePage(pid PageID) error {
	// Calculate the logical page offset.
	off, err := f.logicalOffset(pid)
	if err != nil {
		return err
	}
	// Next, we will create an empty page
	ep := NewPage(uint32(pid), P_FREE)
	// Then, we can attempt to write the contents of the empty page data directly
	// to the calculated offset
	_, err = f.file.WriteAt(ep, off)
	if err != nil {
		return err
	}
	// Don't forget to sync it up
	err = f.file.Sync()
	if err != nil {
		return err
	}
	// Finally, update the index
	// TODO: implement index and update it
	return nil
}

// ReadPage reads the page located at the logical address calculated using the
// page ID provided.
func (f *DiskManager) ReadPage(pid PageID, p Page) error {
	// Calculate the logical page offset.
	off, err := f.logicalOffset(pid)
	if err != nil {
		return err
	}
	// Read page data
	_, err = f.file.ReadAt(p, off)
	if err != nil {
		return err
	}
	return nil
}

// WritePage writes the page located at the logical address calculated using the
// page ID provided.
func (f *DiskManager) WritePage(pid PageID, p Page) error {
	// Calculate the logical page offset.
	off, err := f.logicalOffset(pid)
	if err != nil {
		return err
	}
	// Write page data
	_, err = f.file.WriteAt(p, off)
	if err != nil {
		return err
	}
	// Make sure we sync
	err = f.file.Sync()
	if err != nil {
		return err
	}
	// Finally, update the index
	// TODO: implement index and update it
	return nil
}

// Close closes the current manager instance
func (f *DiskManager) Close() error {
	// Close the underlying io
	return f.file.Close()
}
