package _pager

type diskManager struct {
}

func newDiskManager() *diskManager {
	return new(diskManager)
}

func (d *diskManager) AllocatePage() PageID {
	// TODO implement me
	panic("implement me")
}

func (d *diskManager) ReadPage(pid PageID, dst []byte) error {
	// TODO implement me
	panic("implement me")
}

func (d *diskManager) WritePage(pid PageID, src []byte) error {
	// TODO implement me
	panic("implement me")
}

func (d *diskManager) DeallocatePage(pid PageID) {
	// TODO implement me
	panic("implement me")
}

func (d *diskManager) GetNumReads() uint {
	// TODO implement me
	panic("implement me")
}

func (d *diskManager) GetNumWrites() uint {
	// TODO implement me
	panic("implement me")
}

func (d *diskManager) ShutDown() error {
	// TODO implement me
	panic("implement me")
}

func (d *diskManager) Size() int64 {
	// TODO implement me
	panic("implement me")
}
