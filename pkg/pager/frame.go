package pager

import (
	"fmt"
)

// frame is a page frame
type frame struct {
	pid      pageID  // id of this page
	fid      frameID // id or index of this frame
	pinCount uint32  // how many threads are mutating the frame
	isDirty  bool    // page data has been modified and not flushed
	page             // actual page data
}

func (f *frame) decrPinCount() {
	if f.pinCount > 0 {
		f.pinCount--
	}
}

func initFrame(fid frameID) *frame {
	return &frame{
		pid:      pageID(0),
		fid:      fid,
		pinCount: 0,
		isDirty:  false,
		page:     make([]byte, szPg),
	}
}

func newFrame(pid pageID, fid frameID) frame {
	return frame{
		pid:      pid,
		fid:      fid,
		pinCount: 1,
		isDirty:  false,
		page:     make([]byte, szPg),
	}
}

func (f *frame) reset() {
	f.pid = pageID(0)
	f.fid = frameID(0)
	f.pinCount = 0
	f.isDirty = false
	f.page = nil
}

func (f frame) String() string {
	return fmt.Sprintf(
		"{ pid: %d, fid: %d, pinCount: %d, dirty: %v, page: %v }",
		f.pid, f.fid, f.pinCount, f.isDirty, f.page.size(),
	)
}
