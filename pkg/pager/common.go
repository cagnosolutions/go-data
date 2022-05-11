package pager

import (
	"errors"
)

type pageID uint32
type frameID uint32

var (
	ErrPageNotFound = errors.New("page could not be found")
	ErrPageInUse    = errors.New("page is currently in use or has not been unpinned")
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
