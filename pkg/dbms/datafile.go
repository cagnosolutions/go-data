package dbms

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
)

func main() {

}

const (
	dataFileMax  = 16 << 20
	dataFilePage = 16 << 10
	dataFilePerm = 1466 // 0=none, 1=exec, 2=write, 3=exec+write, 4=read, 5=exec+read, 6=write+read, 7=write+exec+read
)

const (
	bsWS     = 64
	bsL2     = 6
	bsSize   = 16
	bsAllOn  = 0xffffffffffffffff
	bsAllOff = 0x0000000000000000
)

// bitset is a fixed sized bitset
type bitset [bsSize]uint

// has checks the bitset and returns true if the i'th bit is set
func (bs *bitset) has(i uint) bool {
	return (*bs)[i>>bsL2]&(1<<(i&(bsWS-1))) != 0
}

// set takes i and sets the bit at that index to 1
func (bs *bitset) set(i uint) {
	(*bs)[i>>bsL2] |= 1 << (i & (bsWS - 1))
}

// get takes i and returns the value of the bit at that index
func (bs *bitset) get(i uint) uint {
	return (*bs)[i>>bsL2] & (1 << (i & (bsWS - 1)))
}

// unset takes i and sets the bit at that index to 0
func (bs *bitset) unset(i uint) {
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
		if n < ^uint(0) {
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

// DataFile is an *os.File wrapper that contains some extra metadata and
// reads and writes data on page aligned blocks.
type DataFile struct {
	latch   sync.Mutex
	file    *os.File // file pointer
	index   *bitset  // bitset index of entries
	nextOff uint64   // id of next entry (used by index)
	rcount  int      // number of read calls performed
	wcount  int      // number of write calls performed
	size    int64    // current file size
}

// OpenDataFile opens and returns a new DataFile instance. If the data file
// requested does not exist, it is created and returned, along with any
// directories that may also need to be created.
func OpenDataFile(path string) (*DataFile, error) {
	// Clean path
	full, err := filepath.Abs(filepath.ToSlash(path))
	if err != nil {
		panic("pathClean: " + err.Error())
	}
	dir, _ := filepath.Split(full)
	var fp *os.File
	_, err = os.Stat(full)
	if os.IsNotExist(err) {
		// Touch any directories and/or file
		err = os.MkdirAll(dir, os.ModeDir|dataFilePerm)
		if err != nil {
			return nil, err
		}
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
	// Get the current file size info
	fi, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	// Create a new DataFile instance
	df := &DataFile{
		file:    fp,
		index:   new(bitset),
		nextOff: 1,
		rcount:  0,
		wcount:  0,
		size:    fi.Size(),
	}
	// Load our index and metadata and then return
	err = df.loadIndex()
	if err != nil {
		return nil, err
	}
	return df, nil
}

// loadIndex populates the bitset, counters and size information on open.
func (df *DataFile) loadIndex() error {
	return nil
}

// ReadAt reads a page sized slice from the DataFile starting at page offset
// off. It returns an error, if any. At end of file, that error is io.EOF.
func (df *DataFile) ReadAt(p []byte, off int64) (int, error) {
	// Perform actual read
	n, err := df.file.ReadAt(p, off)
	if err != nil {
		return -1, err
	}
	// Update read counter
	df.rcount++
	// Return
	return n, nil
}

// WriteAt writes a page to the DataFile starting at byte offset off. It
// returns an error, if any. If file was opened with the O_APPEND flag,
// WriteAt returns an error.
func (df *DataFile) WriteAt(p []byte, off int64) (int, error) {
	// Perform actual write
	n, err := df.file.WriteAt(p, off)
	if err != nil {
		return -1, err
	}
	// Sync the data
	err = df.file.Sync()
	if err != nil {
		return -1, err
	}
	// Update the bitset index
	df.index.set(uint(off))
	// Update write counter
	df.wcount++
	// Return
	return n, nil
}

// RemoveAt writes a blank page to the DataFile starting at byte offset off.
func (df *DataFile) RemoveAt(off int64) (int, error) {
	// Perform actual write
	n, err := df.file.WriteAt(make([]byte, dataFilePage), off)
	if err != nil {
		return -1, err
	}
	// Sync the data
	err = df.file.Sync()
	if err != nil {
		return -1, err
	}
	// Update the bitset index
	df.index.unset(uint(off))
	// Update write counter
	df.wcount++
	// Return
	return n, nil
}

// NextOff returns the next available offset for writing, either sequentially
// or a removed page space. If no space can be found, an error is returned.
func (df *DataFile) NextOff() (uint, error) {
	// check some stuff
	next := atomic.SwapUint64(&df.nextOff, df.nextOff+1)
	return uint(next), nil
}

// Reads returns the number read operations the DataFile has performed.
func (df *DataFile) Reads() int {
	return df.rcount
}

// Writes returns the number write operations the DataFile has performed.
func (df *DataFile) Writes() int {
	return df.wcount
}

// Size returns the current size of the DataFile
func (df *DataFile) Size() int64 {
	return df.size
}

// Close closes the current DataFile
func (df *DataFile) Close() error {
	return df.file.Close()
}
