package pager

import (
	"io"
	"os"
	"sync/atomic"
)

// Sizes for file header
const (
	_szMinRec = 8       // minimum record size
	_szMaxRec = 2048    // maximum record size
	_szHd     = 16      // fileSize of page header (in bytes)
	_szSl     = 6       // fileSize of slot index (in bytes)
	_szPg     = 4 << 10 // fileSize of page (default)
)

const (
	minPageSize = 0
	maxPageSize = 0

	minPageCount = 0
	maxPageCount = 0

	defaultPageSize  = 0
	defaultPageCount = 0
)

// Binary offsets for file header
const (
	offSign        uint64 = 0  // 0-8   (8 bytes)
	offVers        uint16 = 8  // 8-10  (2 bytes)
	offPgSz        uint16 = 10 // 10-12 (2 bytes)
	offPgCt        uint16 = 12 // 12-14 (2 bytes)
	offRes1        uint16 = 14 // 14-16 (2 bytes)
	offRes2        uint32 = 16 // 16-20 (4 bytes)
	offCRC         uint32 = 20 // 20-24 (4 bytes)
	fileHeaderSize        = 24
)

type fileHeader struct {
	signature uint64 // data file signature
	version   uint16 // data file version number
	pageSize  uint16 // data file page size
	pageCount uint16 // data file size in pages
	res1      uint16 // reserved for expansion
	res2      uint32 // reserved for expansion
	checksum  uint32 // data file header checksum
}

func (dm *diskManager) setHeader(fh *fileHeader) {
	p := make([]byte, fileHeaderSize)
	bin.PutUint64(p[offSign:offSign+8], fh.signature)
	bin.PutUint16(p[offVers:offVers+2], fh.version)
	bin.PutUint16(p[offPgSz:offPgSz+2], fh.pageSize)
	bin.PutUint16(p[offPgCt:offPgCt+2], fh.pageCount)
	bin.PutUint16(p[offRes1:offRes1+2], fh.res1)
	bin.PutUint32(p[offRes2:offRes2+4], fh.res2)
	bin.PutUint32(p[offCRC:offCRC+4], fh.checksum)
	_, err := dm.file.WriteAt(p, 0)
	if err != nil {
		panic("error writing data file header")
	}
}

func (dm *diskManager) getHeader() *fileHeader {
	p := make([]byte, fileHeaderSize)
	_, err := dm.file.ReadAt(p, 0)
	if err != nil {
		panic("error reading data file header")
	}
	return &fileHeader{
		signature: bin.Uint64(p[offSign : offSign+8]),
		version:   bin.Uint16(p[offVers : offVers+2]),
		pageSize:  bin.Uint16(p[offPgSz : offPgSz+2]),
		pageCount: bin.Uint16(p[offPgCt : offPgCt+2]),
		res1:      bin.Uint16(p[offRes1 : offRes1+2]),
		res2:      bin.Uint32(p[offRes2 : offRes2+4]),
		checksum:  bin.Uint32(p[offCRC : offCRC+4]),
	}
}

func (m *fileHeader) read(p []byte) {
	if len(p) < 8 {
		panic("cannot read meta info, buffer is too small")
	}
	m.pageSize = bin.Uint16(p[0:])
	m.pageCount = bin.Uint16(p[4:])
}

func (m fileHeader) write(p []byte) {
	if len(p) < 8 {
		panic("cannot write meta info, buffer is too small")
	}
	bin.PutUint16(p[0:], m.pageSize)
	bin.PutUint16(p[4:], m.pageCount)
}

const (
	dbFileSuffix = `.db`
	dbMetaSuffix = `.meta`
)

// diskManager is a disk storageManager
type diskManager struct {
	file       *os.File
	filePath   string
	nextPageID pageID
	pageSize   uint16
	fileSize   int64
}

// newDiskManagerSize initializes and returns a new diskManager instance using the specified
// pageSize the default maxPageCount.
func newDiskManager(filePath string, pageSize uint16) (*diskManager, error) {
	return newDiskManagerSize(filePath, pageSize, defaultPageCount)
}

