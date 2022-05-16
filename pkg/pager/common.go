package pager

import (
	"errors"
	"unsafe"
)

const (
	kb           = 1 << 10
	mb           = 1 << 20
	pageSize     = 4 * kb
	hdrSize      = uint16(unsafe.Sizeof(_header{}))
	slotSize     = 6
	segmSize     = 2 * mb
	pagesPerSegm = segmSize / pageSize
	minRecSize   = 8
	maxRecSize   = pageSize - hdrSize - slotSize
)

// header is the page's header
type _header struct {
	pid   uint32 // page id
	magic uint16 // status and type (for now, but can include others)
	slots uint16 // number of slots in page
	lower uint16 // free space lower offset
	upper uint16 // free space upper offset
}

type pageID uint32
type frameID uint32

var (
	ErrPageNotFound      = errors.New("page could not be found")
	ErrPageInUse         = errors.New("page is currently in use or has not been unpinned")
	ErrNilPage           = errors.New("page is nil")
	ErrOffsetOutOfBounds = errors.New("calculated offset is outside file bounds")
	ErrPartialPageWrite  = errors.New("page write was not a full page")
	ErrPartialPageRead   = errors.New("page read was not a full page")
	ErrSlotIDOutOfBounds = errors.New("slot id is outside of the lower bounds")
	ErrMinRecSize        = errors.New("record is smaller than the minimum allowed record size")
	ErrMaxRecSize        = errors.New("record is larger than the maximum allowed record size")
	ErrPossiblePageFull  = errors.New("page might be full (but may have fragmented space available)")
	ErrPageFull          = errors.New("page is full and out of available space")
	ErrBadRID            = errors.New("bad record id; either the page id or the slot id did not match")
	ErrRecNotFound       = errors.New("record has not been found")
)

type pageFrameManager interface {
	// fetchPage fetches the requested page frame from the pageFrameManager
	fetchPage(pid pageID) *frame
	// unpinPage unpins the target page frame from the pageFrameManager
	unpinPage(pid pageID, isDirty bool) error
	// flushPage flushes the target page to the storage manager
	flushPage(pid pageID) error
	// newPage allocates a new page in the pageFrameManager requesting it from the storage manager
	newPage() *frame
	// deletePage deletes a page from the pageFrameManager
	deletePage(pid pageID) error
	// getFrame fetches a free page frame, otherwise it victimizes one
	getFrame() (*frameID, error)
	// flushAll flushes all the pinned pages to the storage manager
	flushAll() error
}

type storageManager interface {
	allocate() pageID
	deallocate(pid pageID)
	read(pid pageID, p *page) error
	write(pid pageID, p *page) error
	size() int64
	close() error
}

type replacer interface {
	pin(fid frameID)
	unpin(fid frameID)
	victim() frameID
	size() int64
}

type mapSet map[frameID]struct{}

func makeMapSet(size int) mapSet {
	return make(mapSet, size)
}

func (s mapSet) add(k frameID) {
	s[k] = struct{}{}
}

func (s mapSet) del(k frameID) {
	delete(s, k)
}

func (s mapSet) get() (frameID, bool) {
	for k := range s {
		return k, true
	}
	return 0, false
}
