package pager

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

func newFrame(pid pageID) *frame {
	return &frame{
		pid:      pid,
		fid:      0,
		pinCount: 1,
		isDirty:  false,
	}
}
