package dbms

import (
	"errors"
	"os"
)

// allocateExtent grows provided file by an extent size until it reaches
// the maximum file size, at which point an error will be returned.
func allocateExtent(fd *os.File) (int64, error) {
	// get the current size of the file
	fi, err := fd.Stat()
	if err != nil {
		return -1, err
	}
	size := fi.Size()
	// check to make sure we are not at the max file segment size
	if size == maxSegmentSize {
		return size, errors.New("file has reached the max size")
	}
	// we are below the max file size, so we should have room.
	err = fd.Truncate(size + extentSize)
	if err != nil {
		return size, err
	}
	// successfully allocated an extent, now we can return the
	// updated (current) file size, and a nil error
	fi, err = fd.Stat()
	if err != nil {
		return size, err
	}
	return fi.Size(), nil
}
