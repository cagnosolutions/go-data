package pager

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cagnosolutions/go-data/pkg/slicer"
	"github.com/cagnosolutions/go-data/pkg/util"
)

var recs []*recID
var pg page

func dataUsed(count, size int) int {
	return szHd + (count * szSl) + (count * size)
}

func printDetails(pgsize, datused int) {
	fmt.Printf(
		"pageSize=%d, used=%d, unused=%d, percent=%.2f%%\n", pgsize, datused, pgsize-datused,
		(float32(datused)/float32(pgsize))*100,
	)
}

func printInsertDetails(count, size int) {
	fmt.Printf(
		">> Inserting %d records that are about %d bytes each, amounting to around %.2f kB of total data.\n",
		count, size, float32(count*size)/(1<<10),
	)
}

func TestPage_ForStupidErrorsCuzYouNeverReallyKnowIfYouDidSomethingDumb(t *testing.T) {
	// do stuff
	p := newPage(56)
	p.printHeader()
	time.Sleep(1 * time.Second)
	_, _ = p.addRecord([]byte("foobarbaz-000"))
	time.Sleep(1 * time.Second)
	p.printHeader()
}

func TestPage_FillPercent(t *testing.T) {

	maxRecSize := 500
	var added int

	tw := util.NewTableWriter("rec_size", "rec_count", "percent_full", "bytes_unused")

	for rsize := 10; rsize < maxRecSize; rsize += 10 {
		// create a new page
		newp := newPage(3)
		// start looping to see how many records we can add
		for rcount := 0; err == nil; rcount++ {
			// if we have enough free space, add another record
			err := newp.checkRecord(uint16(rsize))
			if err != nil {
				// no more room!
				if err == ErrNoRoom {
					break
				}
				t.Errorf("record check failed: %q\n", err)
			}
			// add a record
			rec := fmt.Sprintf("%s-%.3d", util.RandBytesN(rsize-4), rcount)
			_, err = newp.addRecord([]byte(rec))
			if err != nil {
				t.Errorf("adding record failed: %q\n", err)
				break
			}
			added++
		}

		tw.WriteRow(rsize, added, newp.FillPercent(), newp.freeSpace())
		// fmt.Printf(
		// 	"rsize=%d, rcount=%d, percentFull=%.2f, free=%d\n",
		// 	rsize, added, newp.FillPercent(), newp.freeSpace(),
		// )
		tw.Flush()
		// wipe page
		newp.clear()
	}

	/*
			Notes for record data using a 4KB page.
		====================================================
		rec size	rec count	  percent full 	bytes unused
		----------------------------------------------------
			10			254			99.80469		8
			20			410			99.60			16
			30			523			99.90			4
			40			611			99.41			24
			50			683			99.02			40
			60			744			98.87			46
			70			797			98.92			44
			80			844			99.26			30
			90			886			99.02			40
			100			924			98.92			44
			110			959			99.70			12
			120			991			99.02			40
			130			1020		96.87			128
			140			1047		96.82			130
			150			1073		99.60			16
			160			1097		97.85			88
			170			1120		99.41			24
			180			1141		95.94			166
			190			1161		96.28			152
			200			1180		96.14			158
			210			1198		95.50			184
			220			1216		99.90			4
			230			1233		98.53			60
			240			1249		96.67			136
			250			1264		94.33			232
			260			1279		97.99			82
			270			1293		94.92			208
			280			1307		98.33			68
			290			1320		94.53			224
			300			1333		97.70			94
			310			1345		93.16			280
			320			1357		96.09			160
			330			1369		99.02			40
			340			1380		93.50			266
			350			1391		96.19			156
			360			1402		98.87			46
			370			1412		92.38			312
			380			1422		94.82			212
			390			1432		97.26			112
			400			1442		99.70			12
			410			1451		91.99			328
			420			1460		94.18			238
			430			1469		96.38			148
			440			1478		98.58			58
			450			1486		89.64			424
			460			1494		91.60			344
			470			1502		93.55			264
			480			1510		95.50			184
			490			1518		97.46			104





			+-----------+----------+--------------+--------------+
		|      64   |     64   |    99.71%    |
		+-----------+----------+--------------+--------------+
		|     128   |     30   |    98.73%    |       52
		+-----------+----------+--------------+--------------+
		|     256   |     15   |    96.53%    |      142
		+-----------+----------+--------------+--------------+
		|     512   |      7   |    89.11%    |      446
		+-----------+----------+--------------+--------------+
		|    1024   |      3   |    76.03%    |      982
		+-----------+----------+--------------+--------------+
		|    2000   |      2   |    98.54%    |		  60
		+-----------+----------+--------------+--------------+
		|    2048   |      1   |    50.73%    |		2018
		+-----------+----------+--------------+--------------+
		|    4000   |      1   |    98.39%    |		  66
		+-----------+----------+--------------+--------------+
		- 15 records of a length of 255 bytes each results in a 96% fill ratio
		-
	*/
	// rsize := 10
	// rcount := 1
	// for rsize := 10; rsize < 300; rsize += 5 {
	// 	rcpg := newPage(3)
	// 	rcount := make([]*recID, 0)
	// 	for i := 0; err == nil; i++ {
	// 		time.Sleep(3 * time.Millisecond)
	// 		rec := fmt.Sprintf("%s-%.3d", util.RandBytesN(rsize-4), i)
	// 		id, err := rcpg.addRecord([]byte(rec))
	// 		if err != nil {
	// 			break
	// 		}
	// 		rcount = append(rcount, id)
	// 	}
	// 	fmt.Printf(
	// 		"rsize=%d, rcount=%d, percentFull=%.2f, free=%d\n",
	// 		rsize, len(rcount), rcpg.FillPercent(), rcpg.freeSpace(),
	// 	)
	// 	rcount = nil
	// 	rcpg = nil
	// }
}

func TestPage_NewPage(t *testing.T) {
	if pg != nil {
		t.Errorf("got %v, expected %v\n", pg, nil)
	}
	pg = newPage(3)
	if pg == nil {
		t.Errorf("got %v, expected %v\n", len(pg), szPg)
	}
	tmp := header{
		pid:      3,
		size:     szPg,
		reserved: 0,
		meta:     mdSlotted | mdRecDynmc,
		status:   stUsed,
		slots:    0,
		lower:    szHd,
		upper:    szPg,
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
		pid:      4,
		size:     szPg,
		reserved: 0,
		meta:     mdSlotted | mdRecDynmc,
		status:   stFree,
		slots:    0,
		lower:    szHd,
		upper:    szPg,
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

var addRecordsSize = func(p page, count, size int) error {
	if size < 5 {
		return errors.New("size must be at least 8")
	}
	for i := 0; i < count; i++ {
		rec := fmt.Sprintf("%s-%.3d", util.RandBytesN(size-4), i)
		_, err := p.addRecord([]byte(rec))
		if err != nil {
			return err
		}
	}
	return nil
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
