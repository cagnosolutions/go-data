package dbms

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"unsafe"
)

var bin = binary.LittleEndian

const (
	dataFileHeaderSize = int(unsafe.Sizeof(DataFileHeader{}))
	offFileID          = 0
	offNamespace       = 4
	lenNamespace       = 32
	offFileName        = 36
	lenFileName        = 32
	offNextPage        = 68
	offPageCount       = 72
	offFreeCount       = 76
	offReadCount       = 80
	offWriteCount      = 88
	offIndex           = 96
)

// DataFileHeader is the header for the data file
type DataFileHeader struct {
	FileID     uint32   // file id
	Namespace  [32]byte // namespace
	FileName   [32]byte // base file name
	NextPage   uint32   // next page offset
	PageCount  uint32   // number of pages allocated, used or unused
	FreeCount  uint32   // number of unused pages
	ReadCount  uint64   // number of read calls performed
	WriteCount uint64   // number of write calls performed
	Index      bitset   // bitset index
}

// ReadDataFileHeader reads from the supplied file pointer and returns a *DataFileHeader
// otherwise returning an error.
func ReadDataFileHeader(fp *os.File) (*DataFileHeader, error) {
	// Ensure the supplied file pointer is not nil
	if fp == nil {
		return nil, os.ErrInvalid
	}
	// Create a buffer to read the header data into
	buf := make([]byte, dataFileHeaderSize)
	// Read the data file header using the provided file pointer
	_, err := fp.ReadAt(buf, 0)
	if err != nil {
		return nil, err
	}
	// Initialize and assemble a data file header to return
	fh := new(DataFileHeader)
	fh.FileID = bin.Uint32(buf[offFileID : offFileID+4])
	copy(fh.Namespace[:], buf[offNamespace:offNamespace+lenNamespace])
	copy(fh.FileName[:], buf[offFileName:offFileName+lenFileName])
	fh.NextPage = bin.Uint32(buf[offNextPage : offNextPage+4])
	fh.PageCount = bin.Uint32(buf[offPageCount : offPageCount+4])
	fh.FreeCount = bin.Uint32(buf[offFreeCount : offFreeCount+4])
	fh.ReadCount = bin.Uint64(buf[offReadCount : offReadCount+8])
	fh.WriteCount = bin.Uint64(buf[offWriteCount : offWriteCount+8])
	for i, off := 0, offIndex; off < dataFileHeaderSize; i, off = i+1, off+8 {
		fh.Index[i] = bin.Uint64(buf[off : off+8])
	}
	// Return new data file header
	return fh, nil
}

// WriteDataFileHeader writes the supplied *DataFileHeader to the supplied file pointer
// and returns any errors.
func WriteDataFileHeader(fp *os.File, fh *DataFileHeader) error {
	// Ensure the supplied file pointer (and header) is not nil
	if fp == nil || fh == nil {
		return os.ErrInvalid
	}
	// Create a buffer to encode the header data into
	buf := make([]byte, dataFileHeaderSize)
	// Encode the supplied header into the buffer before writing
	bin.PutUint32(buf[offFileID:offFileID+4], fh.FileID)
	copy(buf[offNamespace:offNamespace+lenNamespace], fh.Namespace[:])
	copy(buf[offFileName:offFileName+lenFileName], fh.FileName[:])
	bin.PutUint32(buf[offNextPage:offNextPage+4], fh.NextPage)
	bin.PutUint32(buf[offPageCount:offPageCount+4], fh.PageCount)
	bin.PutUint32(buf[offFreeCount:offFreeCount+4], fh.FreeCount)
	bin.PutUint64(buf[offReadCount:offReadCount+8], fh.ReadCount)
	bin.PutUint64(buf[offWriteCount:offWriteCount+8], fh.WriteCount)
	for i, off := 0, offIndex; off < dataFileHeaderSize; i, off = i+1, off+8 {
		fh.Index[i] = bin.Uint64(buf[off : off+8])
	}
	// Write the data file header using the provided file pointer
	_, err := fp.WriteAt(buf, 0)
	if err != nil {
		return err
	}
	// Sync the file write
	err = fp.Sync()
	if err != nil {
		return err
	}
	// Return a nil error
	return nil
}

