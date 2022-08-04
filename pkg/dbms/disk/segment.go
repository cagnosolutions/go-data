package disk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	segmentSize       uint32 = 16 << 20 // either 2, 4, 8 or 16mb;
	segmentHeaderSize uint8  = 16       // bytes
	segmentKind       uint8  = 0xdb
	segmentVersion    uint8  = 0x01
	segmentUnlocked   uint16 = 0x0000
	segmentLocked     uint16 = 0xffff
)

var (
	// segmentSize    uint32 = 16 << 20 // either 2, 4, 8 or 16 mb; 8 or 16 are preferable
	minSegmentSize uint32 = 2 << 20
	maxSegmentSize uint32 = 32 << 20

	ErrSegmentSizeTooSmall     = errors.New("segment size is too small (min 2 MiB)")
	ErrSegmentSizeTooLarge     = errors.New("segment size is too large (max 32 MiB)")
	ErrSegmentHeaderShortWrite = errors.New("buffer is to small to write segment header into")
	ErrSegmentHeaderShortRead  = errors.New("buffer is to small to read segment header from")
	ErrSegmentNotFound         = errors.New("segment has not been found")
)

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
	return int64(segmentHeaderSize), nil
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
	if len(p) < int(segmentHeaderSize) {
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
	return int(segmentHeaderSize), nil
}