// newDiskManagerSize initializes and returns a new diskManager instance using the specified
// pageSize and maxPageCount.
func newDiskManagerSize(filePath string, pageSize uint16, pageCount uint16) (*diskManager, error) {
	// sanitize the provided path, and trim the provided file suffix (if it has any)
	path, err := pathCleanAndTrimSUffix(filePath)
	if err != nil {
		return nil, err
	}
	// check to see if a file exists (if none, create)
	fd, err := fileOpenOrCreate(path + dbFileSuffix)
	if err != nil {
		return nil, err
	}
	// stat db file to get the size
	fi, err := os.Stat(fd.Name())
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	// initialize a new *diskManager instance
	dm := &diskManager{
		file:       fd,
		filePath:   fd.Name(),
		nextPageID: pageID(size / int64(pageSize)),
		pageSize:   pageSize,
		fileSize:   size,
	}
	// check for metafile and read/write
	if !dm.hasMeta(path) {
		// no metafile, so let's write a new one
		err = dm.writeMeta(path, pageSize, pageCount)
		if err != nil && err != ErrMetaFileExists {
			return nil, err
		}
	} else {
		// there must be a metafile, so let's check it
		err = dm.checkMeta(path, pageSize, pageCount)
		if err != nil {
			return nil, err
		}
	}
	// everything looks good, let's return the disk manager
	return dm, nil
}

func (dm *diskManager) hasMeta(name string) bool {
	// stat meta file to get the size
	_, err := os.Stat(name + dbMetaSuffix)
	return os.IsExist(err)
}

func (dm *diskManager) writeMeta(name string, pageSize, pageCount uint16) error {
	// stat meta file to get the size
	fi, err := os.Stat(name + dbMetaSuffix)
	if os.IsExist(err) && fi.Size() > 0 {
		// meta file already exists
		return ErrMetaFileExists
	}
	// meta file does not yet exist, lets write it!
	buf := make([]byte, 8)
	mi := fileHeader{
		pageSize:  pageSize,
		pageCount: pageCount,
	}
	// write meta file contents into the buffer (which will be saved as a file)
	mi.write(buf)
	err = os.WriteFile(name+dbMetaSuffix, buf, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (dm *diskManager) checkMeta(name string, pageSize, pageCount uint16) error {
	// stat meta file to get the size
	fi, err := os.Stat(name + dbMetaSuffix)
	if os.IsNotExist(err) || fi.Size() == 0 {
		// meta file already exists
		return ErrMetaFileNotExists
	}
	// meta file does indeed exist we must check that it is correct
	buf, err := os.ReadFile(name + dbMetaSuffix)
	if err != nil {
		return err
	}
	var mi fileHeader
	mi.read(buf)
	if mi.pageSize != pageSize || mi.pageCount != pageCount {
		return ErrMetaInfoMismatch
	}
	return nil
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
	offset := int64(pid * uint32(dm.pageSize))
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
	if n != int(dm.pageSize) {
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

func (dm *diskManager) hasPage(pid pageID) bool {
	// First we do a little error checking, and calculate what the page
	// offset is supposed to be.
	offset := int64(pid * uint32(dm.pageSize))
	// stat the actual file to ensure it has the potential to contain
	// the page that has been requested
	fi, err := os.Stat(dm.filePath)
	if os.IsNotExist(err) {
		return false
	}
	// check to make sure the page offset is within the file bounds
	if offset >= fi.Size() {
		return false
	}
	// the page is most likely there
	return true
}

// readPage attempts to read the entry. It uses the ID provided to calculate
// the logical offset where the entry should be located and will attempt to
// read the entry contents from that location. Any errors encountered will
// be returned immediately.
func (dm *diskManager) readPage(pid pageID, p page) error {
	// First we do a little error checking, and calculate what the page
	// offset is supposed to be.
	offset := int64(pid * uint32(dm.pageSize))
	// Next, we can attempt to read the contents of the page data
	// directly from the calculated offset. Using ReadAt makes one
	// syscall, vs using Seek+Read which calls syscall twice.
	n, err := dm.file.ReadAt(p, offset)
	if err != nil {
		return err
	}
	// Check to ensure that we actually read the contents of a full page.
	if n < int(dm.pageSize) {
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
	offset := int64(pid * uint32(dm.pageSize))
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
	if n != int(dm.pageSize) {
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

// size returns the number of bytes the storage disk is current using.
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
		_, err := dm.file.ReadAt(buf, (pid*int64(dm.pageSize))+4)
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
