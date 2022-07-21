package pager

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	if err := checkSegmentSize(segmentSize); err != nil {
		panic(err)
	}
}

const (
	segmentHeaderSize        = 16 // bytes
	segmentKind       uint8  = 0xdb
	segmentVersion    uint8  = 0x01
	segmentUnlocked   uint16 = 0x0000
	segmentLocked     uint16 = 0xffff
)

func LockString(lock uint16) string {
	if lock == segmentLocked {
		return "locked"
	}
	if lock == segmentUnlocked {
		return "unlocked"
	}
	return "unknown"
}

var (
	segmentSize    uint32 = 16 << 20 // either 2, 4, 8 or 16 mb; 8 or 16 are preferable
	minSegmentSize uint32 = 2 << 20
	maxSegmentSize uint32 = 32 << 20

	ErrSegmentSizeTooSmall     = errors.New("segment size is too small (min 2 MiB)")
	ErrSegmentSizeTooLarge     = errors.New("segment size is too large (max 32 MiB)")
	ErrSegmentHeaderShortWrite = errors.New("buffer is to small to write segment header into")
	ErrSegmentHeaderShortRead  = errors.New("buffer is to small to read segment header from")
	ErrSegmentNotFound         = errors.New("segment has not been found")
)

func checkSegmentSize(size uint32) error {
	if size < minSegmentSize {
		return ErrSegmentSizeTooSmall
	}
	if maxSegmentSize < size {
		return ErrSegmentSizeTooLarge
	}
	return nil
}

// segmentHeader represents the header for a segment.
type segmentHeader struct {
	sid       uint16 // segment id
	kind      uint8  // kind
	version   uint8  // version
	firstPage uint16 // first logical page in the segment
	lastPage  uint16 // last logical page allowed in the segment (based on the page and segment size)
	maxPages  uint16 // max number of pages per segment
	pageSize  uint16 // page size
	pageCount uint16 // number of pages currently in the segment
	fileLock  uint16 // file wide lock
}

// makeSegmentHeader is the segmentHeader constructor.
func makeSegmentHeader(id, pageSize uint16) *segmentHeader {
	sh := &segmentHeader{
		sid:       id,
		kind:      segmentKind,
		version:   segmentVersion,
		maxPages:  uint16(segmentSize / uint32(pageSize)),
		pageSize:  pageSize,
		pageCount: 0,
		fileLock:  segmentUnlocked,
	}
	sh.firstPage = id * sh.maxPages
	sh.lastPage = sh.firstPage + (sh.maxPages - 1)
	return sh
}

// ReadFrom implements the ReadFrom interface on the segmentHeader.
func (sh *segmentHeader) ReadFrom(r io.Reader) (int64, error) {
	// make header sized buffer
	p := make([]byte, segmentHeaderSize)
	// read data into buffer
	n, err := r.Read(p)
	if err != nil {
		return int64(n), err
	}
	// decode data in buffer into a new segment header
	tmp := segmentHeader{
		sid:       binary.LittleEndian.Uint16(p[0:2]),
		kind:      p[2],
		version:   p[3],
		firstPage: binary.LittleEndian.Uint16(p[4:6]),
		lastPage:  binary.LittleEndian.Uint16(p[6:8]),
		maxPages:  binary.LittleEndian.Uint16(p[8:10]),
		pageSize:  binary.LittleEndian.Uint16(p[10:12]),
		pageCount: binary.LittleEndian.Uint16(p[12:14]),
		fileLock:  binary.LittleEndian.Uint16(p[14:16]),
	}
	// swap the tmp header with the current receiver, replacing it
	*sh = tmp
	_ = tmp
	return segmentHeaderSize, nil
}

// WriteTo implements the WriteTo interface on the segmentHeader.
func (sh *segmentHeader) WriteTo(w io.Writer) (int64, error) {
	// make buffer
	p := make([]byte, segmentHeaderSize)
	// encode data into buffer
	p[2] = sh.kind
	p[3] = sh.version
	binary.LittleEndian.PutUint16(p[0:2], sh.sid)
	binary.LittleEndian.PutUint16(p[4:6], sh.firstPage)
	binary.LittleEndian.PutUint16(p[6:8], sh.lastPage)
	binary.LittleEndian.PutUint16(p[8:10], sh.maxPages)
	binary.LittleEndian.PutUint16(p[10:12], sh.pageSize)
	binary.LittleEndian.PutUint16(p[12:14], sh.pageCount)
	binary.LittleEndian.PutUint16(p[14:16], sh.fileLock)
	// call write
	n, err := w.Write(p)
	return int64(n), err
}

// Read implements the Read interface on the segmentHeader.
func (sh *segmentHeader) Read(p []byte) (int, error) {
	if len(p) < segmentHeaderSize {
		return 0, ErrSegmentHeaderShortRead
	}
	nsh := &segmentHeader{
		sid:       binary.LittleEndian.Uint16(p[0:2]),
		kind:      p[2],
		version:   p[3],
		firstPage: binary.LittleEndian.Uint16(p[4:6]),
		lastPage:  binary.LittleEndian.Uint16(p[6:8]),
		maxPages:  binary.LittleEndian.Uint16(p[8:10]),
		pageSize:  binary.LittleEndian.Uint16(p[10:12]),
		pageCount: binary.LittleEndian.Uint16(p[12:14]),
		fileLock:  binary.LittleEndian.Uint16(p[14:16]),
	}
	*sh = *nsh
	return segmentHeaderSize, nil
}

