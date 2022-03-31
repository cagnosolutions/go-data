package util

import (
	"bufio"
	"io"
)

const (
	defaultBufSize = 4096
	newline        = '\n'
)

// LogBuffer is a log buffer implementation
type LogBuffer struct {
	brw     *bufio.ReadWriter
	buf     []byte
	pos     int // this is the position of the cursor in the buffer
	maxSize int // this is the maximum buffered size
}

// NewLogBuffer returns a new LogBuffer
func NewLogBuffer(r io.Reader, w io.Writer, maxSize int) *LogBuffer {
	if maxSize < defaultBufSize {
		maxSize = defaultBufSize
	}
	br := bufio.NewReader(r)
	bw := bufio.NewWriter(w)
	return &LogBuffer{
		brw:     bufio.NewReadWriter(br, bw),
		buf:     make([]byte, maxSize, maxSize),
		maxSize: maxSize,
	}
}

// hasRoom returns a boolean indicating if there is
// enough room in the buffer to write size bytes
func (lb *LogBuffer) hasRoom(size int) bool {
	return (lb.maxSize - lb.pos) > size
}

// Fill reads from the buffered reader until the
// internal buffer is full
func (lb *LogBuffer) Fill() bool {
	// read until newline
	line, err := lb.brw.ReadBytes(newline)
	if err != nil {
		if err == io.EOF {
			return false
		}
	}
	// check buffer
	if !lb.hasRoom(len(line)) {
		// no more room
		return false
	}
	// otherwise, we are fine, write to internal
	// buffer and, continue
	n := copy(lb.buf[lb.pos:], line)
	// update cursor position
	lb.pos += n
	return true
}

// FlushBuffer flushes the internal buffer to the buffered
// writer, rests the cursor and buffer, and returns the
// number of bytes written along with any errors encountered
func (lb *LogBuffer) FlushBuffer() (int, error) {
	// implement
	return -1, nil
}

// Close calls FlushBuffer, and then closes and reports
// and errors encountered
func (lb *LogBuffer) Close() error {
	// implement
	return nil
}
