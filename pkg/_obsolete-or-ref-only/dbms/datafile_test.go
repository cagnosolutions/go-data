package dbms

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestDataFile_All(t *testing.T) {
	// open a data file
	df, err := OpenDataFile("testing", 0)
	if err != nil {
		t.Error(err)
	}
	defer func(df *DataFile) {
		err := df.Close()
		if err != nil {
			t.Error(err)
		}
	}(df)
}

const (
	matchGlob   = "**.db"
	filePrefix  = "dat-"
	fileSuffix  = ".db"
	metaFile    = filePrefix + "meta" + fileSuffix
	newDataFile = filePrefix + "current" + fileSuffix
)

func makeDataFileName(id int) string {
	return fmt.Sprintf("%s%.8x%s", filePrefix, id, fileSuffix)
}

func getDataFileID(name string) uint32 {
	id, err := strconv.ParseUint(name, 16, 32)
	if err != nil {
		panic(err)
	}
	return uint32(id)
}

// Namespace is a namespace struct
type Namespace struct {
	meta *os.File
	data *os.File
	base string
	name string
}

// OpenNamespace opens and returns a new *Namespace
func OpenNamespace(base, name string) (*Namespace, error) {
	// Add namespace string to base to make complete dir
	dir := filepath.Join(base, name)
	// Strip of any suffixes (if there are any) from the path
	dir = strings.Replace(dir, filepath.Ext(dir), "", -1)
	// Clean the path
	dir, err := filepath.Abs(filepath.ToSlash(dir))
	if err != nil {
		log.Panicf("cleaning path: %s\n", err)
	}
	// Create a new namespace instance we can return later
	ns := &Namespace{
		meta: nil,
		data: nil,
		base: dir,
		name: name,
	}
	// set up our filenames and file pointers for later
	metaName := filepath.Join(dir, metaFile)
	// Check to see if we need to create the files
	_, err = os.Stat(metaName)
	if os.IsNotExist(err) {
		// Touch any directories and/or file
		err = os.MkdirAll(dir, os.ModeDir|dataFilePerm)
		if err != nil {
			return nil, err
		}
		// Create our meta file
		ns.meta, err = os.OpenFile(metaName, os.O_CREATE|os.O_RDWR|os.O_SYNC, dataFilePerm)
		if err != nil {
			return nil, err
		}
		// Now, we can return without walking the directory (because this is the first time)
		return ns, nil
	}
	// Otherwise, we have to open our existing meta file
	ns.meta, err = os.OpenFile(metaName, os.O_RDWR|os.O_SYNC, dataFilePerm)
	if err != nil {
		return nil, err
	}
	// And then, we can return
	return ns, nil
}

// Walk walks a directory, running the supplied function for each glob match encountered
func (ns *Namespace) Walk(matchGlob string, fn func(de fs.DirEntry) error) error {
	// Walk our directory path
	err := filepath.WalkDir(
		filepath.Join(ns.base, ns.name), func(lpath string, info fs.DirEntry, err error) error {
			// Handle any local path errors
			if err != nil {
				log.Printf("prevent panic by handling failure accessing: %q: %v\n", lpath, err)
				return err
			}
			// Check for local file match
			matched, err := filepath.Match(matchGlob, lpath)
			if !info.IsDir() && matched {
				// We have a match, run function...
				err = fn(info)
				if err != nil {
					return err
				}
				// Otherwise, we will just keep going...
			}
			// Otherwise, return a nil error
			return nil
		},
	)
	// Check to see if there were any errors to return
	if err != nil {
		return err
	}
	return nil
}

// Close closes the namespace and files
func (ns *Namespace) Close() error {
	err := ns.meta.Close()
	if err != nil {
		return err
	}
	// err = ns.data.Close()
	// if err != nil {
	// 	return err
	// }
	return nil
}

func TestDataFile_Namespace(t *testing.T) {
	// open namespace
	ns, err := OpenNamespace("data/db", "users")
	if err != nil {
		t.Error(err)
	}
	// don't forget to close
	defer func(ns *Namespace) {
		err := ns.Close()
		if err != nil {
			t.Error(err)
		}
	}(ns)
}

