package buffer

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cagnosolutions/go-data/pkg/engine/page"
)

var (
	ErrUsableFrameNotFound = errors.New("buffer: usable frame ID could not be found; this is not good")
)

// FrameID is a unsigned 32bit integer.
type FrameID = uint32

// Frame is a page frame that is used by the BufferPoolManager to
// hold and cache a page that may be written to disk, or that has
// been read from disk.
type Frame struct {
	pid       page.PageID // id of this page
	fid       FrameID     // id or index of this Frame
	pinCount  uint32      // how many threads are mutating the Frame
	isDirty   bool        // page data has been modified and not flushed
	page.Page             // actual page data
}

func (f *Frame) MarshalJSON() ([]byte, error) {
	if len(f.Page) == 0 {
		return []byte("{}"), nil
	}
	info := struct {
		FID      uint32 `json:"fid"`
		PID      uint32 `json:"pid"`
		PinCount uint32 `json:"pin_count"`
		IsDirty  bool   `json:"is_dirty"`
	}{
		FID:      uint32(f.fid),
		PID:      f.pid,
		PinCount: f.pinCount,
		IsDirty:  f.isDirty,
	}
	return json.Marshal(&info)
}

// newFrame takes a page id, a frame id along with a page size
// and allocates and returns a new Frame instance.
func newFrame(pid page.PageID, fid FrameID, pageSize uint16) Frame {
	return Frame{
		pid:      pid,
		fid:      fid,
		pinCount: 1,
		isDirty:  false,
		Page:     make([]byte, pageSize),
	}
}

// decrPinCount decrements the pin count on the frame by one. If the
// pin count is at zero, it should not decrement lower than zero. The
// pin count represents the number of processes that may currently be
// utilizing this frame.
func (f *Frame) decrPinCount() {
	if f.pinCount > 0 {
		f.pinCount--
	}
}

// resetFrame resets all the values of this frame to their zero values.
func (f *Frame) resetFrame() {
	f.pid = page.PageID(0)
	f.fid = FrameID(0)
	f.pinCount = 0
	f.isDirty = false
	f.Page = nil
}

// String implements the Stringer interface on the frame.
func (f Frame) String() string {
	return fmt.Sprintf(
		"{ pid: %d, fid: %d, pinCount: %d, dirty: %v, page: %v }",
		f.pid, f.fid, f.pinCount, f.isDirty, f.Page.Size(),
	)
}