// Write implements the Write interface on the segmentHeader.
func (sh *segmentHeader) Write(p []byte) (int, error) {
	if len(p) < segmentHeaderSize {
		return 0, ErrSegmentHeaderShortWrite
	}
	p[2] = sh.kind
	p[3] = sh.version
	binary.LittleEndian.PutUint16(p[0:2], sh.sid)
	binary.LittleEndian.PutUint16(p[4:6], sh.firstPage)
	binary.LittleEndian.PutUint16(p[6:8], sh.lastPage)
	binary.LittleEndian.PutUint16(p[8:10], sh.maxPages)
	binary.LittleEndian.PutUint16(p[10:12], sh.pageSize)
	binary.LittleEndian.PutUint16(p[12:14], sh.pageCount)
	binary.LittleEndian.PutUint16(p[14:16], sh.fileLock)
	return segmentHeaderSize, nil
}

// logicalPageOffset returns the logical offset of the requested page number within
// the current segment. If the page number is not present in the segment, an
// error is returned.
func (sh *segmentHeader) logicalPageOffset(pageNo uint16) (uint32, error) {
	// error check the provided page number
	if pageNo < sh.firstPage || pageNo > sh.lastPage {
		return 0, ErrPageNotFound
	}
	// transform the provided page number into a logical address
	addr := uint32(pageNo * sh.pageSize)
	// return values
	return addr, nil
}

// String is the segmentHeader stringer method.
func (sh *segmentHeader) String() string {
	ss := fmt.Sprintf("segment:\n")
	ss += fmt.Sprintf("\tsid: %d\n", sh.sid)
	ss += fmt.Sprintf("\tkind: 0x%.2x\n", sh.kind)
	ss += fmt.Sprintf("\tversion: %d\n", sh.version)
	ss += fmt.Sprintf("\tfirstPage: %d\n", sh.firstPage)
	ss += fmt.Sprintf("\tlastPage: %d\n", sh.lastPage)
	ss += fmt.Sprintf("\tmaxPages: %d\n", sh.maxPages)
	ss += fmt.Sprintf("\tpageSize: %d\n", sh.pageSize)
	ss += fmt.Sprintf("\tpageCount: %d\n", sh.pageCount)
	ss += fmt.Sprintf("\tfileLock: %q\n", LockString(sh.fileLock))
	return ss
}

const (
	segmentPrefix = "dat"
	segmentSuffix = "seg.db"
)

func fileNameHasSuffix(name string) bool {
	return strings.HasSuffix(name, segmentSuffix)
}

func getSIDFromFileName(name string) (uint16, bool) {
	re := regexp.MustCompile(`[0-9a-fA-F]{4,}`)
	sm := re.FindString(name)
	if sm == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(sm, 16, 64)
	if err != nil {
		return 0, false
	}
	return uint16(n), true
}

func makeSegmentFileName(n uint16) string {
	return fmt.Sprintf("%s-%.4x-%s", segmentPrefix, n, segmentSuffix)
}

type file struct {
	fd       *os.File
	path     string
	pageSize uint16
	segments map[uint16]string
	nextSID  uint16
	current  uint16
	fdOpen   bool
}

func makeSegmentFile(path string, pageSize uint16) *file {
	f := &file{
		path:     path,
		pageSize: pageSize,
		segments: make(map[uint16]string),
	}
	err := f.load()
	if err != nil {
		panic(err)
	}
	return f
}

func (f *file) load() error {
	files, err := os.ReadDir(f.path)
	if err != nil {
		return err
	}
	var hi uint16
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if fileNameHasSuffix(file.Name()) {
			sid, ok := getSIDFromFileName(file.Name())
			if !ok {
				continue
			}
			f.segments[sid] = file.Name()
			if sid > hi {
				hi = sid
			}
		}
	}
	f.nextSID = hi + 1
	return nil
}

func (f *file) makeSegment() error {
	sh := makeSegmentHeader(f.nextSID, f.pageSize)
	p := make([]byte, segmentHeaderSize)
	_, err := sh.Write(p)
	if err != nil {
		return err
	}
	name := makeSegmentFileName(sh.sid)
	err = os.WriteFile(filepath.Join(f.path, name), p, 0666)
	if err != nil {
		return err
	}
	f.segments[f.nextSID] = name
	f.current = f.nextSID
	f.nextSID++
	return nil
}

func (f *file) openSegment(sid uint16) error {
	name, found := f.segments[sid]
	if !found {
		return ErrSegmentNotFound
	}
	fd, err := os.OpenFile(filepath.Join(f.path, name), os.O_RDWR|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	f.fd = fd
	f.fdOpen = true
	f.current = sid
	return nil
}

func (f *file) close() error {
	if f.fdOpen {
		return f.fd.Close()
	}
	return nil
}
