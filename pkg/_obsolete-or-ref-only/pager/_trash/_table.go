package _trash

// pageTable maps pageID's to frames
type pageTable struct {
	entries map[pageID]*frame
}

func newPageTable(hint int) *pageTable {
	return &pageTable{
		entries: make(map[pageID]*frame, hint),
	}
}

func (pt *pageTable) getFrame(pid pageID) (*frame, bool) {
	f, found := pt.entries[pid]
	if !found {
		return nil, false
	}
	return f, true
}

func (pt *pageTable) getFrameByFID(fid frameID) (*frame, bool) {
	for _, f := range pt.entries {
		if f.fid == fid {
			return f, true
		}
	}
	return nil, false
}

func (pt *pageTable) addFrame(f *frame) {
	if f == nil {
		return
	}
	_, found := pt.entries[f.pid]
	if found {
		// update
		pt.entries[f.pid] = f
		return
	}
	// insert
	pt.entries[f.pid] = f
	return
}

func (pt *pageTable) delFrame(pid pageID) {
	_, found := pt.entries[pid]
	if !found {
		return
	}
	delete(pt.entries, pid)
}
