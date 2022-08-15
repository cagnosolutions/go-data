package dbms

import (
	"errors"
	"os"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestBufferManager_All(t *testing.T) {

	pageCount := uint16(64)
	testFile := "testing/buffer_manager_tests/"

	bpm, err := OpenBufferManager(testFile, pageCount)
	if err != nil {
		t.Errorf("opening buffer manager: %s\n", err)
	}

	page0 := page.Page(bpm.NewPage())
	// fmt.Println(page0)

	// Scenario 1: The buffer pool is empty. We should be able to create a new page.
	util.Equals(t, page.PageID(0), page0.GetPageID())
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
		p := page.Page(bpm.NewPage())
		util.Equals(t, page.PageID(i), p.GetPageID())
	}
	// log.Printf("[S3] >>> DONE")

	// Scenario 4: Once the buffer pool is full, we should not be able to create any new pages.
	for i := pageCount; i < pageCount*2; i++ {
		util.Equals(t, page.Page(nil), page.Page(bpm.NewPage()))
	}
	// log.Printf("[S4] >>> DONE")

	// Scenario 5: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, bpm.UnpinPage(page.PageID(i), true))
		// log.Println("attempting to flush page", i)
		err := bpm.FlushPage(page.PageID(i))
		if err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < 4; i++ {
		bpm.NewPage()
		// p := bpm.newPage()
		// err = bpm.unpinPage(p.getPageID(), false)
		// if err != nil {
		//	t.Error(err)
		// }
	}
	// log.Printf("[S5] >>> DONE")

	// Scenario 6: We should be able to fetch the data we wrote a while ago.
	page0 = page.Page(bpm.FetchPage(page.PageID(0)))
	rec2, err := page0.GetRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec2)
	// log.Printf("[S6] >>> DONE")

	// Scenario 7: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, bpm.UnpinPage(page.PageID(0), true))

	pg := page.Page(bpm.NewPage())
	util.Equals(t, page.PageID(14), pg.GetPageID())
	util.Equals(t, page.Page(nil), bpm.NewPage())
	// fmt.Println(bpm)
	util.Equals(t, page.Page(nil), bpm.FetchPage(page.PageID(0)))
	// log.Printf("[S7] >>> DONE")

	err = bpm.flushAll()
	if err != nil {
		t.Error(err)
	}

	err = bpm.Close()
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
