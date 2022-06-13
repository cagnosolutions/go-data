package pager

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestBufferPool(t *testing.T) {
	poolSize := 10

	testFile := "testing/diskmanager.db"

	dm := newDiskManager(testFile)
	defer dm.close()
	bpm := newBufferPool(poolSize, dm)

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
	for i := 1; i < poolSize; i++ {
		p := bpm.newPage()
		util.Equals(t, pageID(i), p.getPageID())
	}
	// log.Printf("[S3] >>> DONE")

	// Scenario 4: Once the buffer pool is full, we should not be able to create any new pages.
	for i := poolSize; i < poolSize*2; i++ {
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

	dm.close()
	// time.Sleep(3 * time.Second)

	// remove test files
	err = os.RemoveAll(testFile)
	if err != nil {
		t.Error(err)
	}
}

func TestBufferPool_Race(t *testing.T) {
	poolSize := 10
	testFile := "testing/diskmanager.db"

	dm := newDiskManager(testFile)
	defer dm.close()
	bpm := newBufferPool(poolSize, dm)

	pg := bpm.newPage()

	rids := make([]*recID, 0)

	addData := func() {
		for i := 0; i < 16; i++ {
			data := fmt.Sprintf("foo-bar-%.8d", i)
			rid, err := pg.addRecord([]byte(data))
			if err != nil {
				fmt.Println(">> DUMP (add)", rids)
				t.Error(err)
			}
			rids = append(rids, rid)
		}
	}

	getData := func() {
		for i := 0; i < 16; i++ {
			rid := getRandVal(&rids, false)
			if rid == nil {
				break
			}
			_, err = pg.getRecord(rid)
			if err != nil {
				fmt.Println(">> DUMP (get)", rids)
				t.Error(err, rid)
			}
		}
	}

	delData := func() {
		for i := 0; i < 16; i++ {
			mu.Lock()
			rid := getRandVal(&rids, true)
			if rid == nil {
				mu.Unlock()
				continue
			}
			mu.Unlock()
			err = pg.delRecord(rid)
			if err != nil {
				t.Error(err, rid)
			}
		}
	}

	for i := 0; i < 16; i++ {
		go addData()
		go getData()
	}

	delData()

	// err = bpm.flushAll()
	// if err != nil {
	//	t.Error(err)
	// }
}

func getRandVal(rr *[]*recID, remove bool) *recID {
	if len(*rr) == 0 {
		return nil
	}
	i := rand.Intn(len(*rr))
	if remove {
		var ret *recID
		ret = (*rr)[i]
		if i < len(*rr)-1 {
			copy((*rr)[i:], (*rr)[i+1:])
		}
		(*rr)[len(*rr)-1] = nil // or the zero value of T
		*rr = (*rr)[:len(*rr)-1]
		return ret
	}
	return (*rr)[i]
}

var addBPRecords = func(bp *bufferPool, pid pageID) error {
	pg := bp.fetchPage(pid)
	if pg == nil {
		return ErrPageNotFound
	}
	for i := 0; i < 128; i++ {
		rec := fmt.Sprintf("record-%6d", i)
		_, err = pg.addRecord([]byte(rec))
		if err != nil {
			return err
		}
	}
	return nil
}

var getBPRecords = func(bp *bufferPool, pid pageID) error {
	pg := bp.fetchPage(pid)
	if pg == nil {
		return ErrPageNotFound
	}
	for i := 0; i < 128; i++ {
		rid := &recID{
			pid: pid,
			sid: uint16(i),
		}
		_, err := pg.getRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

var delBPRecords = func(bp *bufferPool, pid pageID) error {
	pg := bp.fetchPage(pid)
	if pg == nil {
		return ErrPageNotFound
	}
	for i := 0; i < 128; i++ {
		rid := &recID{
			pid: pid,
			sid: uint16(i),
		}
		err := pg.delRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestBufferPool_Sync(t *testing.T) {
	poolSize := 10
	testFile := "testing/bp_race.db"
	dm := newDiskManager(testFile)
	defer dm.close()
	bp := newBufferPool(poolSize, dm)
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
			if err != ErrRecordNotFound {
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
}
