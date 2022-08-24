package engine

import (
	"errors"
	"os"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestBufferManager_All(t *testing.T) {

	pageCount := uint16(64)
	testFile := "testing/buffer_manager_tests/"

	pc, err := OpenPageCache(testFile, pageCount)
	if err != nil {
		t.Errorf("opening buffer manager: %s\n", err)
	}

	page0 := Page(pc.NewPage())
	// fmt.Println(page0)

	// Scenario 1: The buffer pool is empty. We should be able to create a new page.
	util.Equals(t, PageID(0), page0.GetPageID())
	// log.Printf("[S1] >>> DONE")

	// Scenario 2: Once we have a page, we should be able to read and write content.
	id0, err := page0.AddRecord([]byte("Hello, World!"))
	if err != nil {
		t.Error(err)
	}
	rec, err := page0.GetRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec)
	// log.Printf("[S2] >>> DONE")

	// Scenario 3: We should be able to create new pages until we fill up the buffer pool.
	for i := uint16(1); i < pageCount; i++ {
		p := Page(pc.NewPage())
		util.Equals(t, PageID(i), p.GetPageID())
	}
	// log.Printf("[S3] >>> DONE")

	// Scenario 4: Once the buffer pool is full, we should not be able to create any new pages.
	for i := pageCount; i < pageCount*2; i++ {
		util.Equals(t, Page(nil), Page(pc.NewPage()))
	}
	// log.Printf("[S4] >>> DONE")

	// Scenario 5: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, pc.UnpinPage(PageID(i), true))
		// log.Println("attempting to flush page", i)
		err := pc.FlushPage(PageID(i))
		if err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < 4; i++ {
		pc.NewPage()
		// p := pc.newPage()
		// err = pc.unpinPage(p.getPageID(), false)
		// if err != nil {
		//	t.Error(err)
		// }
	}
	// log.Printf("[S5] >>> DONE")

	// Scenario 6: We should be able to fetch the data we wrote a while ago.
	page0 = Page(pc.FetchPage(PageID(0)))
	rec2, err := page0.GetRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec2)
	// log.Printf("[S6] >>> DONE")

	// Scenario 7: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, pc.UnpinPage(PageID(0), true))

	pg := Page(pc.NewPage())
	util.Equals(t, PageID(14), pg.GetPageID())
	util.Equals(t, Page(nil), pc.NewPage())
	// fmt.Println(pc)
	util.Equals(t, Page(nil), pc.FetchPage(PageID(0)))
	// log.Printf("[S7] >>> DONE")

	err = pc.flushAll()
	if err != nil {
		t.Error(err)
	}

	err = pc.Close()
	if err != nil {
		t.Error(err)
	}
}

// allocateExtent grows provided current by an extent size until it reaches
// the maximum current size, at which point an error will be returned.
func allocateExtent(fd *os.File) (int64, error) {
	// get the current size of the current
	fi, err := fd.Stat()
	if err != nil {
		return -1, err
	}
	size := fi.Size()
	// check to make sure we are not at the max current segment size
	if size == maxSegmentSize {
		return size, errors.New("current has reached the max size")
	}
	// we are below the max current size, so we should have room.
	err = fd.Truncate(size + extentSize)
	if err != nil {
		return size, err
	}
	// successfully allocated an extent, now we can return the
	// updated (current) current size, and a nil error
	fi, err = fd.Stat()
	if err != nil {
		return size, err
	}
	return fi.Size(), nil
}
