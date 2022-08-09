package frame

import (
	"fmt"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

type frameID uint32
type FrameID = frameID
type Frame = frame

// frame is a page frame
type frame struct {
	PID       page.PageID // id of this page
	FID       frameID     // id or index of this frame
	PinCount  uint32      // how many threads are mutating the frame
	IsDirty   bool        // page data has been modified and not flushed
	page.Page             // actual page data
}

func (f *Frame) DecrPinCount() {
	f.decrPinCount()
}

func (f *frame) decrPinCount() {
	if f.PinCount > 0 {
		f.PinCount--
	}
}

func NewFrame(pid page.PageID, fid FrameID, pageSize uint16) Frame {
	return newFrame(pid, fid, pageSize)
}

func newFrame(pid page.PageID, fid frameID, pageSize uint16) frame {
	return frame{
		PID:      pid,
		FID:      fid,
		PinCount: 1,
		IsDirty:  false,
		Page:     make([]byte, pageSize),
	}
}

func (f *Frame) Reset() {
	f.reset()
}

func (f *frame) reset() {
	f.PID = page.PageID(0)
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
