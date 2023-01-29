package spage

import (
	"fmt"
	"log"
	"testing"
)

var recordIDs []*RecordID

func addRecords(pg *Page) []*RecordID {
	var recs []*RecordID
	for i := 0; i < 10; i++ {
		rec := fmt.Sprintf("this-is-record-%.6x", i)
		rid, err := pg.AddRecord([]byte(rec))
		if err != nil {
			panic(err)
		}
		recs = append(recs, rid)
	}
	return recs
}

func TestPage_AddRecord(t *testing.T) {
	pg := NewPage(1)
	for i := 0; i < 10; i++ {
		rec := fmt.Sprintf("this-is-record-%.6x", i)
		rid, err := pg.AddRecord([]byte(rec))
		if err != nil {
			t.Errorf("[Page] adding record: %s", err)
		}
		recordIDs = append(recordIDs, rid)
	}
}

func TestPage_AddManyRecords(t *testing.T) {
	const count = 128
	pg := NewPageSize(4<<10, 1)
	for i := 0; i < count; i++ {
		rec := fmt.Sprintf("record-%.6x", i)
		_, err := pg.AddRecord([]byte(rec))
		if err != nil {
			t.Errorf("[Page] adding record: %s", err)
		}
	}
	// fmt.Printf(
	// 	"lower_bounds=%d, upper_bounds=%d, slot_count=%d, free_space=%d\n",
	// 	pg.header.freeSpaceLower, pg.header.freeSpaceUpper, pg.header.slotCount, pg.header.FreeSpace(),
	// )
	// fmt.Println(hex.Dump(pg.data))
	fmt.Println(pg)

}

func TestPage_GetRecord(t *testing.T) {
	pg := NewPage(1)
	recs := addRecords(pg)
	for _, rid := range recs {
		rec, err := pg.GetRecord(rid)
		if err != nil {
			t.Errorf("[Page] getting record: %s", err)
		}
		fmt.Printf("%v: %q\n", rid, rec)
	}
}

func TestPage_DelRecord(t *testing.T) {
	pg := NewPage(1)
	log.Printf("adding records...\n")
	recs := addRecords(pg)
	log.Printf("getting records...\n")
	// get them to prove they are there
	for _, rid := range recs {
		rec, err := pg.GetRecord(rid)
		if err != nil {
			t.Errorf("[Page] getting record: %s", err)
		}
		fmt.Printf("%v: %q\n", rid, rec)
	}
	log.Printf("deleting records...\n")
	// attempt to delete half of them
	for i, rid := range recs {
		if i%2 == 0 {
			err := pg.DelRecord(rid)
			if err != nil {
				t.Errorf("[Page] deleting record: %s", err)
			}
		}
	}
	log.Printf("getting records (again)...\n")
	// get them again, to see if they are gone
	for i, rid := range recs {
		rec, err := pg.GetRecord(rid)
		if i%2 != 0 {
			if err != nil {
				t.Errorf("[Page] getting record: %s", err)
			}
			fmt.Printf("%v: %q\n", rid, rec)
		} else {
			fmt.Printf("%v: freed\n", rid)
		}
	}
}

func TestPage_Range(t *testing.T) {
	pg := NewPage(1)
	log.Println("adding records...")
	addRecords(pg)
	log.Println("ranging records...")
	pg.Range(
		func(rid *RecordID) bool {
			rec, err := pg.GetRecord(rid)
			if err != nil {
				t.Errorf("[Page] getting record: %s", err)
			}
			fmt.Printf("%v: %q\n", rid, rec)
			return true
		},
	)
}

func TestLinkPages(t *testing.T) {
	pg1 := NewPage(1)
	pg2 := NewPage(2)
	pg1 = pg1.Link(pg2)
	if pg1.NextID() != pg2.PageID() || pg2.PrevID() != pg1.PageID() {
		t.Errorf("[Page] linking pages: %s", "error linking pages")
	}
}

func Test_pageHeader_FreeSpace(t *testing.T) {
	pg := NewPage(1)
	estimatedUsedSpace := 1000 + (99 * pageSlotSize)
	estimatedFreeSpace := MaxRecordSize - estimatedUsedSpace
	for i := 0; i < 100; i++ {
		// write 10 bytes in each record
		_, err := pg.AddRecord([]byte("=========~"))
		if err != nil {
			t.Errorf("[Page] adding record: %s", err)
		}
	}
	fmt.Printf("%s\n", pg)
	if estimatedFreeSpace != int(pg.header.FreeSpace()) {
		t.Errorf(
			"[Page] estimatedFreeSpace=%d, actualFreeSpace=%d",
			estimatedFreeSpace, pg.header.FreeSpace(),
		)
	}

}
