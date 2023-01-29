package _pager

type pageBuffer struct {
	disk  DiskManager
	pages []page             // list of pages in memory
	free  []FrameID          // free list of page frames
	table map[PageID]FrameID // page table
}

func newPageBuffer(size int) *pageBuffer {
	pb := &pageBuffer{
		disk:  newDiskManager(),
		pages: make([]page, size),
		free:  make([]FrameID, size),
		table: make(map[PageID]FrameID),
	}
	for i := 0; i < size; i++ {
		pb.pages[i] = nil
		pb.free[i] = FrameID(i)
	}
	return pb
}

func (pb *pageBuffer) NewPage() page {
	// Get a page frame to pin to
	fid, fromFree := pb.GetFrameID()
	if &fid == nil {
		// The buffer is full, it can't find a frame
		return nil
	}
	// We should now have a frame in which we can load a page into. If we did not find
	// an open frame in the free list we will proceed by attempting to victimize the
	// current frame.
	if !fromFree {
		// Remove page from current frame.
		pg := pb.pages[fid]
		if pg != nil {
			// If the page is dirty, flush page data to disk before
			// removing the page from the page table.
			if pg.IsDirty() {
				// NOTE: consider handling error here?
				pb.disk.WritePage(PageID(pg.getPageID()), pg.Data())
			}
			// And remove page from the page table.
			delete(pb.table, PageID(pg.getPageID()))
		}
	}
	// Allocate a new page.
	pid := pb.disk.AllocatePage()
	pg := newPage(uint32(pid))

	// Update page table before returning.
	pb.table[pid] = fid
	pb.pages[fid] = pg

	return pg
}

func (pb *pageBuffer) FetchPage(pid PageID) Page {
	// TODO implement me
	panic("implement me")
}

func (pb *pageBuffer) UnpinPage(pid PageID) error {
	// TODO implement me
	panic("implement me")
}

func (pb *pageBuffer) FlushPage(pid PageID) bool {
	// TODO implement me
	panic("implement me")
}

func (pb *pageBuffer) DeletePage(pid PageID) error {
	// TODO implement me
	panic("implement me")
}

func (pb *pageBuffer) GetFrameID() (FrameID, bool) {
	// TODO implement me
	panic("implement me")
}
