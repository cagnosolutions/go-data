package pager

import (
	"io"
	"os"
)

// dMan is a disk manager
type dMan struct {
	file       *os.File
	fileName   string
	nextPageID pageID
	size       int64
}

// newDMan initializes and returns a new dMan instance.
func newDMan(dbFilePath string) *dMan {
	// check to see if a file exists (if none, create)
	fp, err := fileOpenOrCreate(dbFilePath)
	if err != nil {
		panic(err)
	}
	fi, err := os.Stat(fp.Name())
	if err != nil {
		panic(err)
	}
	// check the size of the file
	size := fi.Size()
	// init and return a new *dMan
	return &dMan{
		file:       fp,
		fileName:   fp.Name(),
		nextPageID: pageID(size / szPg),
		size:       size,
	}
}

// allocate returns the next pageID
func (dm *dMan) allocatePage() pageID {
	next := dm.nextPageID
	dm.nextPageID++
	return next
}

// getFreePages searches through the file looking for all the pages
// that have been deallocated and returns a set of page ID's with
// any of the deallocated pages.
func (dm *dMan) getFreePages() []pageID {
	// Check to ensure the file actually contains some kind of data.
	if dm.size < 1 {
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

// deallocatePage wipes the page matching the supplied page ID. If the
// page ID is out of the bounds of the file, this call is ignored.
func (dm *dMan) deallocatePage(pid pageID) error {
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
	// Update the file size if necessary.
	if offset >= dm.size {
		dm.size = offset + int64(n)
	}
	// Before we are finished, we should call sync.
	err = dm.file.Sync()
	if err != nil {
		return err
	}
	// Finally, we can return a nil error because everything is good.
	return nil
}

// writePage attempts to write a page to the underlying storage. It uses
// the pageID provided to calculate the logical page offset and will
// attempt to write the contents of the provided page to that offset.
// Any errors encountered while calculating the logical page offset or
// while writing will be returned.
func (dm *dMan) writePage(pid pageID, p page) error {
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
	// Update the file size if necessary.
	if offset >= dm.size {
		dm.size = offset + int64(n)
	}
	// Before we are finished, we should call sync.
	err = dm.file.Sync()
	if err != nil {
		return err
	}
	// Finally, we can return a nil error because everything is good.
	return nil
}

// readPage tries to read the page data from the disk. It uses the provided page ID
// to calculate the logical offset of the disk resident page. It attempts to read
// the data into the provided page. If no errors are encountered, the function will
// return a nil error.
func (dm *dMan) readPage(pid pageID, p page) error {
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

// close calls sync and close on the underlying file descriptor.
func (dm *dMan) close() error {
	err := dm.file.Close()
	if err != nil {
		return err
	}
	return nil
}

// getFileSize returns the current file size
func (dm *dMan) fileSize() int64 {
	fi, err := dm.file.Stat()
	if err != nil {
		panic(err)
	}
	return fi.Size()
}
