package disk

import (
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

// DiskManager is responsible for interacting with disk
type DiskManager interface {
	ReadPage(page.PageID, []byte) error
	WritePage(page.PageID, []byte) error
	AllocatePage() page.PageID
	DeallocatePage(page.PageID)
	GetNumWrites() uint64
	ShutDown()
	Size() int64
}
