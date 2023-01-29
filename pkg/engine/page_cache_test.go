package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestPageCache(t *testing.T) {

	pageCount := uint16(64)
	testDir := "testing"
	testFile := "page_cache_test.txt"

	pc, err := openPageCache(filepath.Join(testDir, testFile), pageCount)
	if err != nil {
		t.Errorf("opening buffer manager: %s\n", err)
	}

	page0 := pc.newPage()

	// Scenario 1: The buffer pool is empty. We should be able to create a new page.
	util.Equals(t, PageID(0), page0.getPageID())

	r1 := newRecord(rNumStr, []byte{1}, []byte("Hello, World!"))

	// Scenario 2: Once we have a page, we should be able to read and write content.
	id0, err := page0.addRecord(r1)
	if err != nil {
		t.Error(err)
	}
	rec, err := page0.getRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, newRecord(rNumStr, []byte{1}, []byte("Hello, World!")), rec)

	// Scenario 3: We should be able to create new pages until we fill up the buffer pool.
	for i := uint16(1); i < pageCount; i++ {
		p := pc.newPage()
		util.Equals(t, PageID(i), p.getPageID())
	}

	// Scenario 4: Once the buffer pool is full, we should not be able to create any new pages.
	for i := pageCount; i < pageCount*2; i++ {
		util.Equals(t, page(nil), pc.newPage())
	}

	// Scenario 5: After unpinning pages {0, 1, 2, 3, 4} and pinning another 59 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, pc.unpinPage(PageID(i), true))
		err := pc.flushPage(PageID(i))
		if err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < 4; i++ {
		pc.newPage()
	}

	// Scenario 6: We should be able to fetch the data we wrote a while ago.
	page0 = pc.fetchPage(PageID(0))
	rec2, err := page0.getRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, newRecord(rNumStr, []byte{1}, []byte("Hello, World!")), rec2)

	// Scenario 7: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, pc.unpinPage(PageID(0), true))

	pg := pc.newPage()
	util.Equals(t, PageID(pageCount+4), pg.getPageID())
	util.Equals(t, page(nil), pc.newPage())
	util.Equals(t, page(nil), pc.fetchPage(PageID(0)))

	err = pc.flushAll()
	if err != nil {
		t.Error(err)
	}

	err = pc.close()
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

func TestPageCache_HitRate(t *testing.T) {

	pageCount := uint16(16)
	testDir := "testing"
	testFile := "page_cache_test.txt"

	pc, err := openPageCache(filepath.Join(testDir, testFile), pageCount)
	if err != nil {
		t.Errorf("opening buffer manager: %s\n", err)
	}
	go pc.monitor()

	var pageIDs []uint32

	for i := 0; i < 32; i++ {
		p := pc.newPage()
		pid := p.getPageID()
		pageIDs = append(pageIDs, pid)
		log.Printf("Creating a new page (page %d)\nAdding 32 records...\n", pid)
		for j := 0; j < 32; j++ {
			var rid *RecordID
			rid, err = p.addRecord(
				newRecord(
					rNumStr, []byte{byte(j)},
					[]byte(fmt.Sprintf("this record-%d data for page-%d\n", j, i)),
				),
			)
			if err != nil {
				t.Errorf("Adding record %d to page %d failed: %s", rid, pid, err)
			}
			// log.Printf("Added record %v, to page %d\n", rid, pid)

		}
		log.Printf("Unpinning page %d\n", pid)
		err = pc.unpinPage(pid, true)
		if err != nil {
			t.Errorf("Unpinning page %d failed: %s", pid, err)
		}
		log.Printf("Flushing page %d\n", pid)
		err = pc.flushPage(pid)
		if err != nil {
			t.Errorf("Flushing page %d failed: %s", pid, err)
		}
		fmt.Println()
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println()
	for _, pid := range pageIDs[:10] {
		p := pc.fetchPage(pid)
		if p == nil {
			t.Errorf("Got nil page for pid %d\n", pid)
		}
	}

	for _, pid := range pageIDs {
		log.Printf("Fetching page %d\n", pid)
		p := pc.fetchPage(pid)
		if p == nil {
			t.Errorf("Got nil page for pid %d\n", pid)
		}
		time.Sleep(250 * time.Millisecond)
		log.Printf("Unpinning page %d\n", pid)
		err = pc.unpinPage(pid, false)
		if err != nil {
			t.Errorf("Unpinning page %d failed: %s", pid, err)
		}
		time.Sleep(250 * time.Millisecond)
	}

	err = pc.flushAll()
	if err != nil {
		t.Error(err)
	}

	err = pc.close()
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
