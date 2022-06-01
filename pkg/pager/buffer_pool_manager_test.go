package pager

import (
	"fmt"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestSample(t *testing.T) {
	poolSize := 10

	dm := newTempDiskManager("testing/diskmanager.db")
	defer dm.close()
	bpm := newPageManager(poolSize, dm)

	page0 := bpm.newPage()
	fmt.Printf(">>> page header <<<\n%+v\n", page0.getHeader())

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
		util.Equals(t, nil, bpm.newPage())
	}

	// Scenario: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, bpm.unpinPage(pageID(i), true))
		bpm.flushPage(pageID(i))
	}
	for i := 0; i < 4; i++ {
		bpm.newPage()
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

	util.Equals(t, pageID(18), bpm.newPage().getPageID())
	util.Equals(t, (*page)(nil), bpm.newPage())
	util.Equals(t, (*page)(nil), bpm.fetchPage(pageID(0)))
}
