package dbms

import (
	"github.com/cagnosolutions/go-data/pkg/dbms/frame"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

type IOManager interface {
	AllocatePage() page.PageID
	DeallocatePage(pid page.PageID) error
	ReadPage(pid page.PageID, p page.Page) error
	WritePage(pid page.PageID, p page.Page) error
	Close() error
}

type BufferPool interface {
	// AddFreeFrame takes a frameID and adds it to the set of free frames list.
	// USES: freeList
	AddFreeFrame(fid frame.FrameID)
	// GetFrameID attempts to return a frameID. It first checks the free frame
	// list to see if there are any free frames in there. If one is found it
	// will return it along with a boolean indicating true. If none are found,
	// it will then go on to the replacer in search of one.
	// USES: freeList, Replacer
	GetFrameID() (*frame.FrameID, bool)
}

type Replacer interface {
	// Pin pins the frame matching the supplied frame ID, indicating that it should
	// not be victimized until it is unpinned.
	Pin(fid frame.FrameID)
	// Victim removes and returns the next "victim frame", as defined by the policy.
	Victim() *frame.FrameID
	// Unpin unpins the frame matching the supplied frame ID, indicating that it may
	// now be victimized.
	Unpin(fid frame.FrameID)
}
