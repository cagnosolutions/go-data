package ssstore

import (
	"io"
	"os"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

// index serves as a record index.
type index struct {
	file *os.File
	size uint64
}

// newIndex takes a file pointer and a config type.
func newIndex(f *os.File, c Config) (*index, error) {
	i := &index{
		file: f,
	}
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	i.size = uint64(fi.Size())
	err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes))
	if err != nil {
		return nil, err
	}
	// skipp ACTUAL mmap file call here
	return i, nil
}

// Read takes in an offset and returns the associated records position in the store.
func (i *index) Read(in int64) (uint32, uint64, error) {
	if i.size == 0 {
		debug("Read: size is == 0")
		return 0, 0, io.EOF
	}
	var out uint32
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}

	pos := uint64(out) * entWidth
	debug("read: in=%d, out=%d, pos=%d", in, out, pos)

	if i.size < pos+entWidth {
		debug("Read: size < pos+entWidth")
		return 0, 0, io.EOF
	}
	// skipping the ACTUAL mmap decode record offset and length call here
	//
	// make buffer, and read data into the buffer
	mmap := make([]byte, 12)
	_, err := i.file.ReadAt(mmap, in)
	if err != nil {
		return 0, 0, err
	}
	// decode the data that we just read into the buffer
	out = enc.Uint32(mmap[0:4])
	pos = enc.Uint64(mmap[4:12])
	return out, pos, nil
}

// Write appends the fived offset and position to the index. First, we validate that
// we have enough space to write the entry. If there is space, we then encode the
// offset and position and write them to the file. Then we increment the position
// where the next write will go.
func (i *index) Write(off uint32, pos uint64) error {
	// skipping len(mmap) call here, using the file size instead
	fi, err := i.file.Stat()
	if err != nil {
		return err
	}
	if uint64(fi.Size()) < i.size+entWidth {
		return io.EOF
	}
	// skipping mmap encode call here, writing to the file instead
	//
	// make buffer, and encode
	mmap := make([]byte, 12)
	enc.PutUint32(mmap[0:4], off)
	enc.PutUint64(mmap[4:12], pos)
	// write buffer to underlying file
	debug("writing %v at offset %d", mmap, i.size)
	_, err = i.file.WriteAt(mmap, int64(i.size))
	if err != nil {
		return err
	}
	err = i.file.Sync()
	if err != nil {
		return err
	}
	// update the index size
	i.size += uint64(entWidth)
	return nil
}

// Name simply returns the file name
func (i *index) Name() string {
	return i.file.Name()
}

// Close syncs the file contents then trims the index file to size and closes.
func (i *index) Close() error {
	// skipping ACTUAL mmap sync call here..
	err := i.file.Sync()
	if err != nil {
		return err
	}
	err = i.file.Truncate(int64(i.size))
	if err != nil {
		return err
	}
	return i.file.Close()
}
