package engine

import (
	"fmt"
)

type frameID uint32

// frame is a page frame
type frame struct {
	PID      PageID  // id of this page
	FID      frameID // id or index of this frame
	PinCount uint32  // how many threads are mutating the frame
	IsDirty  bool    // page data has been modified and not flushed
	Page             // actual page data
}

func newFrame(pid PageID, fid frameID, pageSize uint16) frame {
	return frame{
		PID:      pid,
		FID:      fid,
		PinCount: 1,
		IsDirty:  false,
		Page:     make([]byte, pageSize),
	}
}

func (f *frame) decrPinCount() {
	if f.PinCount > 0 {
		f.PinCount--
	}
}

func (f *frame) reset() {
	f.PID = PageID(0)
	f.FID = frameID(0)
	f.PinCount = 0
	f.IsDirty = false
	f.Page = nil
}

func (f frame) String() string {
	return fmt.Sprintf(
		"{ PID: %d, FID: %d, PinCount: %d, dirty: %v, page: %v }",
		f.PID, f.FID, f.PinCount, f.IsDirty, f.Page.Size(),
	)
}
