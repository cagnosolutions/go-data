package page

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestPage_HexDump(t *testing.T) {
	p := NewPage(3, P_USED)
	_, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(p.HexDump())
	// fmt.Println(hex.Dump(p))
	// os.WriteFile("pagedump.txt", p, 0666)
}

func TestPage_NewPage(t *testing.T) {
	var p Page
	if p != nil {
		t.Errorf("got %v, expected %v\n", p, nil)
	}
	p = NewPage(3, P_USED)
	if p == nil {
		t.Errorf("got %v, expected %v\n", len(p), PageSize)
	}
	tmp := PageHeader{
		ID:    3,
		Prev:  0,
		Next:  0,
		Flags: P_USED,
		Cells: 0,
		Free:  0,
		Lower: pageHeaderSize,
		Upper: PageSize,
	}
	hdr := p.GetPageHeader()
	if *hdr != tmp || hdr == nil {
		t.Errorf("got %v, expected %v\n", hdr, tmp)
	}
}

func TestPage_NewEmptyPage(t *testing.T) {
	var ep Page
	if ep != nil {
		t.Errorf("got %v, expected %v\n", ep, nil)
	}
	ep = NewPage(4, P_FREE)
	if ep == nil {
		t.Errorf("got %v, expected %v\n", len(ep), PageSize)
	}
	tmp := PageHeader{
		ID:    4,
		Prev:  0,
		Next:  0,
		Flags: P_FREE,
		Cells: 0,
		Free:  0,
		Lower: pageHeaderSize,
		Upper: PageSize,
	}
	hdr := ep.GetPageHeader()
	if *hdr != tmp || hdr == nil {
		t.Errorf("got %v, expected %v\n", hdr, tmp)
	}
}

func TestPage_addRecord(t *testing.T) {
	p := NewPage(3, P_USED)
	_, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(p.String())
}

