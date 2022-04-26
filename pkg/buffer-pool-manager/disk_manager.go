package buffer_pool_manager

// DiskManager is responsible for interacting with disk
type DiskManager interface {
	ReadPage(PageID) (*Page, error)
	WritePage(*Page) error
	AllocatePage() *PageID
	DeallocatePage(PageID)
}

// DiskMaxNumPages sets the disk capacity
const DiskMaxNumPages = 15
