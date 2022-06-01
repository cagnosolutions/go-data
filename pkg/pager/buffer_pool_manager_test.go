package pager

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestSample(t *testing.T) {
	poolSize := 10

	testFile := "testing/diskmanager.db"

	dm := newTempDiskManager(testFile)
	defer dm.close()
	bpm := newPageManager(poolSize, dm)

	page0 := bpm.newPage()
	fmt.Println(page0)

	// Scenario: The buffer pool is empty. We should be able to create a new page.
	util.Equals(t, pageID(0), page0.getPageID())

	// Scenario: Once we have a page, we should be able to read and write content.
	id0, err := page0.addRecord([]byte("Hello, World!"))
	if err != nil {
		t.Error(err)
	}
	rec, err := page0.getRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec)

	// Scenario: We should be able to create new pages until we fill up the buffer pool.
	for i := 1; i < poolSize; i++ {
		p := bpm.newPage()
		util.Equals(t, pageID(i), p.getPageID())
	}
	// Scenario: Once the buffer pool is full, we should not be able to create any new pages.
	for i := poolSize; i < poolSize*2; i++ {
		util.Equals(t, page(nil), bpm.newPage())
	}

	// Scenario: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, bpm.unpinPage(pageID(i), true))
		log.Println("attempting to flush page", i)
		err := bpm.flushPage(pageID(i))
		if err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < 4; i++ {
		bpm.newPage()
		// p := bpm.newPage()
		// err = bpm.unpinPage(p.getPageID(), false)
		// if err != nil {
		//	t.Error(err)
		// }
	}
	// Scenario: We should be able to fetch the data we wrote a while ago.
	page0 = bpm.fetchPage(pageID(0))
	rec2, err := page0.getRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec2)

	// Scenario: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, bpm.unpinPage(pageID(0), true))

	pg := bpm.newPage()
	util.Equals(t, pageID(14), pg.getPageID())
	util.Equals(t, page(nil), bpm.newPage())
	fmt.Println(bpm)
	util.Equals(t, page(nil), bpm.fetchPage(pageID(0)))

	dm.close()
	// time.Sleep(3 * time.Second)

	// remove test files
	err = os.RemoveAll(testFile)
	if err != nil {
		t.Error(err)
	}
}