const (
	dataFileMax  = 16 << 20
	dataFilePage = 16 << 10
	dataFilePerm = 1466 // 0=none, 1=exec, 2=write, 3=exec+write, 4=read, 5=exec+read, 6=write+read, 7=write+exec+read
)

// makeFileName takes an uint32 ID and returns a file name using it.
func makeFileName(id uint32) string {
	return fmt.Sprintf("dat-%.8x.seg", id)
}

// getFileID takes a filename and returns an uint32 file ID
func getFileID(name string) uint32 {
	id, err := strconv.ParseUint(name[4:len(name)-4], 16, 32)
	if err != nil {
		panic(err)
	}
	return uint32(id)
}

// DataFile is an *os.File wrapper that contains some extra metadata and
// reads and writes data on page aligned blocks.
type DataFile struct {
	latch           sync.Mutex
	*DataFileHeader          // data file Header
	fp              *os.File // file pointer
}

// OpenDataFile opens and returns a new DataFile instance. If the data file
// requested does not exist, it is created and returned, along with any
// directories that may also need to be created.
func OpenDataFile(namespace string, id uint32) (*DataFile, error) {
	// Clean path
	path, err := filepath.Abs(filepath.ToSlash(namespace))
	if err != nil {
		panic("pathClean: " + err.Error())
	}
	full := filepath.Join(path, makeFileName(id))
	var fp *os.File
	_, err = os.Stat(full)
	if os.IsNotExist(err) {
		// Touch any directories and/or file
		err = os.MkdirAll(filepath.Dir(full), os.ModeDir|dataFilePerm)
		if err != nil {
			return nil, err
		}
		// we are creating a new file, so we need to create a new file name
		fp, err = os.OpenFile(full, os.O_CREATE|os.O_TRUNC, dataFilePerm)
		if err != nil {
			return nil, err
		}
		err = fp.Close()
		if err != nil {
			return nil, err
		}
	}
	// Open file at the fully cleaned path
	fp, err = os.OpenFile(full, os.O_RDWR|os.O_SYNC, dataFilePerm)
	if err != nil {
		return nil, err
	}
	// Create a new DataFile instance
	df := &DataFile{
		DataFileHeader: new(DataFileHeader),
		fp:             fp,
	}
	// Fill out some header information out in case this is our first
	// go around, it doesn't really matter since the header will be
	// overwritten anyway if there is already a header that exists.
	df.FileID = id
	copy(df.Namespace[:], namespace)
	copy(df.FileName[:], makeFileName(id))
	df.NextPage = 1
	df.Index = *new(bitset)
	// Load our file header, and index and then return
	err = df.load()
	if err != nil {
		return nil, err
	}
	return df, nil
}

// load attempts to read the file pointer opened and given to the DataFile
// and read the header and bitset contents. It returns any errors encountered.
func (df *DataFile) load() error {
	// Latch down our file
	df.latch.Lock()
	defer df.latch.Unlock()
	// get the current size of the file
	fi, err := df.fp.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		// This is a brand-new file, there is no header to load yet. We
		// already have some initial values for the header that we set in
		// the caller, so we can simply write the header so next time we
		// go to open the file, there is something to read.
		err = WriteDataFileHeader(df.fp, df.DataFileHeader)
		if err != nil {
			return err
		}
		// Return nil
		return nil
	}
	// Otherwise, this is not our first time, and we need to read our header
	fh, err := ReadDataFileHeader(df.fp)
	if err != nil {
		return err
	}
	// Assign our file header to our data file
	df.DataFileHeader = fh
	// Return nil
	return nil
}

// ReadAt reads a page sized slice from the DataFile starting at page offset
// off. It returns an error, if any. At end of file, that error is io.EOF.
func (df *DataFile) ReadAt(p []byte, off int64) (int, error) {
	// Perform actual read
	n, err := df.fp.ReadAt(p, off)
	if err != nil {
		return -1, err
	}
	// Update read counter
	df.ReadCount++
	// Return
	return n, nil
}

