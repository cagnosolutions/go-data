package engine

import (
	"fmt"
	"runtime"
	"testing"
)

func TestPage_Vacuum(t *testing.T) {
	var rids []*RecordID
	p := NewPage(1, P_USED)
	id, err := p.AddRecord(NewRecord(R_STR_STR, []byte("rec-01"), []byte("this is record 01")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-02"), []byte("this is record 02")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-03"), []byte("this is record 03")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-04"), []byte("this is record 04")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-05"), []byte("this is record 05")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-06"), []byte("this is record 06")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-07"), []byte("this is record 07")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-08"), []byte("this is record 08")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-09"), []byte("this is record 09")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-10"), []byte("this is record 10")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-11"), []byte("this is record 11")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-12"), []byte("this is record 12")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-13"), []byte("this is record 13")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-14"), []byte("this is record 14")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-15"), []byte("this is record 15")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-16"), []byte("this is record 16")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)
	id, err = p.AddRecord(NewRecord(R_STR_STR, []byte("rec-17"), []byte("this is record 17")))
	if err != nil {
		t.Error(err)
	}
	rids = append(rids, id)

	var nrids []*RecordID
	for i := range rids {
		if i%3 == 0 {
			err := p.DelRecord(rids[i])
			if err != nil {
				t.Error(err)
			}
		} else {
			nrids = append(nrids, rids[i])
		}
	}
	rids = nil
	fmt.Println("record ids:", nrids)

	fmt.Printf("%s\n", p.String())
	err = p.Vacuum()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", p.String())

	p.Clear()
	p = nil
	runtime.GC()
}
