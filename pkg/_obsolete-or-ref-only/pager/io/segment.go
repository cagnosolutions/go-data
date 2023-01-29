package io

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	segmentSize = 1024 // 16 << 20 // 16 MB

	segmentPerm = 0644
	segmentFlag = os.O_RDWR | os.O_SYNC

	defaultSegmentName = "main"
	filePrefix         = "seg"
	fileSuffix         = ".db"

	minSegmentSize = 1024 // 64 << 10
	maxSegmentSize = 32 << 20
)

func getSegName(seqID uint32) string {
	return fmt.Sprintf("%s-%.5d-%.4s%.3s", filePrefix, seqID, defaultSegmentName, fileSuffix)
}

func getSegByName(name string, seqID uint) string {
	return fmt.Sprintf("%s-%.5d-%.4s%.3s", filePrefix, seqID, name, fileSuffix)
}

func getSegID(name string) uint32 {
	i := strings.Index(name, "-")
	id, err := strconv.ParseUint(name[i+1:i+6], 10, 16)
	if err != nil {
		panic("error: could not get sequence id from name; " + err.Error())
	}
	return uint32(id)
}

var (
	ErrSegmentSizeTooSmall = errors.New("segment size is smaller than the min (64 KiB)")
	ErrSegmentSizeTooLarge = errors.New("segment size is larger than the max (32 MiB)")
)

// segment is a simple wrapper around a file. It supports reading and reading writing.
type segment struct {
	fp    *os.File
	id    uint32
	psize uint32
	size  uint32
}

func (s *segment) String() string {
	return fmt.Sprintf("segment{ id=%d, file=%q, size=%d }\n", s.id, s.fp.Name(), s.size)
}

// newSegment initializes and returns a new segment.
func newSegment(dir string, id uint32) (*segment, error) {
	s := &segment{
		fp:   nil,
		id:   id,
		size: 0,
	}
	var err error
	file := filepath.ToSlash(filepath.Join(dir, getSegName(id)))
	s.fp, err = os.OpenFile(file, segmentFlag, segmentPerm)
	if err != nil {
		return nil, err
	}
	fi, err := s.fp.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Size() > 0 {
		s.size = uint32(fi.Size())
		s.getPageSize()
	}
	return s, nil
}

func (s *segment) getPageSize() uint32 {
	if s.size > 0 {
		buf := make([]byte, 4)
		_, err := s.fp.ReadAt(buf, 4)
		if err != nil {
			panic("error reading page size: " + err.Error())
		}
		s.psize = binary.LittleEndian.Uint32(buf)
	}
	return s.psize
}

func openSegment(dir string, id uint32) (*segment, error) {
	return newSegment(dir, id)
}

// ReadAt reads len(p) bytes into p beginning at the off offset in the segment's file.
// It implements the io.ReaderAt interface on the segment type.
func (s *segment) ReadAt(p []byte, off int64) (int, error) {
	// run read at
	return s.fp.ReadAt(p, off)
}

// WriteAt writes len(p) bytes from p beginning at the off offset in the segment's file.
// It implements the io.WriterAt interface on the segment type.
func (s *segment) WriteAt(p []byte, off int64) (int, error) {
	// check to ensure the segment is not full
	if !s.hasRoom(len(p)) {
		return 0, io.EOF
	}
	// run write at
	n, err := s.fp.WriteAt(p, off)
	if err != nil {
		return n, err
	}
	// sync data
	err = s.fp.Sync()
	if err != nil {
		return n, err
	}
	// update size
	s.size += uint32(n)
	// return
	return n, nil
}

// hasRoom returns whether the segment will have room to fit n more bytes.
func (s *segment) hasRoom(n int) bool {
	return s.available() < uint32(n)
}

// available returns the number of bytes that are still unused.
func (s *segment) available() uint32 {
	if s.size == 0 {
		return segmentSize
	}
	return segmentSize - s.size
}

// IsFull returns whether the segment has reached its max size.
func (s *segment) IsFull() bool {
	return s.size >= segmentSize
}

// Remove closes the segment and removes the underlying file.
func (s *segment) Remove() error {
	file := s.fp.Name()
	err := s.fp.Close()
	if err != nil {
		return err
	}
	n := strings.LastIndex(file, "/")
	err = os.Remove(file[n+1:])
	if err != nil {
		return err
	}
	return nil
}

// Close closes the segment.
func (s *segment) Close() error {
	err := s.fp.Close()
	if err != nil {
		return err
	}
	return nil
}

// nearestMultiple returns the nearest and lesser multiple of k in j. We take the
// lesser multiple to make sure we stay under the user's disk capacity.
func nearestMultiple(j, k uint32) uint32 {
	if j > 0 {
		return (j / k) * k
	}
	return ((j - k + 1) / k) * k
}

func (s *segment) addrInSegment(addr uint32) bool {
	return s.startOffset() <= addr && addr <= s.endingOffset()
}

func (s *segment) startOffset() uint32 {
	if s.psize == 0 {
		return 0
	}
	return (s.id * s.maxRecords()) * s.psize
}

func (s *segment) endingOffset() uint32 {
	if s.psize == 0 {
		return 0
	}
	return s.startOffset() + ((s.maxRecords() - 1) * s.psize)
}

func (s *segment) maxRecords() uint32 {
	if s.size > 0 && s.psize == 0 {
		s.getPageSize()
	}
	return segmentSize / s.psize
}
