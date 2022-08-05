package buffer

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/dbms/errs"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestBufferPool_All(t *testing.T) {
	pageSize := uint16(page.DefaultPageSize)
	pageCount := uint16(10)
	testFile := "testing/bp_all_test.db"

	bpm := newBufferPool(testFile, pageSize, pageCount)

	page0 := page.Page(bpm.newPage())
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
		p := page.Page(bpm.newPage())
		util.Equals(t, page.PageID(i), p.GetPageID())
	}
	// log.Printf("[S3] >>> DONE")

	// Scenario 4: Once the buffer pool is full, we should not be able to create any new pages.
	for i := pageCount; i < pageCount*2; i++ {
		util.Equals(t, page.Page(nil), page.Page(bpm.newPage()))
	}
	// log.Printf("[S4] >>> DONE")

	// Scenario 5: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, bpm.unpinPage(page.PageID(i), true))
		// log.Println("attempting to flush page", i)
		err := bpm.flushPage(page.PageID(i))
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
	// log.Printf("[S5] >>> DONE")

	// Scenario 6: We should be able to fetch the data we wrote a while ago.
	page0 = page.Page(bpm.fetchPage(page.PageID(0)))
	rec2, err := page0.GetRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec2)
	// log.Printf("[S6] >>> DONE")

	// Scenario 7: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, bpm.unpinPage(page.PageID(0), true))

	pg := page.Page(bpm.newPage())
	util.Equals(t, page.PageID(14), pg.GetPageID())
	util.Equals(t, page.Page(nil), bpm.newPage())
	// fmt.Println(bpm)
	util.Equals(t, page.Page(nil), bpm.fetchPage(page.PageID(0)))
	// log.Printf("[S7] >>> DONE")

	err = bpm.flushAll()
	if err != nil {
		t.Error(err)
	}

	err = bpm.close()
	if err != nil {
		t.Error(err)
	}
	// err = cleanup(testFile)
	// if err != nil {
	//	t.Error(err)
	// }
	// time.Sleep(3 * time.Second)
}

/*
func _TestBufferPool_All(t *testing.T) {
	pageSize := DefaultPageSize
	pageCount := 10
	testFile := "testing/bp_all_test.db"

	dm, err := newDiskManager(testFile, uint32(pageSize), uint32(pageCount))
	if err != nil {
		t.Error(err)
	}
	bpm := newBufferPool(uint16(pageSize), pageCount, dm)

	page0 := bpm.newPage()
	// fmt.Println(page0)

	// Scenario 1: The buffer pool is empty. We should be able to create a new page.
	util.Equals(t, pageID(0), page0.getPageID())
	// log.Printf("[S1] >>> DONE")

	// Scenario 2: Once we have a page, we should be able to read and write content.
	id0, err := page0.addRecord([]byte("Hello, World!"))
	if err != nil {
		t.Error(err)
	}
	rec, err := page0.getRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec)
	// log.Printf("[S2] >>> DONE")

	// Scenario 3: We should be able to create new pages until we fill up the buffer pool.
	for i := 1; i < pageCount; i++ {
		p := bpm.newPage()
		util.Equals(t, pageID(i), p.getPageID())
	}
	// log.Printf("[S3] >>> DONE")

	// Scenario 4: Once the buffer pool is full, we should not be able to create any new pages.
	for i := pageCount; i < pageCount*2; i++ {
		util.Equals(t, page(nil), bpm.newPage())
	}
	// log.Printf("[S4] >>> DONE")

	// Scenario 5: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, bpm.unpinPage(pageID(i), true))
		// log.Println("attempting to flush page", i)
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
	// log.Printf("[S5] >>> DONE")

	// Scenario 6: We should be able to fetch the data we wrote a while ago.
	page0 = bpm.fetchPage(pageID(0))
	rec2, err := page0.getRecord(id0)
	if err != nil {
		t.Error(err)
	}
	util.Equals(t, []byte("Hello, World!"), rec2)
	// log.Printf("[S6] >>> DONE")

	// Scenario 7: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, bpm.unpinPage(pageID(0), true))

	pg := bpm.newPage()
	util.Equals(t, pageID(14), pg.getPageID())
	util.Equals(t, page(nil), bpm.newPage())
	// fmt.Println(bpm)
	util.Equals(t, page(nil), bpm.fetchPage(pageID(0)))
	// log.Printf("[S7] >>> DONE")

	err = bpm.flushAll()
	if err != nil {
		t.Error(err)
	}

	err = bpm.close()
	if err != nil {
		t.Error(err)
	}
	err = cleanup(testFile)
	if err != nil {
		t.Error(err)
	}
	// time.Sleep(3 * time.Second)
}
*/

var cleanup = func(testFile string) error {
	// remove test files
	err := os.RemoveAll(testFile)
	if err != nil {
		return err
	}
	return nil
}

var addBPRecords = func(bp *bufferPool, pid page.PageID) error {
	pg := page.Page(bp.fetchPage(pid))
	if pg == nil {
		return errs.ErrPageNotFound
	}
	for i := 0; i < 128; i++ {
		rec := fmt.Sprintf("record-%6d", i)
		_, err := pg.AddRecord([]byte(rec))
		if err != nil {
			return err
		}
	}
	return nil
}

var getBPRecords = func(bp *bufferPool, pid page.PageID) error {
	pg := page.Page(bp.fetchPage(pid))
	if pg == nil {
		return errs.ErrPageNotFound
	}
	for i := 0; i < 128; i++ {
		rid := &page.RecID{
			PID: pid,
			SID: uint16(i),
		}
		_, err := pg.GetRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

var delBPRecords = func(bp *bufferPool, pid page.PageID) error {
	pg := page.Page(bp.fetchPage(pid))
	if pg == nil {
		return errs.ErrPageNotFound
	}
	for i := 0; i < 128; i++ {
		rid := &page.RecID{
			PID: pid,
			SID: uint16(i),
		}
		err := pg.DelRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestBufferPool_Sync(t *testing.T) {
	pageSize := uint16(page.DefaultPageSize)
	pageCount := uint16(10)
	testFile := "testing/bp_race_test.db"

	bp := newBufferPool(testFile, pageSize, pageCount)
	_ = bp.newPage()
	err := addBPRecords(bp, 0)
	if err != nil {
		t.Error(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err := getBPRecords(bp, 0)
		if err != nil {
			if err != errs.ErrRecordNotFound {
				t.Error(err)
			}
		}
		wg.Done()
	}()
	go func() {
		err := delBPRecords(bp, 0)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()
	err = bp.flushAll()
	if err != nil {
		t.Error(err)
	}
	wg.Wait()
	err = bp.close()
	if err != nil {
		t.Error(err)
	}
	// err = cleanup(testFile)
	// if err != nil {
	//	t.Error(err)
	// }
}
