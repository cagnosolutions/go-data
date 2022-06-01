package pager

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

// tempDiskManager is the disk implementation of DiskManager
type tempDiskManager struct {
	db         *os.File
	fileName   string
	nextPageID pageID
	numWrites  uint64
	size       int64
}

// newTempDiskManager returns a DiskManager instance
func newTempDiskManager(dbFilename string) *tempDiskManager {
	file, err := makeOrOpenFile(dbFilename)
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
	nPages := fileSize / szPg

	nextPageID := pageID(0)
	if nPages > 0 {
		nextPageID = pageID(int32(nPages + 1))
	}
	return &tempDiskManager{
		db:         file,
		fileName:   dbFilename,
		nextPageID: nextPageID,
		numWrites:  0,
		size:       fileSize,
	}
}

// ShutDown closes of the database file
func (d *tempDiskManager) close() {
	d.db.Close()
}

// write writes a page to the database file
func (d *tempDiskManager) write(pid pageID, p page) error {
	offset := int64(pid * szPg)
	_, err := d.db.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	bytesWritten, err := d.db.Write(p)
	if err != nil {
		return err
	}
	if bytesWritten != szPg {
		return errors.New("bytes written not equal to page size")
	}
	if offset >= d.size {
		d.size = offset + int64(bytesWritten)
	}
	err = d.db.Sync()
	if err != nil {
		return err
	}
	return nil
}

// read a page from the database file
func (d *tempDiskManager) read(pid pageID, p page) error {
	offset := int64(pid * szPg)
	fileInfo, err := d.db.Stat()
	if err != nil {
		return errors.New("file info error")
	}
	if offset > fileInfo.Size() {
		return errors.New("I/O error past end of file")
	}
	_, err = d.db.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	bytesRead, err := d.db.Read(p)
	if err != nil {
		return errors.New("I/O error while reading")
	}
	if bytesRead < szPg {
		for i := 0; i < szPg; i++ {
			p[i] = 0
		}
	}
	return nil
}

//  AllocatePage allocates a new page
//  For now just keep an increasing counter
func (d *tempDiskManager) allocate() pageID {
	ret := d.nextPageID
	d.nextPageID++
	return ret
}

// DeallocatePage deallocates page
// Need bitmap in header page for tracking pages
// This does not actually need to do anything for now.
func (d *tempDiskManager) deallocate(pid pageID) {
}

// GetNumWrites returns the number of disk writes
func (d *tempDiskManager) GetNumWrites() uint64 {
	return d.numWrites
}

// Size returns the size of the file in disk
func (d *tempDiskManager) Size() int64 {
	return d.size
}

func makeOrOpenFile(path string) (*os.File, error) {
	// sanitize path
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// split path
	dir, name := filepath.Split(filepath.ToSlash(path))
	// init files and dirs
	var fp *os.File
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// create dir
		err = os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			return nil, err
		}
		// create file
		fp, err = os.Create(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		// close file
		err = fp.Close()
		if err != nil {
			return nil, err
		}
	}
	// open existing file
	fp, err = os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	// return file and nil error
	return fp, nil
}
