package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestPageCache(t *testing.T) {

	pageCount := uint16(64)
	testDir := "testing"
	testFile := "page_cache_test.txt"

	pc, err := OpenPageCache(filepath.Join(testDir, testFile), pageCount)
	if err != nil {
		t.Errorf("opening buffer manager: %s\n", err)
	}

	page0 := pc.NewPage()

	// Scenario 1: The buffer pool is empty. We should be able to create a new page.
	util.Equals(t, PageID(0), page0.getPageID())

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

	// Scenario 3: We should be able to create new pages until we fill up the buffer pool.
	for i := uint16(1); i < pageCount; i++ {
		p := pc.NewPage()
		util.Equals(t, PageID(i), p.getPageID())
	}

	// Scenario 4: Once the buffer pool is full, we should not be able to create any new pages.
	for i := pageCount; i < pageCount*2; i++ {
		util.Equals(t, Page(nil), pc.NewPage())
	}

	// Scenario 5: After unpinning pages {0, 1, 2, 3, 4} and pinning another 59 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, pc.UnpinPage(PageID(i), true))
		err := pc.FlushPage(PageID(i))
		if err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < 4; i++ {
		pc.NewPage()
	}

	// Scenario 6: We should be able to fetch the data we wrote a while ago.
	page0 = pc.FetchPage(PageID(0))
	rec2, err := page0.GetRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec2)

	// Scenario 7: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, pc.UnpinPage(PageID(0), true))

	pg := pc.NewPage()
	util.Equals(t, PageID(pageCount+4), pg.getPageID())
	util.Equals(t, Page(nil), pc.NewPage())
	util.Equals(t, Page(nil), pc.FetchPage(PageID(0)))

	err = pc.flushAll()
	if err != nil {
		t.Error(err)
	}

	err = pc.Close()
	if err != nil {
		t.Error(err)
	}

	err = os.Remove(filepath.Join(testDir, testFile))
	if err != nil {
		t.Error(err)
	}

	err = os.Remove(testDir)
	if err != nil {
		t.Error(err)
	}

}
