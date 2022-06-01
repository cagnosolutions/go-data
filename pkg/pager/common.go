package pager

import (
	"errors"
)

type pageID = uint32
type frameID = uint32

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

	ErrUsableFrameNotFound = errors.New("usable frame ID could not be found; this is not good")
)

type frameManager interface {
	// fetchPage fetches the requested page pageFrame from the pageFrameManager
	fetchPage(pid pageID) *page
	// unpinPage unpins the target page pageFrame from the pageFrameManager
	unpinPage(pid pageID, isDirty bool) error
	// flushPage flushes the target page to the storage manager
	flushPage(pid pageID) error
	// newPage allocates a new page in the pageFrameManager requesting it from the storage manager
	newPage() *page
	// deletePage deletes a page from the pageFrameManager
	deletePage(pid pageID) error
	// getFrame fetches a free page pageFrame, otherwise it victimizes one
	getFrame() (*frame, error)
	// flushAll flushes all the pinned pages to the storage manager
	flushAll() error
}

type hashSet[T comparable] map[T]struct{}

func makeMapSet[T comparable](size int) hashSet[T] {
	return make(hashSet[T], size)
}

func (hs hashSet[T]) add(data T) {
	hs[data] = struct{}{}
}

func (hs hashSet[T]) del(data T) {
	delete(hs, data)
}

func (hs hashSet[T]) get() (T, bool) {
	for d := range hs {
		return d, true
	}
	var zero T
	return zero, false
}
