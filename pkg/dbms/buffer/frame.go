package buffer

import (
	"fmt"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

type frameID uint32
type FrameID = frameID

// frame is a page frame
type frame struct {
	pid       page.PageID // id of this page
	fid       frameID     // id or index of this frame
	pinCount  uint32      // how many threads are mutating the frame
	isDirty   bool        // page data has been modified and not flushed
	page.Page             // actual page data
}

func (f *frame) decrPinCount() {
	if f.pinCount > 0 {
		f.pinCount--
	}
}

func newFrame(pid page.PageID, fid frameID, pageSize uint16) frame {
	return frame{
		pid:      pid,
		fid:      fid,
		pinCount: 1,
		isDirty:  false,
		Page:     make([]byte, pageSize),
	}
}

func (f *frame) reset() {
	f.pid = page.PageID(0)
	f.fid = frameID(0)
	f.pinCount = 0
	f.isDirty = false
	f.Page = nil
}

func (f frame) String() string {
	return fmt.Sprintf(
		"{ pid: %d, fid: %d, pinCount: %d, dirty: %v, page: %v }",
		f.pid, f.fid, f.pinCount, f.isDirty, f.Page.Size(),
	)
}
