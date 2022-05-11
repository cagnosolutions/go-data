package pager

import (
	"os"
)

type diskManager struct {
	name   string
	fp     *os.File
	nextID pageID
}

// newDiskManager instantiates and returns a new diskManager instance
func newDiskManager(filename string) *diskManager {
	// Create a new diskManager instance to return.
	dm := &diskManager{
		name:   filename,
		nextID: 0,
	}
	// open file
	// get next page offset and update nextID
	return dm
}

func (dm *diskManager) allocate() pageID {
	id := dm.nextID
	dm.nextID++
	return id
}

func (dm *diskManager) deallocate(pid pageID) {
	// TODO implement me
	panic("implement me")
}

func (dm *diskManager) read(pid pageID, p *page) error {
	// TODO implement me
	panic("implement me")
}

func (dm *diskManager) write(pid pageID, p *page) error {
	// TODO implement me
	panic("implement me")
}

func (dm *diskManager) size() int64 {
	// TODO implement me
	panic("implement me")
}

func (dm *diskManager) close() error {
	err := dm.fp.Close()
	if err != nil {
		return err
	}
	return nil
}
