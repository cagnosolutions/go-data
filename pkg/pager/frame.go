package pager

type frame struct {
	pid      pageID // id of this frame
	pinCount uint32 // count of how many threads are accessing this frame
	isDirty  bool   // this page frame was modified but not flushed
	page     *page  // actual raw page
}

func newFrame(pid pageID) *frame {
	return &frame{
		pid:      pid,
		pinCount: 1,
		isDirty:  false,
		page:     newPage(uint32(pid)),
	}
}

func (f *frame) decrPinCount() {
	if f.pinCount > 0 {
		f.pinCount--
	}
}

func (f *frame) getPageID() pageID {
	return f.pid
}
