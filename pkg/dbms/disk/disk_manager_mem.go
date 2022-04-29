package disk

import (
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

const DiskMaxNumPages = 5

// MemoryDiskManager is a memory mock for disk manager
type MemoryDiskManager struct {
	numPage int // tracks the number of pages. -1 indicates that there is no page, and the next to be allocates is 0
	pages   map[page.PageID]*page.Page
	writes  uint64
}

// NewMemoryDiskManagerreturns a in-memory mock of disk manager
func NewMemoryDiskManager() DiskManager {
	return &MemoryDiskManager{
		numPage: -1,
		pages:   make(map[page.PageID]*page.Page),
	}
}

// ReadPage reads a page from pages
func (d *MemoryDiskManager) ReadPage(id page.PageID, data []byte) error {
	if pg, ok := d.pages[id]; ok {
		copy(data, (*pg).Data()[:])
		return nil
	}
	return ErrPageNotFound
}

// WritePage writes a page in memory to pages
func (d *MemoryDiskManager) WritePage(id page.PageID, data []byte) error {
	p, ok := d.pages[id]
	if ok {
		p.Copy(0, data)
		d.writes++
	}
	return nil
}

// AllocatePage allocates one more page
func (d *MemoryDiskManager) AllocatePage() page.PageID {
	// if d.numPage == DiskMaxNumPages-1 {
	// 	return -1
	// }
	d.numPage = d.numPage + 1
	pageID := page.PageID(d.numPage)
	return pageID
}

// DeallocatePage removes page from disk
func (d *MemoryDiskManager) DeallocatePage(pageID page.PageID) {
	delete(d.pages, pageID)
}

func (d *MemoryDiskManager) GetNumWrites() uint64 {
	return d.writes
}

func (d *MemoryDiskManager) ShutDown() {
	return
}

func (d *MemoryDiskManager) Size() int64 {
	return int64((d.writes + 1) * page.PageSize)
	// return int64(len(d.pages) * page.PageSize)
}