func TestPage_addRecordAndRange(t *testing.T) {
	p := NewPage(3, P_USED)
	_, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	var i int
	err = p.RangeRecords(
		func(r *Record) error {
			fmt.Printf("record #%.3d, key=%q, val=%q\n", i, r.Key(), r.Val())
			i++
			return nil
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func TestPage_getRecord(t *testing.T) {
	p := NewPage(3, P_USED)
	rids, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	err = getRecords(p, rids)
	if err != nil {
		t.Error(err)
	}
}

func TestPage_DelRecord(t *testing.T) {
	p := NewPage(3, P_USED)
	rids, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	sz := p.Size()
	if sz == 0 {
		t.Errorf("got %v, expected %v\n", sz, 3)
	}
	err = delRecords(p, rids)
	if err != nil {
		t.Error(err)
	}
}

func TestPage_Sync(t *testing.T) {
	p := NewPage(3, P_USED)
	ids, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err := getRecords(p, ids)
		if err != nil {
			if err != ErrRecordNotFound {
				t.Error(err)
			}
		}
		wg.Done()
	}()
	go func() {
		err := delRecords(p, ids)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestPage_RandomStuff(t *testing.T) {
	const N = 32
	var ids []*RecordID

	makeRec := func(n int) Record {
		rk := fmt.Sprintf("%.2d", n)
		rv := fmt.Sprintf("[record number %.2d]", n)
		return NewRecord(R_STR, R_STR, []byte(rk), []byte(rv))
	}

	makeRecSize := func(n int, data []byte) Record {
		rk := fmt.Sprintf("%.2d", n)
		rv := fmt.Sprintf("[record number %.2d %s]", n, string(data))
		return NewRecord(R_STR, R_STR, []byte(rk), []byte(rv))
	}

	p := NewPage(1, P_USED)
	fmt.Println(p.String())
	fmt.Println(">>>>> [01 ADDING] <<<<<")
	fmt.Printf("created Page, adding %d records...\n", N)
	for i := 0; i < N; i++ {
		id, err := p.AddRecord(makeRec(i))
		if err != nil {
			panic(err)
		}
		// ids[i] = id
		ids = append(ids, id)
	}
	fmt.Println()
	fmt.Println(p.String())
	fmt.Println(">>>>> [02 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.GetRecord(id)
		if err != nil {
			panic(err)
		}
		fmt.Printf("get(%v)=%q\n", id, rec)
	}
	fmt.Println()
	fmt.Println(">>>>> [03 DELETING] <<<<<")
	fmt.Printf("now, we will be removing some records...\n")
	for _, id := range ids {
		if (id.CellID+1)%3 == 0 || id.CellID == 31 {
			fmt.Printf("deleting record: %v\n", id)
			err := p.DelRecord(id)
			if err != nil {
				panic(err)
			}
			// slicer.DelPtr(&ids, i)
		}
	}

	var newids []*RecordID
	for _, id := range ids {
		if (id.CellID+1)%3 != 0 && id.CellID != 31 {
			fmt.Printf("deleting from slice: %s\n", id)
			// remove = append(remove, id)
			// slicer.DelPtr(&ids, i)
			// ids = slicer.del[*RecordID](ids, i)
			newids = append(newids, id)
		}
	}
	ids = newids
	fmt.Println()
	fmt.Println(p.String())
	fmt.Println(">>>>> [04 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.GetRecord(id)
		if err != nil {
			if err == ErrRecordNotFound {
				continue
			}
			panic(err)
		}
		fmt.Printf("get(%v)=%q\n", id, rec)
	}
	fmt.Println()
	fmt.Printf("taking a look at the Page details...\n")
	fmt.Println(p.String())
	ph := p.GetPageHeader()
	adding := (N + 13) - N
	fmt.Printf(
		">>>>> (cells=%d, free=%d, adding %d records, cells_after=%d)\n",
		ph.Cells, ph.Free, adding, int(ph.Cells-ph.Free)+adding,
	)
	fmt.Printf(">>>>> [05 ADDING (%d) MORE] <<<<<\n", adding)
	for i := 32; i < N+13; i++ {
		id, err := p.AddRecord(makeRec(i))
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}
	id, err := p.AddRecord(makeRecSize(255, bytes.Repeat([]byte("A"), 13000)))
	if err != nil {
		panic(err)
	}
	ids = append(ids, id)
	fmt.Println()
	fmt.Println(p.String())
	fmt.Println(">>>>> [06 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.GetRecord(id)
		if err != nil {
			if err == ErrRecordNotFound {
				continue
			}
			panic(err)
		}
		fmt.Printf("get(%v)=%q\n", id, rec)
	}
	fmt.Println()
	fmt.Println(">>>>> [07 NEW PAGE] <<<<<")
	p = NewPage(2, P_USED)
	for i := 0; ; i++ {
		_, err := p.AddRecord(makeRec(i))
		if err != nil {
			if err == ErrNoRoom {
				fmt.Println(">>>>>+<<ErrNoRoom>>+<<<<<")
				break
			}
			panic(err)
		}
	}
	fmt.Println()
	fmt.Println(p.String())
	fmt.Println(">>>>> [08 COMPACTION] <<<<<")
	p.Vacuum()

	fmt.Println()
	fmt.Println(p.String())
	fmt.Println()
}

func TestPage_Vacuum(t *testing.T) {
	var rids []*RecordID
	p := NewPage(1, P_USED)
	id, err := p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-01"), []byte("this is record 01")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-02"), []byte("this is record 02")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-03"), []byte("this is record 03")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-04"), []byte("this is record 04")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-05"), []byte("this is record 05")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-06"), []byte("this is record 06")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-07"), []byte("this is record 07")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-08"), []byte("this is record 08")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-09"), []byte("this is record 09")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-10"), []byte("this is record 10")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-11"), []byte("this is record 11")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-12"), []byte("this is record 12")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-13"), []byte("this is record 13")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-14"), []byte("this is record 14")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-15"), []byte("this is record 15")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-16"), []byte("this is record 16")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR, R_STR, []byte("rec-17"), []byte("this is record 17")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)

	var deleted uint16
	for i := range rids {
		if i%3 == 0 {
			err := p.DelRecord(rids[i])
			if err != nil {
				t.Error(err)
			}
			deleted++
		}
	}

	rcb := p.GetNumCells()
	rcbf := p.GetNumFree()
	p.Vacuum()
	rca := p.GetNumCells()
	rcaf := p.GetNumFree()

	if rcb == rca && rcb-deleted != rca {
		t.Errorf("total records: expected=%d, got=%d\n", rcb-deleted, rca)
	}

	if rcbf == rcaf {
		t.Errorf("records free: expected=%d, got=%d\n", rcb-deleted, rca)
	}

	p.Clear()
	p = nil
	runtime.GC()
}

// make sure to run with `env GODEBUG=gctrace=1 godoc -http=:6060`

func TestPageGC(t *testing.T) {
	// Create some pages...
	var pages []Page
	var size int
	pages, size = createPages(2)
	fmt.Printf("Created %d pages totaling %dKB\n", len(pages), size)

	// Add them to a map, then watch the gc stats
	pool := make(map[uint32]Page, len(pages))
	for i := range pages {
		pool[pages[i].GetPageID()] = pages[i]
	}

	time.Sleep(1 * time.Minute)

	// Wait for user input before exiting...
	fmt.Println("Press any key to exit...")
	_, err := fmt.Scanln()
	if err != nil {
		return
	}
	runtime.GC()
}

var createPages = func(numPages int) (pages []Page, totalSize int) {
	for i := 0; i < numPages; i++ {
		pages = append(pages, NewPage(uint32(i), P_USED))
		totalSize += 16
	}
	return pages, totalSize
}

var addRecords = func(p Page) ([]*RecordID, error) {
	var ids []*RecordID
	for i := 0; i < 128; i++ {
		rk := fmt.Sprintf("record-%.6d", i)
		rv := fmt.Sprintf("this is the value for record #%.6d", i)
		id, err := p.AddRecord(NewRecord(R_STR, R_STR, []byte(rk), []byte(rv)))
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

var getRecords = func(p Page, ids []*RecordID) error {
	for i := range ids {
		rid := ids[i]
		_, err := p.GetRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

var delRecords = func(p Page, ids []*RecordID) error {
	for i := range ids {
		rid := ids[i]
		err := p.DelRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}
