package disk

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

// FileDiskManager is the disk implementation of DiskManager
type FileDiskManager struct {
	db         *os.File
	fileName   string
	nextPageID page.PageID
	numWrites  uint64
	size       int64
}

// NewFileDiskManager returns a DiskManager instance
func NewFileDiskManager(dbFilename string) DiskManager {
	file, err := os.OpenFile(dbFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("can't open db file")
		return nil
	}

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalln("file info error")
		return nil
	}

	fileSize := fileInfo.Size()
	nPages := fileSize / page.PageSize

	nextPageID := page.PageID(0)
	if nPages > 0 {
		nextPageID = page.PageID(int32(nPages + 1))
	}

	return &FileDiskManager{
		db:         file,
		fileName:   dbFilename,
		nextPageID: nextPageID,
		numWrites:  0,
		size:       fileSize,
	}
}

// ShutDown closes of the database file
func (d *FileDiskManager) ShutDown() {
	d.db.Close()
}

// WritePage writes a page to the database file
func (d *FileDiskManager) WritePage(pageId page.PageID, pageData []byte) error {
	offset := int64(pageId * page.PageSize)
	d.db.Seek(offset, io.SeekStart)
	bytesWritten, err := d.db.Write(pageData)
	if err != nil {
		return err
	}

	if bytesWritten != page.PageSize {
		return errors.New("bytes written not equals page size")
	}

	if offset >= d.size {
		d.size = offset + int64(bytesWritten)
	}

	d.db.Sync()
	return nil
}

// ReadPage reads a page from the database file
func (d *FileDiskManager) ReadPage(pageID page.PageID, pageData []byte) error {
	offset := int64(pageID * page.PageSize)

	fileInfo, err := d.db.Stat()
	if err != nil {
		return errors.New("file info error")
	}

	if offset > fileInfo.Size() {
		return errors.New("I/O error past end of file")
	}

	d.db.Seek(offset, io.SeekStart)

	bytesRead, err := d.db.Read(pageData)
	if err != nil {
		return errors.New("I/O error while reading")
	}

	if bytesRead < page.PageSize {
		for i := 0; i < page.PageSize; i++ {
			pageData[i] = 0
		}
	}
	return nil
}

// AllocatePage allocates a new page; for now just keep an increasing counter
func (d *FileDiskManager) AllocatePage() page.PageID {
	ret := d.nextPageID
	d.nextPageID++
	return ret
}

// DeallocatePage deallocates page
// Need bitmap in header page for tracking pages
// This does not actually need to do anything for now.
func (d *FileDiskManager) DeallocatePage(pageID page.PageID) {
}

// GetNumWrites returns the number of disk writes
func (d *FileDiskManager) GetNumWrites() uint64 {
	return d.numWrites
}

// Size returns the size of the file in disk
func (d *FileDiskManager) Size() int64 {
	return d.size
}
