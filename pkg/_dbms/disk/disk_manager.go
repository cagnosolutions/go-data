package disk

import (
	"errors"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

var ErrPageNotFound = errors.New("page not found")

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
