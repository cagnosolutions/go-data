package pager

import (
	"fmt"
)

// frame is a page frame
type frame struct {
	pid      pageID  // id of this page
	fid      frameID // id or index of this frame
	pinCount uint32  // count of how many threads are accessing this pageFrame
	isDirty  bool    // this page pageFrame was modified but not flushed
	page
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

func newFrame(pid pageID) *frame {
	return &frame{
		pid:      pid,
		fid:      frameID(0),
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

func (f *frame) String() string {
	ss := fmt.Sprintf("page frame:\n")
	ss += fmt.Sprintf("\tframe.pid=%d\n", f.pid)
	ss += fmt.Sprintf("\tframe.fid=%d\n", f.fid)
	ss += fmt.Sprintf("\tframe.pinCount=%d\n", f.pinCount)
	ss += fmt.Sprintf("\tframe.isDirty=%v\n", f.isDirty)
	if f.page == nil {
		ss += fmt.Sprintf("\tframe.page=%v\n", f.page)
	} else {
		ss += fmt.Sprintf("\tframe.page=%d bytes\n", len(f.page))
	}
	return ss
}
