package storage

import (
	"github.com/cagnosolutions/go-data/pkg/engine/page"
)

// Storer is an interface describing the basic operations used
// to request and persist page data to and from some underlying
// storage layer.
type Storer interface {
	// AllocatePage allocates and returns the next sequential page.PageID.
	// in some cases, if there are a lot of empty fragmented pages, it may
	// return a non-sequential page.PageID.
	AllocatePage() page.PageID
	// DeallocatePage takes a page.PageID and attempts to locate and mark
	// the associated page status as free to use in the future. The data
	// may be wiped, so this is a destructive call and should be used with
	// care.
	DeallocatePage(pid page.PageID) error
	// ReadPage takes a page.PageID, as well as a (preferably empty) page.Page,
	// attempts to locate and copy the contents into p.
	ReadPage(pid page.PageID, p page.Page) error
	// WritePage takes a page.PageID, as well as a page.Page, attempts to locate
	// and copy and flush the contents of p onto the io.
	WritePage(pid page.PageID, p page.Page) error
	// Close closes the io manager.
	Close() error

	String() string
}
