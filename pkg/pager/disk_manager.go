package pager

import (
	"io"
	"os"
	"sync/atomic"
)

// diskManager is a manager storageManager
type diskManager struct {
	file       *os.File
	fileName   string
	nextPageID pageID
	fileSize   int64
}

// newDMan initializes and returns a new diskManager instance.
func newDiskManager(dbFilePath string) *diskManager {
	// check to see if a file exists (if none, create)
	fp, err := fileOpenOrCreate(dbFilePath)
	if err != nil {
		panic(err)
	}
	fi, err := os.Stat(fp.Name())
	if err != nil {
		panic(err)
	}
	// check the fileSize of the file
	size := fi.Size()
	// init and return a new *diskManager
	return &diskManager{
		file:       fp,
		fileName:   fp.Name(),
		nextPageID: pageID(size / szPg),
		fileSize:   size,
	}
}

// allocatePage returns, then increments the ID or offset of the next entry.
func (dm *diskManager) allocatePage() pageID {
	return atomic.SwapUint32(&dm.nextPageID, dm.nextPageID+1)
	// next := dm.nextPageID
	// dm.nextPageID++
	// return next
}

// deallocatePage wipes the entry in the location matching the provided ID
// or offset. If out of bounds, it will return an error.
func (dm *diskManager) deallocatePage(pid pageID) error {
	// calculate the logical page offset should be.
	offset := int64(pid * szPg)
	if offset < 0 {
		return ErrOffsetOutOfBounds
	}
	// Create an empty page
	ep := newEmptyPage(pid)
	// Next, we can attempt to write the contents of an empty page data
	// directly to the calculated offset.
	n, err := dm.file.WriteAt(ep, offset)
	if err != nil {
		return err
	}
	// Check to ensure that we actually wrote the contents of a full page.
	if n != szPg {
		return ErrPartialPageWrite
	}
	// Update the file fileSize if necessary.
	if offset >= dm.fileSize {
		dm.fileSize = offset + int64(n)
	}
	// Before we are finished, we should call sync.
	err = dm.file.Sync()
	if err != nil {
		return err
	}
	// Finally, we can return a nil error because everything is good.
	return nil
}

// readPage attempts to read the entry. It uses the ID provided to calculate
// the logical offset where the entry should be located and will attempt to
// read the entry contents from that location. Any errors encountered will
// be returned immediately.
func (dm *diskManager) readPage(pid pageID, p page) error {
	// First we do a little error checking, and calculate what the page
	// offset is supposed to be.
	offset := int64(pid * szPg)
	// Next, we can attempt to read the contents of the page data
	// directly from the calculated offset. Using ReadAt makes one
	// syscall, vs using Seek+Read which calls syscall twice.
	n, err := dm.file.ReadAt(p, offset)
	if err != nil {
		return err
	}
	// Check to ensure that we actually read the contents of a full page.
	if n < szPg {
		// **It should be noted that we could also alternatively choose
		// to pad out the remaining byte of the page right in here if
		// we want to. For now though, we will error out. I feel that
		// we will always be wanting to read and write full pages so
		// in that case an error is a valid response.
		return ErrPartialPageRead
	}
	// Finally, we can return a nil error
	return nil
}

// writePage attempts to write an entry. It uses the ID provided to
// calculate the logical offset where the entry should be inserted
// and will attempt to write the contents of the provided entry to
// that location. Any errors encountered will be returned immediately.
func (dm *diskManager) writePage(pid pageID, p page) error {
	// calculate the logical page offset should be.
	offset := int64(pid * szPg)
	if offset < 0 {
		return ErrOffsetOutOfBounds
	}
	// Next, we can attempt to write the contents of the page data
	// directly to the calculated offset.
	// Note: Using WriteAt makes one syscall, vs using Seek+Write
	//  which calls syscall twice. We should reserve the Seek+Write
	//  pattern only if we are dealing with append only writing type
	//  of instance.
	n, err := dm.file.WriteAt(p, offset)
	if err != nil {
		return err
	}
	// Check to ensure that we actually wrote the contents of a full page.
	if n != szPg {
		return ErrPartialPageWrite
	}
	// Update the file fileSize if necessary.
	if offset >= dm.fileSize {
		dm.fileSize = offset + int64(n)
	}
	// Before we are finished, we should call sync.
	err = dm.file.Sync()
	if err != nil {
		return err
	}
	// Finally, we can return a nil error because everything is good.
	return nil
}

// close attempts to finalize and close any open streams. Any errors
// encountered will be returned immediately.
func (dm *diskManager) close() error {
	err := dm.file.Close()
	if err != nil {
		return err
	}
	return nil
}

// size returns the number of bytes the storage manager is current using.
func (dm *diskManager) size() int {
	fi, err := dm.file.Stat()
	if err != nil {
		panic(err)
	}
	return int(fi.Size())
}

// getFreePages searches through the file looking for all the pages
// that have been deallocated and returns a set of page ID's with
// any of the deallocated pages.
func (dm *diskManager) getFreePages() []pageID {
	// Check to ensure the file actually contains some kind of data.
	if dm.fileSize < 1 {
		return nil
	}
	// Start at the beginning of the file, checking each page status
	// and build a list of free page ID's.
	var pid int64
	buf := make([]byte, 2)
	var free []pageID
	for {
		_, err := dm.file.ReadAt(buf, (pid*szPg)+4)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil
		}
		magic := bin.Uint16(buf)
		if magic&stFree > 0 {
			free = append(free, pageID(pid))
		}
		pid++
	}
	// Check and return our set of free / deallocated page ID's.
	if len(free) < 1 {
		return nil
	}
	return free
}

func NewDiskManager(path string) StorageManager {
	return newDiskManager(path)
}

func (dm *diskManager) Allocate() PageID {
	return dm.allocatePage()
}

func (dm *diskManager) Deallocate(pid PageID) error {
	return dm.deallocatePage(pid)
}

func (dm *diskManager) ReadPage(pid PageID, pg Page) error {
	return dm.readPage(pid, pg)
}

func (dm *diskManager) WritePage(pid PageID, pg Page) error {
	return dm.writePage(pid, pg)
}

func (dm *diskManager) Close() error {
	return dm.close()
}

func (dm *diskManager) Size() int {
	return dm.size()
}