// map index: (index: map[int]bitset) - meta struct in bytes: 8754 (64 indexes)
// segment index: (index: []segment) - meta struct in bytes: 8898 (64 indexes)
// bitset index: (index: []bitmap) - meta struct in bytes: 8258 (64 indexes)
func TestMeta_All(t *testing.T) {
	fmt.Println(">>>", util.Sizeof(Meta{}))
	m, err := NewMeta("data/db", "users")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(m)
}

// type segment struct {
// 	id        uint16 // segment list
// 	begOffset uint32 //
// 	endOffset uint32
// 	index     [16]uint64
// }

const (
	maxBaseLen     = 72
	maxNameLen     = 24
	metaHeaderSize = 256
)

var (
	ErrBasePathTooLong = fmt.Errorf("base is too long, max size id %d bytes", maxBaseLen)
	ErrNameTooLong     = fmt.Errorf("name is too long, max size is %d bytes", maxNameLen)
)

type Meta struct {
	base       [maxBaseLen]byte // base path
	name       [maxNameLen]byte // namespace
	fstSegment uint16           // first segment
	lstSegment uint16           // last segment
	nxtSegment uint16           // next segment
	nxtOffset  uint32           // next page offset
	pad1       uint64
	pad2       uint64
	pad3       uint32
	pad4       uint16
	index      bitset // index of the most current segment
}

func NewMeta(base, name string) (*Meta, error) {
	// Check the base and name length and return error if necessary
	if len(base) > maxBaseLen {
		return nil, ErrBasePathTooLong
	}
	if len(name) > maxNameLen {
		return nil, ErrNameTooLong
	}
	// Add name to base to make complete path
	path := filepath.Join(base, name)
	// Strip of any suffixes (if there are any) from the path
	path = strings.Replace(path, filepath.Ext(path), "", -1)
	// And, finally clean the path
	path, err := filepath.Abs(filepath.ToSlash(path))
	if err != nil {
		log.Panicf("cleaning path: %s\n", err)
	}
	// Create a new Meta instance we can return later
	m := &Meta{
		fstSegment: 0,
		lstSegment: 0,
		nxtSegment: 1,
		nxtOffset:  0,
	}
	copy(m.base[:], base)
	copy(m.name[:], name)
	// set up our filenames and file pointers for later
	fileName := filepath.Join(path, metaFile)
	var fd *os.File
	// Check to see if we need to create the files
	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		// Touch any directories and/or file
		err = os.MkdirAll(path, os.ModeDir|dataFilePerm)
		if err != nil {
			return nil, err
		}
		// Open our file
		fd, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, dataFilePerm)
		if err != nil {
			return nil, err
		}
		// Write our data
		_, err = m.WriteTo(fd)
		if err != nil {
			return nil, err
		}
		// Sync the data
		err = fd.Sync()
		if err != nil {
			return nil, err
		}
		// Close our file
		err = fd.Close()
		if err != nil {
			return nil, err
		}
		// And now, we can return our meta struct
		return m, nil
	}
	// Open our file
	fd, err = os.OpenFile(fileName, os.O_RDWR|os.O_SYNC, dataFilePerm)
	if err != nil {
		return nil, err
	}
	// Read our data
	_, err = m.ReadFrom(fd)
	if err != nil {
		return nil, err
	}
	// Close our file
	err = fd.Close()
	if err != nil {
		return nil, err
	}
	// And now, we can return our meta struct
	return m, nil
}

