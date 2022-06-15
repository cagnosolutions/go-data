package pager

import (
	"fmt"
	"sync"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/slicer"
)

var recs []*recID
var pg page

func TestPage_NewPage(t *testing.T) {
	if pg != nil {
		t.Errorf("got %v, expected %v\n", pg, nil)
	}
	pg = newPage(3)
	if pg == nil {
		t.Errorf("got %v, expected %v\n", len(pg), szPg)
	}
	tmp := header{
		pid:    3,
		meta:   mdSlotted | mdRecDynmc,
		status: stUsed,
		slots:  0,
		lower:  szHd,
		upper:  szPg,
	}
	hdr := pg.getHeader()
	if *hdr != tmp || hdr == nil {
		t.Errorf("got %v, expected %v\n", hdr, tmp)
	}
}

func TestPage_NewEmptyPage(t *testing.T) {
	var epg page
	if epg != nil {
		t.Errorf("got %v, expected %v\n", epg, nil)
	}
	epg = newEmptyPage(4)
	if epg == nil {
		t.Errorf("got %v, expected %v\n", len(epg), szPg)
	}
	tmp := header{
		pid:    4,
		meta:   mdSlotted | mdRecDynmc,
		status: stFree,
		slots:  0,
		lower:  szHd,
		upper:  szPg,
	}
	hdr := epg.getHeader()
	if *hdr != tmp || hdr == nil {
		t.Errorf("got %v, expected %v\n", hdr, tmp)
	}
}

func _addRecords(t *testing.T) {
	pg = newPage(3)
	id1, err := pg.addRecord([]byte("this is record number one"))
	if err != nil || id1 == nil {
		t.Errorf("got %v, expected %v\n", err, nil)
	}
	id2, err := pg.addRecord([]byte("this is record number two"))
	if err != nil || id2 == nil {
		t.Errorf("got %v, expected %v\n", err, nil)
	}
	id3, err := pg.addRecord([]byte("this is record number three"))
	if err != nil || id3 == nil {
		t.Errorf("got %v, expected %v\n", err, nil)
	}
	recs = append(recs, id1)
	recs = append(recs, id2)
	recs = append(recs, id3)
}

func TestPage_AddRecord(t *testing.T) {
	pg = newPage(3)
	err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(pg)
}

func TestPage_GetRecord(t *testing.T) {
	pg = newPage(3)
	err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	err = getRecords(pg)
	if err != nil {
		t.Error(err)
	}
}

func TestPage_DelRecord(t *testing.T) {
	pg = newPage(3)
	err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	sz := pg.size()
	if sz == 0 {
		t.Errorf("got %v, expected %v\n", sz, 3)
	}
	err = delRecords(pg)
	if err != nil {
		t.Error(err)
	}
	// for i := range recs {
	// 	err = pg.delRecord(recs[i])
	// 	if err != nil {
	// 		t.Errorf("got %v, expected %v\n", err, nil)
	// 	}
	// 	_, err := pg.getRecord(recs[i])
	// 	if err == nil {
	// 		t.Errorf("got %v, expected %v\n", "<error>", err)
	// 	}
	// }
}

var addRecords = func(p page) error {
	for i := 0; i < 128; i++ {
		rec := fmt.Sprintf("record-%6d", i)
		_, err := p.addRecord([]byte(rec))
		if err != nil {
			return err
		}
	}
	return nil
}

var getRecords = func(p page) error {
	for i := 0; i < 128; i++ {
		rid := &recID{
			pid: 3,
			sid: uint16(i),
		}
		_, err := p.getRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

var delRecords = func(p page) error {
	for i := 0; i < 128; i++ {
		rid := &recID{
			pid: 3,
			sid: uint16(i),
		}
		err := p.delRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestPage_Sync(t *testing.T) {
	pg = newPage(3)
	err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err := getRecords(pg)
		if err != nil {
			if err != ErrRecordNotFound {
				t.Error(err)
			}
		}
		wg.Done()
	}()
	go func() {
		err := delRecords(pg)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestPage_Rando(t *testing.T) {
	pageTests()
}

const N = 32

var ids []recID

func pageTests() {
	p := newPage(1)
	info(p)
	fmt.Println(">>>>> [01 ADDING] <<<<<")
	fmt.Printf("created page, adding %d records...\n", N)
	for i := 0; i < N; i++ {
		data := fmt.Sprintf("[record number %.2d]", i)
		id, err := p.addRecord([]byte(data))
		if err != nil {
			panic(err)
		}
		ids = append(ids, *id)
	}
	fmt.Println()
	info(p)
	fmt.Println(">>>>> [02 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.getRecord(&id)
		if err != nil {
			panic(err)
		}
		fmt.Printf("get(%v)=%q\n", id, rec)
	}
	fmt.Println()
	fmt.Println(">>>>> [03 DELETING] <<<<<")
	fmt.Printf("now, we will be removing some records...\n")
	for i, id := range ids {
		if (id.sid+1)%3 == 0 || id.sid == 31 {
			fmt.Printf("deleting record: %v\n", id)
			err := p.delRecord(&id)
			if err != nil {
				panic(err)
			}
			slicer.DelPtr(&ids, i)
		}
	}
	fmt.Println()
	info(p)
	fmt.Println(">>>>> [04 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.getRecord(&id)
		if err != nil {
			if err == ErrRecordNotFound {
				continue
			}
			panic(err)
		}
		fmt.Printf("get(%v)=%q\n", id, rec)
	}
	fmt.Println()
	fmt.Printf("taking a look at the page details...\n")
	info(p)
	fmt.Println(">>>>> [05 ADDING (9) MORE] <<<<<")
	for i := 32; i < N+8; i++ {
		data := fmt.Sprintf("[record number %.2d]", i)
		id, err := p.addRecord([]byte(data))
		if err != nil {
			panic(err)
		}
		ids = append(ids, *id)
	}
	id, err := p.addRecord([]byte("[large record that will not fit in existing space]"))
	if err != nil {
		panic(err)
	}
	ids = append(ids, *id)
	fmt.Println()
	info(p)
	fmt.Println(">>>>> [06 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.getRecord(&id)
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
	p = newPage(2)
	for i := 0; ; i++ {
		data := fmt.Sprintf("adding another record (%.2d)", i)
		_, err := p.addRecord([]byte(data))
		if err != nil {
			if err == ErrNoRoom {
				break
			}
			panic(err)
		}
	}
	fmt.Println()
	info(p)
	fmt.Println(">>>>> [08 COMPACTION] <<<<<")
	if err = p.compact(); err != nil {
		panic(err)
	}
	fmt.Println()
	info(p)
	fmt.Println()
}

func info(p page) {
	fmt.Println(p.DumpPage(false))
}
