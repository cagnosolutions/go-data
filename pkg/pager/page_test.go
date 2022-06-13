package pager

import (
	"fmt"
	"sync"
	"testing"
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
		pid:   3,
		magic: stUsed,
		slots: 0,
		lower: szHd,
		upper: szPg,
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
		pid:   4,
		magic: stFree,
		slots: 0,
		lower: szHd,
		upper: szPg,
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
	// fmt.Println(pg)
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