func (m *Meta) WriteTo(w io.Writer) (int64, error) {
	if m == nil {
		return -1, os.ErrInvalid
	}
	b := make([]byte, metaHeaderSize)
	var n int64
	copy(b[n:n+maxBaseLen], m.base[:])
	n += maxBaseLen
	copy(b[n:n+maxNameLen], m.name[:])
	n += maxNameLen
	binary.LittleEndian.PutUint16(b[n:n+2], m.fstSegment)
	n += 2
	binary.LittleEndian.PutUint16(b[n:n+2], m.lstSegment)
	n += 2
	binary.LittleEndian.PutUint16(b[n:n+2], m.nxtSegment)
	n += 2
	binary.LittleEndian.PutUint32(b[n:n+4], m.nxtOffset)
	n += 4
	binary.LittleEndian.PutUint64(b[n:n+8], m.pad1)
	n += 8
	binary.LittleEndian.PutUint64(b[n:n+8], m.pad2)
	n += 8
	binary.LittleEndian.PutUint32(b[n:n+4], m.pad3)
	n += 4
	binary.LittleEndian.PutUint16(b[n:n+2], m.pad4)
	n += 2
	for i := range m.index {
		binary.LittleEndian.PutUint64(b[n:n+8], m.index[i])
		n += 8
	}
	wrote, err := w.Write(b)
	if err != nil {
		return int64(wrote), err
	}
	if n != int64(wrote) {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func (m *Meta) ReadFrom(r io.Reader) (int64, error) {
	if m == nil {
		return -1, os.ErrInvalid
	}
	b := make([]byte, metaHeaderSize)
	read, err := r.Read(b)
	if err != nil {
		return int64(read), err
	}
	var n int64
	copy(b[n:n+maxBaseLen], m.base[:])
	n += maxBaseLen
	copy(b[n:n+maxNameLen], m.name[:])
	n += maxNameLen
	m.fstSegment = binary.LittleEndian.Uint16(b[n : n+2])
	n += 2
	m.lstSegment = binary.LittleEndian.Uint16(b[n : n+2])
	n += 2
	m.nxtSegment = binary.LittleEndian.Uint16(b[n : n+2])
	n += 2
	m.nxtOffset = binary.LittleEndian.Uint32(b[n : n+4])
	n += 4
	m.pad1 = binary.LittleEndian.Uint64(b[n : n+8])
	n += 8
	m.pad2 = binary.LittleEndian.Uint64(b[n : n+8])
	n += 8
	m.pad3 = binary.LittleEndian.Uint32(b[n : n+4])
	n += 4
	m.pad4 = binary.LittleEndian.Uint16(b[n : n+2])
	n += 2
	for i := range m.index {
		m.index[i] = binary.LittleEndian.Uint64(b[n : n+8])
		n += 8
	}
	if n != int64(read) {
		return n, io.ErrShortBuffer
	}
	return n, nil
}

func (m *Meta) String() string {
	s1 := string(m.base[:bytes.IndexByte(m.base[:], 0x00)])
	s2 := string(m.name[:bytes.IndexByte(m.name[:], 0x00)])
	ss := fmt.Sprintf("metadata:\n")
	ss += fmt.Sprintf("  base:\t\t\t\t%q\n", s1)
	ss += fmt.Sprintf("  name:\t\t\t\t%q\n", s2)
	ss += fmt.Sprintf("  first-segment:\t0x%.4x\n", m.fstSegment)
	ss += fmt.Sprintf("  last segment:\t\t0x%.4x\n", m.lstSegment)
	ss += fmt.Sprintf("  next segment:\t\t0x%.4x\n", m.nxtSegment)
	ss += fmt.Sprintf("  next offset:\t\t0x%.8x\n", m.nxtOffset)
	ss += fmt.Sprintf("  padding 1:\t\t0x%.16x\n", m.pad1)
	ss += fmt.Sprintf("  padding 2:\t\t0x%.16x\n", m.pad2)
	ss += fmt.Sprintf("  padding 3:\t\t0x%.8x\n", m.pad3)
	ss += fmt.Sprintf("  padding 4:\t\t0x%.4x\n", m.pad4)
	ss += fmt.Sprintf("  index:\n")
	for i := range m.index {
		ss += fmt.Sprintf("    extent[%.2d]=%.64b\n", i, m.index[i])
	}
	return ss
}
