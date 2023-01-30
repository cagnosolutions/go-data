package ssstore

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	// enc defines the encoding in which we persist record sizes and index entries.
	enc = binary.BigEndian
)

const (
	// lenWidth defines the number of bytes used to store the record length.
	lenWidth = 8
)

// store is a simple wrapper around a file. It supports appending and reading.
type store struct {
	*os.File               // underlying file
	mu       sync.Mutex    // writer lock
	buf      *bufio.Writer // buffered writer
	size     uint64        // file size in bytes
}

// newStore creates a store for the given file.
func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		buf:  bufio.NewWriter(f),
		size: size,
	}, nil
}

// Append persists the given bytes to the store. We will write the length of the record
// so that, when we read the record, we know how many bytes to read. We write to the buffered
// writer instead of directly to the file to reduce the number of system calls and improve
// performance.
func (s *store) Append(p []byte) (uint64, uint64, error) {
	// lock for writing (and defer unlock)
	s.mu.Lock()
	defer s.mu.Unlock()
	// get the current position which will be the beginning of this record before we write.
	pos := s.size
	// write the record length.
	err := binary.Write(s.buf, enc, uint64(len(p)))
	if err != nil {
		return 0, 0, err
	}
	// write the actual record
	n, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	// add the record length field to the bytes written
	n += lenWidth
	// updated the size to reflect the length written.
	s.size += uint64(n)
	// return bytes written, the starting position of the record and a nil error.
	return uint64(n), pos, nil
}

// Read returns the record stored at the given position. First it flushes the writer buffer, in
// case we are about to try to read a record that the buffer hasn't flushed to disk yet. We find
// out how many bytes we have to read to get the whole record, and then we fetch and return the
// record.
//
// **The compiler allocates byte slices that don't escape the functions that they are
// declared in on the stack. A value escapes when it lives beyond the lifetime of the function
// call--if you return the value for example.**
func (s *store) Read(pos uint64) ([]byte, error) {
	// lock for reading (and defer unlock)
	s.mu.Lock()
	defer s.mu.Unlock()
	// flush the write buffer
	err := s.buf.Flush()
	if err != nil {
		return nil, err
	}
	// read the record size
	size := make([]byte, lenWidth)
	_, err = s.File.ReadAt(size, int64(pos))
	if err != nil {
		return nil, err
	}
	// read the record data
	data := make([]byte, enc.Uint64(size))
	_, err = s.File.ReadAt(data, int64(pos+lenWidth))
	if err != nil {
		return nil, err
	}
	// return record data
	return data, nil
}

// ReadAt reads len(p) bytes into p beginning at the off offset in the store's file. It implements
// the io.ReaderAt interface on the store type.
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	// lock for reading (and defer unlock)
	s.mu.Lock()
	defer s.mu.Unlock()
	// flush the write buffer
	err := s.buf.Flush()
	if err != nil {
		return 0, err
	}
	// run read at
	return s.File.ReadAt(p, off)
}

// Close persists any buffered data before closing the file.
func (s *store) Close() error {
	// lock for reading (and defer unlock)
	s.mu.Lock()
	defer s.mu.Unlock()
	// flush the write buffer
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	// close the file
	return s.File.Close()
}
