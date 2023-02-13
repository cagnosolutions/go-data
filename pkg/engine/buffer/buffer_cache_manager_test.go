package buffer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cagnosolutions/go-data/pkg/engine/page"
	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestPageCache(t *testing.T) {

	cleanUpWhenDone := true

	pageCount := uint16(64)
	testDir := "testing"
	testFile := "page_cache_test.txt"

	pc, err := OpenBufferCacheManager(filepath.Join(testDir, testFile), pageCount)
	if err != nil {
		t.Errorf("opening buffer manager: %s\n", err)
	}

	fmt.Println(">>>\n", pc)

	page0 := pc.NewPage()

	// Scenario 1: The buffer pool is empty. We should be able to create a new page.
	util.Equals(t, page.PageID(0), page0.GetPageID())

	r1 := page.NewRecord(page.R_NUM, page.R_STR, []byte{1}, []byte("Hello, World!"))

	// Scenario 2: Once we have a page, we should be able to read and write content.
	id0, err := page0.AddRecord(r1)
	if err != nil {
		t.Error(err)
	}
	rec, err := page0.GetRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, page.NewRecord(page.R_NUM, page.R_STR, []byte{1}, []byte("Hello, World!")), rec)

	fmt.Println(">>>\n", pc)

	// Scenario 3: We should be able to create new pages until we fill up the buffer pool.
	for i := uint16(1); i < pageCount; i++ {
		p := pc.NewPage()
		util.Equals(t, page.PageID(i), p.GetPageID())
	}

	fmt.Println(">>>\n", pc)

	// Scenario 4: Once the buffer pool is full, we should not be able to create any new pages.
	for i := pageCount; i < pageCount*2; i++ {
		util.Equals(t, page.Page(nil), pc.NewPage())
	}

	// Scenario 5: After unpinning pages {0, 1, 2, 3, 4} and pinning another 59 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, pc.UnpinPage(page.PageID(i), true))
		err := pc.FlushPage(page.PageID(i))
		if err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < 4; i++ {
		pc.NewPage()
	}

	fmt.Println(">>>\n", pc)

	// Scenario 6: We should be able to fetch the data we wrote a while ago.
	page0 = pc.FetchPage(page.PageID(0))
	rec2, err := page0.GetRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, page.NewRecord(page.R_NUM, page.R_STR, []byte{1}, []byte("Hello, World!")), rec2)

	// Scenario 7: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, pc.UnpinPage(page.PageID(0), true))

	pg := pc.NewPage()
	util.Equals(t, page.PageID(pageCount+4), pg.GetPageID())
	util.Equals(t, page.Page(nil), pc.NewPage())
	util.Equals(t, page.Page(nil), pc.FetchPage(page.PageID(0)))

	err = pc.FlushAll()
	if err != nil {
		t.Error(err)
	}

	err = pc.Close()
	if err != nil {
		t.Error(err)
	}

	if cleanUpWhenDone {
		err = os.Remove(filepath.Join(testDir, testFile))
		if err != nil {
			t.Error(err)
		}

		err = os.Remove(testDir)
		if err != nil {
			t.Error(err)
		}
	}

}

func TestPageCache_HitRate(t *testing.T) {

	pageCount := uint16(16)
	testDir := "testing"
	testFile := "page_cache_test.txt"

	pc, err := OpenBufferCacheManager(filepath.Join(testDir, testFile), pageCount)
	if err != nil {
		t.Errorf("opening buffer manager: %s\n", err)
	}
	go pc.monitor()

	var pageIDs []uint32

	for i := 0; i < 32; i++ {
		p := pc.NewPage()
		pid := p.GetPageID()
		pageIDs = append(pageIDs, pid)
		log.Printf("Creating a new page (page %d)\nAdding 32 records...\n", pid)
		for j := 0; j < 32; j++ {
			var rid *page.RecordID
			rid, err = p.AddRecord(
				page.NewRecord(
					page.R_NUM, page.R_STR, []byte{byte(j)},
					[]byte(fmt.Sprintf("this record-%d data for page-%d\n", j, i)),
				),
			)
			if err != nil {
				t.Errorf("Adding record %d to page %d failed: %s", rid, pid, err)
			}
			// log.Printf("Added record %v, to page %d\n", rid, pid)

		}
		log.Printf("Unpinning page %d\n", pid)
		err = pc.UnpinPage(pid, true)
		if err != nil {
			t.Errorf("Unpinning page %d failed: %s", pid, err)
		}
		log.Printf("Flushing page %d\n", pid)
		err = pc.FlushPage(pid)
		if err != nil {
			t.Errorf("Flushing page %d failed: %s", pid, err)
		}
		fmt.Println()
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println()
	for _, pid := range pageIDs[:10] {
		p := pc.FetchPage(pid)
		if p == nil {
			t.Errorf("Got nil page for pid %d\n", pid)
		}
	}

	for _, pid := range pageIDs {
		log.Printf("Fetching page %d\n", pid)
		p := pc.FetchPage(pid)
		if p == nil {
			t.Errorf("Got nil page for pid %d\n", pid)
		}
		time.Sleep(250 * time.Millisecond)
		log.Printf("Unpinning page %d\n", pid)
		err = pc.UnpinPage(pid, false)
		if err != nil {
			t.Errorf("Unpinning page %d failed: %s", pid, err)
		}
		time.Sleep(250 * time.Millisecond)
	}

	err = pc.FlushAll()
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