// Write implements the Write interface on the segmentHeader.
func (sh *segmentHeader) Write(p []byte) (int, error) {
	if len(p) < int(segmentHeaderSize) {
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
	return int(segmentHeaderSize), nil
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
	if sh.fileLock == segmentLocked {
		ss += fmt.Sprintf("\tfileLock: %q\n", "locked")
	}
	if sh.fileLock == segmentUnlocked {
		ss += fmt.Sprintf("\tfileLock: %q\n", "unlocked")
	}
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

const ()

type SegmentFile struct {
	lock        sync.Mutex
	path        string
	prefix      string
	segmentSize uint32
	pageSize    uint16
	fd          *os.File
	firstIndex  uint16
	lastIndex   uint16
	segments    []segmentHeader
	active      *segmentHeader
}

func OpenSegmentedFile(path, prefix string, pageSize uint16) (*SegmentFile, error) {
	// make sure we are working with absolute paths
	base, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// sanitize any path separators
	base = filepath.ToSlash(base)
	// create any directories if they are not there
	err = os.MkdirAll(base, os.ModeDir)
	if err != nil {
		return nil, err
	}
	// create a new segment file instance
	s := &SegmentFile{
		path:        base,
		prefix:      prefix,
		segmentSize: segmentSize,
		pageSize:    pageSize,
		firstIndex:  0,
		lastIndex:   1,
		segments:    make([]segmentHeader, 0),
	}
	// attempt to index any existing segments
	err = s.indexSegments()
	if err != nil {
		return nil, err
	}
	// return segment file
	return s, nil
}

func (s *SegmentFile) indexSegments() error {
	// lock
	s.lock.Lock()
	defer s.lock.Unlock()
	// get the files in the base directory path
	files, err := os.ReadDir(s.path)
	if err != nil {
		return err
	}
	// list the files in the base directory path and attempt to index the entries
	for _, f := range files {
		// skip non data files
		if f.IsDir() || !strings.HasPrefix(f.Name(), s.prefix) {
			continue // skip this, continue on to the next file
		}
		// check the size of segment file
		fi, err := f.Info()
		if err != nil {
			return err
		}
		// if the file is empty, remove it and skip to next file
		if fi.Size() == 0 {
			err = os.Remove(filepath.Join(s.path, f.Name()))
			if err != nil {
				return err
			}
			continue // make sure we skip to next segment
		}
		// attempt to read segment file header
		sh, err := s.readSegmentFileHeader(filepath.Join(s.path, f.Name()))
		if err != nil {
			return err
		}
		// segment has been loaded successfully, append to the segments list
		s.segments = append(s.segments, *sh)
	}
	// check to see if any segments were found. If not, initialize a new one
	if len(s.segments) == 0 {
		// create a new segment file
		sh, err := s.writeSegmentFileHeader(s.lastIndex)
		if err != nil {
			return err
		}
		// segment has been created successfully, append to the segments list
		s.segments = append(s.segments, *sh)
	}
	// segments have either been loaded or created, so now we should go about
	// updating the active segment pointer to point to the "tail" (the last
	// segment in the segment list)
	s.active = s.getLastSegmentFileHeader()
	// we should be good to go, lets attempt to open a file descriptor to work
	// with the active segment
	s.fd, err = os.OpenFile(s.path, os.O_SYNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	// finally, update the firstIndex and lastIndex
	s.firstIndex = s.segments[0].firstPage
	// and update last index
	s.lastIndex = s.getLastSegmentFileHeader().lastPage
	return nil
}

func (s *SegmentFile) readSegmentFileHeader(path string) (*segmentHeader, error) {
	// check to make sure path exists before continuing
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	// attempt to open existing segment file for reading
	fd, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	// defer file close
	defer func(fd *os.File) {
		_ = fd.Close()
	}(fd)
	// create new segment header to fill out
	var sh *segmentHeader
	// read segment file header
	_, err = sh.ReadFrom(fd)
	if err != nil {
		return nil, err
	}
	// return segment file header
	return sh, nil
}

func (s *SegmentFile) writeSegmentFileHeader(id uint16) (*segmentHeader, error) {
	// create a new file
	path := filepath.Join(s.path, makeSegmentFileName(id))
	fd, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	// create new segment file header
	sh := makeSegmentHeader(id, s.pageSize)
	// write segment file header to disk
	_, err = sh.WriteTo(fd)
	if err != nil {
		return nil, err
	}
	// don't forget to close it
	err = fd.Close()
	if err != nil {
		return nil, err
	}
	// return segment file header
	return sh, nil
}

func (s *SegmentFile) getLastSegmentFileHeader() *segmentHeader {
	return &s.segments[len(s.segments)-1]
}

func (s *SegmentFile) cycleSegmentFile() error {
	// sync and close current file segment
	err := s.fd.Close()
	if err != nil {
		return err
	}
	// make new segment file
	sh, err := s.writeSegmentFileHeader(s.lastIndex)
	if err != nil {
		return err
	}
	// add segment to segment index list
	s.segments = append(s.segments, *sh)
	// update the active segment pointer
	s.active = s.getLastSegmentFileHeader()
	// close current segment file, file descriptor
	err = s.fd.Close()
	if err != nil {
		return err
	}
	// open file descriptor
	s.fd, err = os.OpenFile(s.path, os.O_SYNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (s *SegmentFile) CloseAndRemove() error {
	// lock
	s.lock.Lock()
	defer s.lock.Unlock()
	// sync and close writer
	err := s.fd.Close()
	if err != nil {
		return err
	}
	// reset the segments
	s.segments = make([]segmentHeader, 0)
	// reset first and last index
	s.firstIndex = 0
	s.lastIndex = 1
	// erase all files
	err = os.RemoveAll(s.path)
	if err != nil {
		return err
	}
	return nil
}

func (s *SegmentFile) findSegmentFileHeader(id uint32) int {
	// declare for later
	i, j := 0, len(s.segments)
	// otherwise, perform binary search
	for i < j {
		h := i + (j-i)/2
		if uint16(id) >= s.segments[h].sid {
			i = h + 1
		} else {
			j = h
		}
	}
	return i - 1
}

func (s *SegmentFile) Read(id uint32, p []byte) error {
	// read lock
	s.lock.Lock()
	defer s.lock.Unlock()
	// error checking
	if uint16(id) < s.firstIndex || uint16(id) > s.lastIndex {
		return errors.New("out of bounds")
	}
	var err error
	// get the segment where the id is located
	sid := uint32(math.Floor(float64(((id) * uint32(s.pageSize)) / s.segmentSize)))
	// find the segment containing the provided index
	sh := s.segments[s.findSegmentFileHeader(sid)]
	// get the logical offset address of the id provided
	addr, err := sh.logicalPageOffset(uint16(id))
	if err != nil {
		return err
	}
	// get the full path from the segment file header
	path := filepath.Join(s.path, makeSegmentFileName(sh.sid))
	// open a new file descriptor at the selected path
	fd, err := os.OpenFile(path, os.O_SYNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	// read entry
	_, err = fd.ReadAt(p, int64(addr))
	if err != nil {
		return err
	}
	// don't forget to close
	err = fd.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *SegmentFile) Write(id uint32, p []byte) error {
	// lock
	s.lock.Lock()
	defer s.lock.Unlock()
	var err error
	// get the segment where the id is located
	sid := uint32(math.Floor(float64(((id) * uint32(s.pageSize)) / s.segmentSize)))
	// find the segment containing the provided index
	sh := s.segments[s.findSegmentFileHeader(sid)]
	// get the logical offset address of the id provided
	addr, err := sh.logicalPageOffset(uint16(id))
	if err != nil {
		return err
	}
	// get the full path from the segment file header
	path := filepath.Join(s.path, makeSegmentFileName(sh.sid))
	// open a new file descriptor at the selected path
	fd, err := os.OpenFile(path, os.O_SYNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	// read entry
	_, err = fd.WriteAt(p, int64(addr))
	if err != nil {
		return err
	}
	// don't forget to close
	err = fd.Close()
	if err != nil {
		return err
	}
	// update lastIndex
	s.lastIndex++
	// check to see if the active segment needs to be cycled
	if s.active.pageCount >= s.active.maxPages {
		err = s.cycleSegmentFile()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SegmentFile) Close() error {
	// lock
	s.lock.Lock()
	defer s.lock.Unlock()
	// sync and close writer
	err := s.fd.Close()
	if err != nil {
		return err
	}
	// clean everything else up
	s.fd = nil
	s.firstIndex = 0
	s.lastIndex = 0
	s.segments = nil
	s.active = nil
	// force gc for good measure
	runtime.GC()
	return nil
}

func (s *SegmentFile) String() string {
	var ss string
	ss += fmt.Sprintf("\n\n[segment file]\n")
	ss += fmt.Sprintf("path: %q\n", s.path)
	ss += fmt.Sprintf("firstIndex: %d\n", s.firstIndex)
	ss += fmt.Sprintf("lastIndex: %d\n", s.lastIndex)
	ss += fmt.Sprintf("segments: %d\n", len(s.segments))
	if s.active != nil {
		path := filepath.Join(s.path, makeSegmentFileName(s.active.sid))
		ss += fmt.Sprintf("active: %q\n", path)
	}
	if len(s.segments) > 0 {
		for i, sh := range s.segments {
			path := filepath.Join(s.path, makeSegmentFileName(sh.sid))
			ss += fmt.Sprintf("segment[%d]:\n", i)
			ss += fmt.Sprintf("\tpath: %q\n", path)
			ss += sh.String()
		}
	}
	ss += "\n"
	return ss
}