// WriteAt writes a page to the DataFile starting at byte offset off. It
// returns an error, if any. If file was opened with the O_APPEND flag,
// WriteAt returns an error.
func (df *DataFile) WriteAt(p []byte, off int64) (int, error) {
	// Perform actual write
	n, err := df.fp.WriteAt(p, off)
	if err != nil {
		return -1, err
	}
	// Sync the data
	err = df.fp.Sync()
	if err != nil {
		return -1, err
	}
	// Update the bitset index
	df.Index.set(uint64(off))
	// Update write counter
	df.WriteCount++
	// Return
	return n, nil
}

// RemoveAt writes a blank page to the DataFile starting at byte offset off.
func (df *DataFile) RemoveAt(off int64) (int, error) {
	// Perform actual write
	n, err := df.fp.WriteAt(make([]byte, dataFilePage), off)
	if err != nil {
		return -1, err
	}
	// Sync the data
	err = df.fp.Sync()
	if err != nil {
		return -1, err
	}
	// Update the bitset index
	df.Index.unset(uint64(off))
	// Update write counter
	df.WriteCount++
	// Return
	return n, nil
}

// NextOff returns the next available offset for writing, either sequentially
// or a removed page space. If no space can be found, an error is returned.
func (df *DataFile) NextOff() (uint32, error) {
	// check some stuff
	next := atomic.SwapUint32(&df.NextPage, df.NextPage+1)
	return next, nil
}

// Reads returns the number read operations the DataFile has performed.
func (df *DataFile) Reads() int {
	return int(df.ReadCount)
}

// Writes returns the number write operations the DataFile has performed.
func (df *DataFile) Writes() int {
	return int(df.WriteCount)
}

// Size returns the current size of the DataFile
func (df *DataFile) Size() int64 {
	return int64(df.PageCount * dataFilePage)
}

// Close closes the current DataFile
func (df *DataFile) Close() error {
	return df.fp.Close()
}

// bitset constants
const (
	bsWS     = 64 // choices 64,32,16,8
	bsL2     = 6  // choices 6 for 64, 5 for 32, 4 for 16 , and 3 for 8
	bsSize   = 16 // fixed size of the bitset
	bsAllOn  = 0xffffffffffffffff
	bsAllOff = 0x0000000000000000
)

// bitset is a fixed sized bitset
type bitset [bsSize]uint64

// has checks the bitset and returns true if the i'th bit is set
func (bs *bitset) has(i uint64) bool {
	return (*bs)[i>>bsL2]&(1<<(i&(bsWS-1))) != 0
}

// set takes i and sets the bit at that index to 1
func (bs *bitset) set(i uint64) {
	(*bs)[i>>bsL2] |= 1 << (i & (bsWS - 1))
}

// get takes i and returns the value of the bit at that index
func (bs *bitset) get(i uint64) uint64 {
	return (*bs)[i>>bsL2] & (1 << (i & (bsWS - 1)))
}

// unset takes i and sets the bit at that index to 0
func (bs *bitset) unset(i uint64) {
	(*bs)[i>>bsL2] &^= 1 << (i & (bsWS - 1))
}

// setall turns all the bits on
func (bs *bitset) setall() {
	for i := range bs {
		(*bs)[i] |= bsAllOn
	}
}

// clear clears all the bits
func (bs *bitset) clear() {
	for i := range bs {
		(*bs)[i] &= bsAllOff
	}
}

// getFree locates and returns the first free bit index set to 0
func (bs *bitset) getFree() int {
	for j, n := range bs {
		if n < ^uint64(0) {
			for i := 0; i < bitsetWS; i++ {
				if ((n >> i) & 1) == 0 {
					// below is shorthand for: (j * bitsetWS)+i
					return (j << bitsetL2) ^ i
				}
			}
		}
	}
	return -1
}

// String is our string method
func (bs *bitset) String() string {
	arch := 64
	resstr := strconv.Itoa(arch)
	ss := fmt.Sprintf("bitset (%d bits, %d words):\n", arch*len(*bs), len(*bs))
	for i := 0; i < len(*bs); i++ {
		ss += fmt.Sprintf("  word %d: %."+resstr+"b\n", i, (*bs)[i])
	}
	return ss
}
