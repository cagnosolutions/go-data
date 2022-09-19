package engine

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func TestPage_NewPage(t *testing.T) {
	var p page
	if p != nil {
		t.Errorf("got %v, expected %v\n", p, nil)
	}
	p = newPage(3, P_USED)
	if p == nil {
		t.Errorf("got %v, expected %v\n", len(p), PageSize)
	}
	tmp := pageHeader{
		ID:    3,
		Prev:  0,
		Next:  0,
		Flags: P_USED,
		Cells: 0,
		Free:  0,
		Lower: pageHeaderSize,
		Upper: PageSize,
	}
	hdr := p.getPageHeader()
	if *hdr != tmp || hdr == nil {
		t.Errorf("got %v, expected %v\n", hdr, tmp)
	}
}

func TestPage_NewEmptyPage(t *testing.T) {
	var ep page
	if ep != nil {
		t.Errorf("got %v, expected %v\n", ep, nil)
	}
	ep = newPage(4, P_FREE)
	if ep == nil {
		t.Errorf("got %v, expected %v\n", len(ep), PageSize)
	}
	tmp := pageHeader{
		ID:    4,
		Prev:  0,
		Next:  0,
		Flags: P_FREE,
		Cells: 0,
		Free:  0,
		Lower: pageHeaderSize,
		Upper: PageSize,
	}
	hdr := ep.getPageHeader()
	if *hdr != tmp || hdr == nil {
		t.Errorf("got %v, expected %v\n", hdr, tmp)
	}
}

func TestPage_addRecord(t *testing.T) {
	p := newPage(3, P_USED)
	_, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(p.String())
}

func TestPage_addRecordAndRange(t *testing.T) {
	p := newPage(3, P_USED)
	_, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	var i int
	err = p.rangeRecords(
		func(r *record) error {
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
	p := newPage(3, P_USED)
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
	p := newPage(3, P_USED)
	rids, err := addRecords(p)
	if err != nil {
		t.Error(err)
	}
	sz := p.size()
	if sz == 0 {
		t.Errorf("got %v, expected %v\n", sz, 3)
	}
	err = delRecords(p, rids)
	if err != nil {
		t.Error(err)
	}
}

func TestPage_Sync(t *testing.T) {
	p := newPage(3, P_USED)
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

var R_STR_STR uint32 = RK_STR | RV_STR

func TestPage_RandomStuff(t *testing.T) {
	const N = 32
	var ids []*RecordID

	makeRec := func(n int) record {
		rk := fmt.Sprintf("%.2d", n)
		rv := fmt.Sprintf("[record number %.2d]", n)
		return newRecord(R_STR_STR, []byte(rk), []byte(rv))
	}

	makeRecSize := func(n int, data []byte) record {
		rk := fmt.Sprintf("%.2d", n)
		rv := fmt.Sprintf("[record number %.2d %s]", n, string(data))
		return newRecord(R_STR_STR, []byte(rk), []byte(rv))
	}

	p := newPage(1, P_USED)
	fmt.Println(p.String())
	fmt.Println(">>>>> [01 ADDING] <<<<<")
	fmt.Printf("created Page, adding %d records...\n", N)
	for i := 0; i < N; i++ {
		id, err := p.addRecord(makeRec(i))
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
		rec, err := p.getRecord(id)
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
			err := p.delRecord(id)
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
			// ids = slicer.Del[*RecordID](ids, i)
			newids = append(newids, id)
		}
	}
	ids = newids
	fmt.Println()
	fmt.Println(p.String())
	fmt.Println(">>>>> [04 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.getRecord(id)
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
	fmt.Println(">>>>> [05 ADDING (12) MORE] <<<<<")
	for i := 32; i < N+13; i++ {
		id, err := p.addRecord(makeRec(i))
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}
	id, err := p.addRecord(makeRecSize(255, bytes.Repeat([]byte("A"), 13000)))
	if err != nil {
		panic(err)
	}
	ids = append(ids, id)
	fmt.Println()
	fmt.Println(p.String())
	fmt.Println(">>>>> [06 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.getRecord(id)
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
	p = newPage(2, P_USED)
	for i := 0; ; i++ {
		_, err := p.addRecord(makeRec(i))
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
	p := newPage(1, P_USED)
	id, err := p.addRecord(newRecord(R_STR_STR, []byte("rec-01"), []byte("this is record 01")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-02"), []byte("this is record 02")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-03"), []byte("this is record 03")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-04"), []byte("this is record 04")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-05"), []byte("this is record 05")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-06"), []byte("this is record 06")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-07"), []byte("this is record 07")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-08"), []byte("this is record 08")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-09"), []byte("this is record 09")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-10"), []byte("this is record 10")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-11"), []byte("this is record 11")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-12"), []byte("this is record 12")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-13"), []byte("this is record 13")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-14"), []byte("this is record 14")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-15"), []byte("this is record 15")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-16"), []byte("this is record 16")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.addRecord(newRecord(R_STR_STR, []byte("rec-17"), []byte("this is record 17")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)

	var deleted uint16
	for i := range rids {
		if i%3 == 0 {
			err := p.delRecord(rids[i])
			if err != nil {
				t.Error(err)
			}
			deleted++
		}
	}

	rcb := p.getNumCells()
	rcbf := p.getNumFree()
	p.Vacuum()
	rca := p.getNumCells()
	rcaf := p.getNumFree()

	if rcb == rca && rcb-deleted != rca {
		t.Errorf("total records: expected=%d, got=%d\n", rcb-deleted, rca)
	}

	if rcbf == rcaf {
		t.Errorf("records free: expected=%d, got=%d\n", rcb-deleted, rca)
	}

	p.clear()
	p = nil
	runtime.GC()
}

var addRecords = func(p page) ([]*RecordID, error) {
	var ids []*RecordID
	for i := 0; i < 128; i++ {
		rk := fmt.Sprintf("record-%.6d", i)
		rv := fmt.Sprintf("this is the value for record #%.6d", i)
		id, err := p.addRecord(newRecord(R_STR_STR, []byte(rk), []byte(rv)))
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

var getRecords = func(p page, ids []*RecordID) error {
	for i := range ids {
		rid := ids[i]
		_, err := p.getRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

var delRecords = func(p page, ids []*RecordID) error {
	for i := range ids {
		rid := ids[i]
		err := p.delRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}
