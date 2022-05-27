package pager

type pageFrame struct {
	pid      pageID // id of this pageFrame
	pinCount uint32 // count of how many threads are accessing this pageFrame
	isDirty  bool   // this page pageFrame was modified but not flushed
	page            // embedded raw page
}

func newFrame(pid pageID) *pageFrame {
	return &pageFrame{
		pid:      pid,
		pinCount: 1,
		isDirty:  false,
		page:     newPage(uint32(pid)),
	}
}

func (f *pageFrame) decrPinCount() {
	if f.pinCount > 0 {
		f.pinCount--
	}
}

func (f *pageFrame) getPageID() pageID {
	return f.pid
}
