package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	filePerm = 0644
	fileFlag = os.O_RDWR | os.O_CREATE | os.O_SYNC
)

type diskStore struct {
	// fp      *os.File
	dir      string   // base directory
	segments []uint32 // list of segment sequence id's
	active   uint32   // active segment in pointer
	ptr      *segment // the active segment
}

// openDiskStore creates of
func openDiskStore(base string) (*diskStore, error) {
	// Create directory or path if it doesn't exist.
	err := os.MkdirAll(base, filePerm)
	if err != nil {
		return nil, err
	}
	// Create a new diskStore instance, and run setup
	d := &diskStore{
		dir:      base,
		segments: make([]uint32, 0),
	}
	// Run setup
	err = d.setup()
	if err != nil {
		return nil, err
	}
	// return diskStore
	return d, nil
}

func openFile(dir, name string) (*os.File, error) {
	fp, err := os.OpenFile(filepath.Join(dir, name), fileFlag, filePerm)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

func (d *diskStore) setup() error {
	files, err := os.ReadDir(d.dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), filePrefix) &&
			strings.HasSuffix(file.Name(), fileSuffix) {
			sid := getSegID(file.Name())
			d.segments = append(d.segments, sid)
			if sid > d.active {
				d.active = sid
			}
		}
	}
	d.ptr, err = openSegment(d.dir, d.active)
	if err != nil {
		return err
	}
	return nil
}

func (d *diskStore) allocate() uint32 {
	return 0
}

func (d *diskStore) deallocate(pid uint32) error {
	return nil
}

func (d *diskStore) read(pid uint32, p []byte) error {
	return nil
}

func (d *diskStore) write(pid uint32, p []byte) error {
	// get logical offset address
	addr := pid * uint32(len(p))
	// locate proper segment
	if !d.ptr.addrInSegment(addr) {
		sid := d.findSegment(addr)
		panic(fmt.Sprintf("addr=%d, not in current segment (%d), but should be in segment %d\n",
			addr, d.active, sid))
	}
	// write data
	_, err := d.ptr.WriteAt(p, int64(addr))
	if err != nil {
		return err
	}
	return nil
}

func (d *diskStore) findSegment(addr uint32) uint32 {
	max := d.ptr.maxRecords()
	return nearestMultiple(addr, max) / max
}

func (d *diskStore) close() error {
	if d.ptr == nil {
		return nil
	}
	return d.ptr.Remove()
}

func (d *diskStore) String() string {
	ss := fmt.Sprintf("disk store\n")
	ss += fmt.Sprintf("\tdir=%q\n", d.dir)
	ss += fmt.Sprintf("\tsegments=%v\n", d.segments)
	ss += fmt.Sprintf("\tactive=%d\n", d.active)
	ss += fmt.Sprintf("\tptr=%v\n", d.ptr)
	ss += "\n"
	return ss
}
