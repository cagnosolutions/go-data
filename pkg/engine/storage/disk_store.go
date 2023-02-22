package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/cagnosolutions/go-data/pkg/engine/page"
)

const dataFilePerm = 1466

// DiskStore is a structure responsible for creating and managing access with
// the actual files stored on disk. The current disk manager instance is only
// responsible for dealing with one file at a time.
type DiskStore struct {
	sync.RWMutex
	file    *os.File
	nextPID page.PageID
	size    int64
}

// Open opens an existing disk manager instance if one exists with the same
// name, otherwise it creates a new instance and returns it along with any potential
// errors encountered.
func Open(path string) (*DiskStore, error) {
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
	// get the current file size
	fi, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	nextPageID := page.PageID(0)
	if size > 0 {
		nextPageID = page.PageID(size / page.PageSize)
	}
	// Initialize a new DiskStore instance
	fm := &DiskStore{
		file:    fp,
		nextPID: nextPageID,
		size:    size,
	}
	// Load the meta info for the DiskStore instance
	err = fm.load()
	if err != nil {
		return nil, err
	}
	// Return our instance
	return fm, nil
}

// load attempts to populate our DiskStore instance with metadata about the file.
func (s *DiskStore) load() error {
	return nil
}

// logicalOffset checks for any out of bounds errors, and returns an error if there
// is one. Otherwise, it takes a page ID and returns a logical page offset.
func (s *DiskStore) logicalOffset(pid page.PageID) (int64, error) {
	// Check to see if the requested pid falls within the set that has been distributed
	if pid > s.nextPID {
		return -1, page.ErrPageIDHasNotBeenAllocated(pid)
	}
	// We are good, so we will calculate the logical page offset.
	return int64(pid * page.PageSize), nil
}

// AllocatePage simply returns the next logical page ID that is can be written to.
func (s *DiskStore) AllocatePage() page.PageID {
	// increment and return the nextpage.PageID
	return page.PageID(atomic.SwapUint32((*uint32)(&s.nextPID), uint32(s.nextPID+1)))
}

// DeallocatePage writes zeros to the page located at the logical address
// calculated using the page ID provided.
func (s *DiskStore) DeallocatePage(pid page.PageID) error {
	// Calculate the logical page offset.
	off, err := s.logicalOffset(pid)
	if err != nil {
		return err
	}
	// Next, we will create an empty page
	ep := page.NewPage(uint32(pid), page.P_FREE)
	// Then, we can attempt to write the contents of the empty page data directly
	// to the calculated offset
	_, err = s.file.WriteAt(ep, off)
	if err != nil {
		return err
	}
	// Don't forget to sync it up
	err = s.file.Sync()
	if err != nil {
		return err
	}
	// Finally, update the index
	// TODO: implement index and update it
	return nil
}

// ReadPage reads the page located at the logical address calculated using the
// page ID provided.
func (s *DiskStore) ReadPage(pid page.PageID, p page.Page) error {
	// Calculate the logical page offset.
	off, err := s.logicalOffset(pid)
	if err != nil {
		return err
	}
	// Read page data
	_, err = s.file.ReadAt(p, off)
	if err != nil {
		return err
	}
	return nil
}

// WritePage writes the page located at the logical address calculated using the
// page ID provided.
func (s *DiskStore) WritePage(pid page.PageID, p page.Page) error {
	// Calculate the logical page offset.
	off, err := s.logicalOffset(pid)
	if err != nil {
		return err
	}
	// Write page data
	_, err = s.file.WriteAt(p, off)
	if err != nil {
		return err
	}
	// Make sure we sync
	err = s.file.Sync()
	if err != nil {
		return err
	}
	// Finally, update the index
	// TODO: implement index and update it
	return nil
}

// Close closes the current manager instance
func (s *DiskStore) Close() error {
	// close the underlying io
	return s.file.Close()
}

func (s *DiskStore) String() string {
	return s.JSON()
}

func (s *DiskStore) JSON() string {
	fi, err := s.file.Stat()
	if err != nil {
		panic(err)
	}
	info := struct {
		BasePath string `json:"base_path"`
		FileName string `json:"file_name"`
		FileSize int64  `json:"file_size"`
		NextPID  uint32 `json:"next_pid"`
		Size     int64  `json:"size"`
	}{
		BasePath: filepath.Dir(s.file.Name()),
		FileName: filepath.Base(s.file.Name()),
		FileSize: fi.Size(),
		NextPID:  s.nextPID,
		Size:     s.size,
	}
	b, err := json.MarshalIndent(&info, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
