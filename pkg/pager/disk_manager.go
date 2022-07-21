package pager

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
)

var headerBufPool = sync.Pool{
	New: func() any {
		return make([]byte, fileHeaderSize)
	},
}

func clearHeaderBuf(hb []byte) {
	for i := range hb {
		hb[i] = 0x00
	}
}

const (
	// minPageSize is the minimum page size allowed. It is small enough to fit 16 min
	// sized virtual pages into one physical 4KB address space.
	minPageSize = 255

	// maxPageSize is the maximum page size allowed. It is lage enough to map 16 physical
	// 4KB page address into one 64KB virtual address space.
	maxPageSize = 65535

	// fileHeaderSize is the space reserved for file header information. There may be
	// several file header types, but each file has a statically reserved header size.
	// The file header size does not have to be full.
	fileHeaderSize = 24

	// maxSegmentSize is the MAX FILE SIZE, IT'S GOOD, KEEP IT!
	maxSegmentSize = 16 << 20
)

// segment > extent > page

/*  		File header structure byte index, field name and field bits overview
			+------+------+------+------+------+------+------+------+
 byte index	|  00  |  01  |  02  |  03  |   04   |   05   |   06   |   07   | byte index
			+------+------+------+--------+--------+--------+--------+--------+
 field name	|  lock  | vers | kind |    page_size    |    page_count   | field
			+------+------+-----------------+-----------------+-----------------+
 field bits |   u8   |   u8   |   u8   |       u16       |       u16       | bits
			+------+------+-----------------+-----------------+-----------------+

			file_lock=u8, magic=u8, file_kind=u16, page_size=u16, page_count=u16
*/

// File header binary offsets
const (
	offSign uint64 = 0  // 0-8   (8 bytes)
	offVers uint16 = 8  // 8-10  (2 bytes)
	offPgSz uint16 = 10 // 10-12 (2 bytes)
	offPgCt uint16 = 12 // 12-14 (2 bytes)
	offRes1 uint16 = 14 // 14-16 (2 bytes)
	offRes2 uint32 = 16 // 16-20 (4 bytes)
	offCRC  uint32 = 20 // 20-24 (4 bytes)

	fileSignature uint64 = 0xDEADBEEF
	fileVersion   uint16 = 0x0001
)

type FileHeader interface {
	size() uint8
	kind() uint8
}

type fileheader struct {
	signature uint64 // data file signature
	version   uint16 // data file version number
	pageSize  uint16 // data file page size
	pageCount uint16 // data file size in pages
	res1      uint16 // reserved for expansion
	res2      uint32 // reserved for expansion
	checksum  uint32 // data file header checksum
}

func getChecksum(signature uint64, version, pageSize, pageCount uint16) uint32 {
	// little endian encoding
	return checksum(
		[]byte{
			// signature
			byte(signature),
			byte(signature >> 8),
			byte(signature >> 16),
			byte(signature >> 24),
			byte(signature >> 32),
			byte(signature >> 40),
			byte(signature >> 48),
			byte(signature >> 56),
			// version
			byte(version),
			byte(version >> 8),
			// pageSize
			byte(pageSize),
			byte(pageSize >> 8),
			// pageCount
			byte(pageCount),
			byte(pageCount >> 8),
		},
	)
}

func (dm *diskManager) writeHeader() error {
	// get out byte slice buffer from the pool
	hb := headerBufPool.Get().([]byte)
	defer headerBufPool.Put(hb)
	// serialize the fileHeader into our buffer
	bin.PutUint64(hb[offSign:offSign+8], dm.header.signature)
	bin.PutUint16(hb[offVers:offVers+2], dm.header.version)
	bin.PutUint16(hb[offPgSz:offPgSz+2], dm.header.pageSize)
	bin.PutUint16(hb[offPgCt:offPgCt+2], dm.header.pageCount)
	bin.PutUint16(hb[offRes1:offRes1+2], dm.header.res1)
	bin.PutUint32(hb[offRes2:offRes2+4], dm.header.res2)
	bin.PutUint32(hb[offCRC:offCRC+4], dm.header.checksum)
	// test the checksum (before writing)
	sum := checksum(hb[offSign : offPgCt+2])
	if sum != dm.header.checksum {
		return ErrCRCFileHeader
	}
	// write the file header from the buffer to disk
	_, err := dm.file.WriteAt(hb, 0)
	if err != nil {
		return err
	}
	return nil
}

func (dm *diskManager) readHeader() error {
	// get out byte slice buffer from the pool
	hb := headerBufPool.Get().([]byte)
	defer headerBufPool.Put(hb)
	// read the file header into the buffer from the disk
	_, err := dm.file.ReadAt(hb, 0)
	if err != nil {
		return err
	}
	// test the checksum (before de-serializing)
	sum := checksum(hb[offSign : offPgCt+2])
	if sum != bin.Uint32(hb[offCRC:offCRC+4]) {
		return ErrCRCFileHeader
	}
	// de-serialize the buffer into the header struct
	dm.header.signature = bin.Uint64(hb[offSign : offSign+8])
	dm.header.version = bin.Uint16(hb[offVers : offVers+2])
	dm.header.pageSize = bin.Uint16(hb[offPgSz : offPgSz+2])
	dm.header.pageCount = bin.Uint16(hb[offPgCt : offPgCt+2])
	dm.header.res1 = bin.Uint16(hb[offRes1 : offRes1+2])
	dm.header.res2 = bin.Uint32(hb[offRes2 : offRes2+4])
	dm.header.checksum = bin.Uint32(hb[offCRC : offCRC+4])
	return nil
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
	header     *fileHeader
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
		header: &fileHeader{
			signature: fileSignature,
			version:   fileVersion,
			pageSize:  pageSize,
			pageCount: 0,
			res1:      0,
			res2:      0,
			checksum:  getChecksum(fileSignature, fileVersion, pageSize, pageCount),
		},
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
	dm.header.pageCount++
	err := dm.writeHeader()
	if err != nil {
		panic(ErrWriteFileHeader)
	}
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
	// Update the file header if necessary.
	if pid > uint32(dm.header.pageCount) {
		dm.header.pageCount++
		err = dm.writeHeader()
		if err != nil {
			return err
		}
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
