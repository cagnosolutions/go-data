package page

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cagnosolutions/go-data/pkg/slicer"
	"github.com/cagnosolutions/go-data/pkg/util"
)

var recs []*RecID
var pg Page

func dataUsed(count, size int) int {
	return pageHeaderSize + (count * pageCellSize) + (count * size)
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

func TestPage_Flags(t *testing.T) {
	// setup options for the record
	ropts := RK_STR | RV_NUM | R_OVERFLOW
	fmt.Println(ropts&RK_NUM, "(should not be set)")
	fmt.Println(ropts&RK_STR, "(should be set)")
	fmt.Println(ropts&RV_NUM, "(should be set)")
	fmt.Println(ropts&RV_STR, "(should not be set)")
	fmt.Println(ropts&R_OVERFLOW, "(should be set)")
}

func TestPage_ForStupidErrorsCuzYouNeverReallyKnowIfYouDidSomethingDumb(t *testing.T) {
	// do stuff
	p := NewPage(56)
	p.printHeader()
	time.Sleep(1 * time.Second)
	_, _ = p.AddRecord([]byte("foobarbaz-000"))
	time.Sleep(1 * time.Second)
	p.printHeader()
}

func _TestPage_FillPercent(t *testing.T) {

	maxRecSize := 100
	var added int

	// tw := util.NewTableWriter("rec_size", "rec_count", "percent_full", "bytes_unused")
	fmt.Printf("rsize\trcount\t%%full\tnumFree\n")

	var err error
	for rsize := 10; rsize < maxRecSize; rsize += 10 {
		// create a new Page
		newp := NewPage(3)
		// start looping to see how many records we can add
		for rcount := 0; err == nil; rcount++ {
			// if we have enough numFree space, add another record
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
			_, err = newp.AddRecord([]byte(rec))
			if err != nil {
				t.Errorf("adding record failed: %q\n", err)
				break
			}
			added++
		}

		// tw.WriteRow(rsize, added, newp.FillPercent(), newp.freeSpace())
		fmt.Printf(
			"%.4d\t\t%.4d\t\t%.2f\t%.4d\n",
			rsize, added, newp.FillPercent(), newp.freeSpace(),
		)
		// tw.Flush()
		// wipe Page
		added = 0
		newp.clear()
	}

	/*
			Record data using a 4KB Page.
		====================================
		rsize		rcount		%full	numFree
		------------------------------------
		0010		0254		99.80	0008
		0020		0156		99.61	0016
		0030		0113		99.90	0004
		0040		0088		99.41	0024
		0050		0072		99.02	0040
		0060		0061		98.88	0046
		0070		0053		98.93	0044
		0080		0047		99.27	0030
		0090		0042		99.02	0040
		0100		0038		98.93	0044
		0100		0038		98.93	0044
		0150		0026		99.61	0016
		0200		0019		96.14	0158
		0250		0015		94.34	0232
		0300		0013		97.71	0094
		0350		0011		96.19	0156
		0400		0010		99.71	0012
		0450		0008		89.65	0424
		0500		0008		99.41	0024
		0550		0007		95.61	0180
		0600		0006		89.36	0436
		0650		0006		96.68	0136
		0700		0005		86.77	0542
		0750		0005		92.87	0292
		0800		0005		98.97	0042
		0850		0004		84.18	0648
		0900		0004		89.06	0448
		0950		0004		93.95	0248
		1000		0004		98.83	0048
		1050		0003		77.93	0904
		1100		0003		81.59	0754
		1150		0003		85.25	0604
		1200		0003		88.92	0454
		1250		0003		92.58	0304
		1300		0003		96.24	0154
		1350		0003		99.90	0004
		1400		0002		69.24	1260
		1450		0002		71.68	1160
		1500		0002		74.12	1060
		1550		0002		76.56	0960
		1600		0002		79.00	0860
		1650		0002		81.45	0760
		1700		0002		83.89	0660
		1750		0002		86.33	0560
		1800		0002		88.77	0460
		1850		0002		91.21	0360
		1900		0002		93.65	0260
		1950		0002		96.09	0160
		2000		0002		98.54	0060
		2050		0001		50.78	2016
		2100		0001		52.00	1966
		2150		0001		53.22	1916
		2200		0001		54.44	1866
		2250		0001		55.66	1816
		2300		0001		56.88	1766
		2350		0001		58.11	1716
		2400		0001		59.33	1666
		2450		0001		60.55	1616
		2500		0001		61.77	1566
		2550		0001		62.99	1516
		2600		0001		64.21	1466
		2650		0001		65.43	1416
		2700		0001		66.65	1366
		2750		0001		67.87	1316
		2800		0001		69.09	1266
		2850		0001		70.31	1216
		2900		0001		71.53	1166
		2950		0001		72.75	1116
		3000		0001		73.97	1066
		3050		0001		75.20	1016
		3100		0001		76.42	0966
		3150		0001		77.64	0916
		3200		0001		78.86	0866
		3250		0001		80.08	0816
		3300		0001		81.30	0766
		3350		0001		82.52	0716
		3400		0001		83.74	0666
		3450		0001		84.96	0616
		3500		0001		86.18	0566
		3550		0001		87.40	0516
		3600		0001		88.62	0466
		3650		0001		89.84	0416
		3700		0001		91.06	0366
		3750		0001		92.29	0316
		3800		0001		93.51	0266
		3850		0001		94.73	0216
		3900		0001		95.95	0166
		3950		0001		97.17	0116
		4000		0001		98.39	0066

			Record data using a 8KB Page.
		====================================
		rsize		rcount		%full	numFree
		------------------------------------
		0010		0510		99.90	0008
		0020		0314		99.95	0004
		0030		0227		100.00	0000
		0040		0177		99.68	0026
		0050		0145		99.41	0048
		0060		0123		99.39	0050
		0070		0107		99.56	0036
		0080		0095		100.00	0000
		0090		0085		99.90	0008
		0050		0145		99.41	0048
		0100		0077		99.93	0006
		0150		0052		99.32	0056
		0200		0039		98.36	0134
		0250		0031		97.17	0232
		0300		0026		97.41	0212
		0350		0022		95.90	0336
		0400		0020		99.41	0048
		0450		0017		94.92	0416
		0500		0016		99.12	0072
		0550		0014		95.31	0384
		0600		0013		96.46	0290
		0650		0012		96.39	0296
		0700		0011		95.09	0402
		0750		0010		92.58	0608
		0800		0010		98.68	0108
		0850		0009		94.34	0464
		0900		0009		99.83	0014
		0950		0008		93.65	0520
		1000		0008		98.54	0120
		1050		0007		90.53	0776
		1100		0007		94.80	0426
		1150		0007		99.07	0076
		1200		0006		88.62	0932
		1250		0006		92.29	0632
		1300		0006		95.95	0332
		1350		0006		99.61	0032
		1400		0005		86.11	1138
		1450		0005		89.16	0888
		1500		0005		92.21	0638
		1550		0005		95.26	0388
		1600		0005		98.32	0138
		1650		0004		81.15	1544
		1700		0004		83.59	1344
		1750		0004		86.04	1144
		1800		0004		88.48	0944
		1850		0004		90.92	0744
		1900		0004		93.36	0544
		1950		0004		95.80	0344
		2000		0004		98.24	0144
		2050		0003		75.59	2000
		2100		0003		77.42	1850
		2150		0003		79.25	1700
		2200		0003		81.08	1550
		2250		0003		82.91	1400
		2300		0003		84.74	1250
		2350		0003		86.57	1100
		2400		0003		88.40	0950
		2450		0003		90.23	0800
		2500		0003		92.07	0650
		2550		0003		93.90	0500
		2600		0003		95.73	0350
		2650		0003		97.56	0200
		2700		0003		99.39	0050
		2750		0002		67.58	2656
		2800		0002		68.80	2556
		2850		0002		70.02	2456
		2900		0002		71.24	2356
		2950		0002		72.46	2256
		3000		0002		73.68	2156
		3050		0002		74.90	2056
		3100		0002		76.12	1956
		3150		0002		77.34	1856
		3200		0002		78.56	1756
		3250		0002		79.79	1656
		3300		0002		81.01	1556
		3350		0002		82.23	1456
		3400		0002		83.45	1356
		3450		0002		84.67	1256
		3500		0002		85.89	1156
		3550		0002		87.11	1056
		3600		0002		88.33	0956
		3650		0002		89.55	0856
		3700		0002		90.77	0756
		3750		0002		91.99	0656
		3800		0002		93.21	0556
		3850		0002		94.43	0456
		3900		0002		95.65	0356
		3950		0002		96.88	0256
		4000		0002		98.10	0156
		4050		0002		99.32	0056
		4100		0001		50.42	4062
		4150		0001		51.03	4012
		4200		0001		51.64	3962
		4250		0001		52.25	3912
		4300		0001		52.86	3862
		4350		0001		53.47	3812
		4400		0001		54.08	3762
		4450		0001		54.69	3712
		4500		0001		55.30	3662
		4550		0001		55.91	3612
		4600		0001		56.52	3562
		4650		0001		57.13	3512
		4700		0001		57.74	3462
		4750		0001		58.35	3412
		4800		0001		58.96	3362
		4850		0001		59.57	3312
		4900		0001		60.18	3262
		4950		0001		60.79	3212
		5000		0001		61.40	3162
		5050		0001		62.01	3112
		5100		0001		62.62	3062
		5150		0001		63.23	3012
		5200		0001		63.84	2962
		5250		0001		64.45	2912
		5300		0001		65.06	2862
		5350		0001		65.67	2812
		5400		0001		66.28	2762
		5450		0001		66.89	2712
		5500		0001		67.50	2662
		5550		0001		68.12	2612
		5600		0001		68.73	2562
		5650		0001		69.34	2512
		5700		0001		69.95	2462
		5750		0001		70.56	2412
		5800		0001		71.17	2362
		5850		0001		71.78	2312
		5900		0001		72.39	2262
		5950		0001		73.00	2212
		6000		0001		73.61	2162
		6050		0001		74.22	2112
		6100		0001		74.83	2062
		6150		0001		75.44	2012
		6200		0001		76.05	1962
		6250		0001		76.66	1912
		6300		0001		77.27	1862
		6350		0001		77.88	1812
		6400		0001		78.49	1762
		6450		0001		79.10	1712
		6500		0001		79.71	1662
		6550		0001		80.32	1612
		6600		0001		80.93	1562
		6650		0001		81.54	1512
		6700		0001		82.15	1462
		6750		0001		82.76	1412
		6800		0001		83.37	1362
		6850		0001		83.98	1312
		6900		0001		84.59	1262
		6950		0001		85.21	1212
		7000		0001		85.82	1162
		7050		0001		86.43	1112
		7100		0001		87.04	1062
		7150		0001		87.65	1012
		7200		0001		88.26	0962
		7250		0001		88.87	0912
		7300		0001		89.48	0862
		7350		0001		90.09	0812
		7400		0001		90.70	0762
		7450		0001		91.31	0712
		7500		0001		91.92	0662
		7550		0001		92.53	0612
		7600		0001		93.14	0562
		7650		0001		93.75	0512
		7700		0001		94.36	0462
		7750		0001		94.97	0412
		7800		0001		95.58	0362
		7850		0001		96.19	0312
		7900		0001		96.80	0262
		7950		0001		97.41	0212
		8000		0001		98.02	0162
		8050		0001		98.63	0112
	*/
}

func TestPage_NewPage(t *testing.T) {
	if pg != nil {
		t.Errorf("got %v, expected %v\n", pg, nil)
	}
	pg = NewPage(3)
	if pg == nil {
		t.Errorf("got %v, expected %v\n", len(pg), pageSize)
	}
	tmp := pageHeader{
		pid:      3,
		prev:     0,
		next:     0,
		flags:    P_USED,
		numFree:  0,
		numCells: 0,
		lower:    pageHeaderSize,
		upper:    pageSize,
	}
	hdr := pg.getPageHeader()
	if *hdr != tmp || hdr == nil {
		t.Errorf("got %v, expected %v\n", hdr, tmp)
	}
}

func TestPage_NewEmptyPage(t *testing.T) {
	var epg Page
	if epg != nil {
		t.Errorf("got %v, expected %v\n", epg, nil)
	}
	epg = NewEmptyPage(4)
	if epg == nil {
		t.Errorf("got %v, expected %v\n", len(epg), pageSize)
	}
	tmp := pageHeader{
		pid:      4,
		prev:     0,
		next:     0,
		flags:    P_FREE,
		numFree:  0,
		numCells: 0,
		lower:    pageHeaderSize,
		upper:    pageSize,
	}
	hdr := epg.getPageHeader()
	if *hdr != tmp || hdr == nil {
		t.Errorf("got %v, expected %v\n", hdr, tmp)
	}
}

func _addRecords(t *testing.T) {
	pg = NewPage(3)
	id1, err := pg.AddRecord([]byte("this is record number one"))
	if err != nil || id1 == nil {
		t.Errorf("got %v, expected %v\n", err, nil)
	}
	id2, err := pg.AddRecord([]byte("this is record number two"))
	if err != nil || id2 == nil {
		t.Errorf("got %v, expected %v\n", err, nil)
	}
	id3, err := pg.AddRecord([]byte("this is record number three"))
	if err != nil || id3 == nil {
		t.Errorf("got %v, expected %v\n", err, nil)
	}
	recs = append(recs, id1)
	recs = append(recs, id2)
	recs = append(recs, id3)
}

func TestPage_AddRecord(t *testing.T) {
	pg = NewPage(3)
	_, err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(pg)
}

func TestPage_AddRecordAndIterate(t *testing.T) {
	pg = NewPage(3)
	_, err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	iter, err := pg.newIter(true)
	if err != nil {
		t.Error(err)
	}
	var i int
	for cp := iter.next(); iter.hasMore(); cp = iter.next() {
		beg, end := cp.bounds()
		fmt.Printf("record #%.3d, data=%q\n", i, string(pg[beg:end]))
		i++
	}
}

func TestPage_GetRecord(t *testing.T) {
	pg = NewPage(3)
	rids, err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	err = getRecords(pg, rids)
	if err != nil {
		t.Error(err)
	}
}

func TestPage_DelRecord(t *testing.T) {
	pg = NewPage(3)
	rids, err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	sz := pg.Size()
	if sz == 0 {
		t.Errorf("got %v, expected %v\n", sz, 3)
	}
	err = delRecords(pg, rids)
	if err != nil {
		t.Error(err)
	}
	// for i := range recs {
	// 	err = pg.delRecord(recs[i])
	// 	if err != nil {
	// 		t.Errorf("got %v, expected %v\n", err, nil)
	// 	}
	// 	_, err := pg.GetRecord(recs[i])
	// 	if err == nil {
	// 		t.Errorf("got %v, expected %v\n", "<error>", err)
	// 	}
	// }
}

var addRecordsSize = func(p Page, count, size int) error {
	if size < 5 {
		return errors.New("prev must be at least 8")
	}
	for i := 0; i < count; i++ {
		rec := fmt.Sprintf("%s-%.3d", util.RandBytesN(size-4), i)
		_, err := p.AddRecord([]byte(rec))
		if err != nil {
			return err
		}
	}
	return nil
}

var addRecords = func(p Page) ([]*RecID, error) {
	var ids []*RecID
	for i := 0; i < 128; i++ {
		rec := fmt.Sprintf("record-%6d", i)
		id, err := p.AddRecord([]byte(rec))
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

var getRecords = func(p Page, ids []*RecID) error {
	for i := range ids {
		rid := ids[i]
		_, err := p.GetRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

var delRecords = func(p Page, ids []*RecID) error {
	for i := range ids {
		rid := ids[i]
		err := p.delRecord(rid)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestPage_Sync(t *testing.T) {
	pg = NewPage(3)
	ids, err := addRecords(pg)
	if err != nil {
		t.Error(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err := getRecords(pg, ids)
		if err != nil {
			if err != ErrRecordNotFound {
				t.Error(err)
			}
		}
		wg.Done()
	}()
	go func() {
		err := delRecords(pg, ids)
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

var ids []RecID

func pageTests() {
	p := NewPage(1)
	info(p)
	fmt.Println(">>>>> [01 ADDING] <<<<<")
	fmt.Printf("created Page, adding %d records...\n", N)
	for i := 0; i < N; i++ {
		data := fmt.Sprintf("[record number %.2d]", i)
		id, err := p.AddRecord([]byte(data))
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
		rec, err := p.GetRecord(&id)
		if err != nil {
			panic(err)
		}
		fmt.Printf("get(%v)=%q\n", id, rec)
	}
	fmt.Println()
	fmt.Println(">>>>> [03 DELETING] <<<<<")
	fmt.Printf("now, we will be removing some records...\n")
	for i, id := range ids {
		if (id.CID+1)%3 == 0 || id.CID == 31 {
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
		rec, err := p.GetRecord(&id)
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
	info(p)
	fmt.Println(">>>>> [05 ADDING (9) MORE] <<<<<")
	for i := 32; i < N+8; i++ {
		data := fmt.Sprintf("[record number %.2d]", i)
		id, err := p.AddRecord([]byte(data))
		if err != nil {
			panic(err)
		}
		ids = append(ids, *id)
	}
	id, err := p.AddRecord([]byte("[large record that will not fit in existing space]"))
	if err != nil {
		panic(err)
	}
	ids = append(ids, *id)
	fmt.Println()
	info(p)
	fmt.Println(">>>>> [06 GETTING] <<<<<")
	fmt.Printf("now, we will be getting the records...\n")
	for _, id := range ids {
		rec, err := p.GetRecord(&id)
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
	p = NewPage(2)
	for i := 0; ; i++ {
		data := fmt.Sprintf("adding another record (%.2d)", i)
		_, err := p.AddRecord([]byte(data))
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

func info(p Page) {
	fmt.Println(p.DumpPage(false))
}

type records struct {
	index   map[int]int
	records []*RecID
}

func initRecords() *records {
	return &records{
		index:   make(map[int]int),
		records: make([]*RecID, 0),
	}
}

func (r *records) add(id int, rid *RecID) {
	for i := 0; i < len(r.records); i++ {
		if r.records[i] == nil {
			r.records[i] = rid
			r.index[id] = i
			return
		}
	}
	r.records = append(r.records, rid)
	r.index[id] = len(r.records) - 1
}

func (r *records) get(id int) *RecID {
	i, found := r.index[id]
	if !found {
		return nil
	}
	return r.records[i]
}

func (r *records) del(id int) {
	i, found := r.index[id]
	if !found {
		panic("could not locate or remove record")
	}
	// if i < len(r.records)-1 {
	//	copy(r.records[i:], r.records[i+1:])
	// }
	// r.records[len(r.records)-1] = nil // or the zero value of T
	// r.records = r.records[:len(r.records)-1]
	r.records[i] = nil
	delete(r.index, id)
}

func (r *records) String() string {
	var ss string
	for k, v := range r.index {
		ss += fmt.Sprintf("{k:%d, v:%d}", k, v)
	}
	// for i := range r.records {
	// 	ss += fmt.Sprintf("[%d]", r.records[i].CID)
	// }
	return ss
}

type testRecord struct {
	id   int
	data []byte
}

func getSlotAndRecsFromPage(pg Page) string {
	iter, err := pg.newIter(false)
	if err != nil {
		panic(err)
	}
	var ss []string
	// var CID int
	for cp := iter.next(); iter.hasMore(); cp = iter.next() {
		flags := cp.getFlags()
		beg, end := cp.bounds()
		ss = append(ss, fmt.Sprintf("numFree=%v, val=%.1x", flags, string(pg[beg:end])))
		//	CID++
	}
	return strings.Join(ss, ",")
}

func TestPage_IterateOrder(t *testing.T) {

	pg := NewPage(1)
	recs := initRecords()

	fmt.Println("inserting...")
	ins1 := []testRecord{
		{1, []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}}, // 1
		{2, []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}}, // 2
		{4, []byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44}}, // 4
		{6, []byte{0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66}}, // 6
		{8, []byte{0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88}}, // 8
	}
	for i := range ins1 {
		r, err := pg.AddRecord(ins1[i].data)
		if err != nil {
			t.Errorf("something went wrong (ins1): %s", err)
		}
		recs.add(ins1[i].id, r)
	}
	// should contain: 1,2,4,6,8
	fmt.Printf("should contain: 1,2,4,6,8\nrecords: %s\nindex: %s\n\n", getSlotAndRecsFromPage(pg), recs)

	fmt.Println("deleting...")
	del1 := []testRecord{
		{2, []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}}, // 2
		{1, []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}}, // 1
	}
	for i := range del1 {
		r := recs.get(del1[i].id)
		err := pg.delRecord(r)
		if err != nil {
			t.Errorf("something went wrong (del1): %s", err)
		}
		recs.del(del1[i].id)
	}
	// should contain: X,X,4,6,8
	fmt.Printf("should contain: X,X,4,6,8\nrecords: %s\nindex: %s\n\n", getSlotAndRecsFromPage(pg), recs)

	fmt.Println("inserting...")
	ins2 := []testRecord{
		{1, []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}}, // 1
		{3, []byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}}, // 3
		{5, []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}}, // 5
		{7, []byte{0x77, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77}}, // 7
		{9, []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99}}, // 9
	}
	for i := range ins2 {
		r, err := pg.AddRecord(ins2[i].data)
		if err != nil {
			t.Errorf("something went wrong (ins2): %s", err)
		}
		recs.add(ins2[i].id, r)
	}
	// should contain: 1,3,4,6,8,5,7,9
	fmt.Printf("should contain: 1,3,4,6,8,5,7,9\nrecords: %s\nindex: %s\n\n", getSlotAndRecsFromPage(pg), recs)

	fmt.Println("deleting...")
	del2 := []testRecord{
		{5, []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}}, // 5
		{1, []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}}, // 1
		{4, []byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44}}, // 4
		{7, []byte{0x77, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77}}, // 7
	}
	for i := range del2 {
		r := recs.get(del2[i].id)
		err := pg.delRecord(r)
		if err != nil {
			t.Errorf("something went wrong (del2): %s", err)
		}
		recs.del(del2[i].id)
	}
	// should contain: X,3,X,6,8,X,X,9
	fmt.Printf("should contain: X,3,X,6,8,X,X,9\nrecords: %s\nindex: %s\n\n", getSlotAndRecsFromPage(pg), recs)

	fmt.Println("inserting...")
	ins3 := []testRecord{
		{5, []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}},  // 5
		{7, []byte{0x77, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77, 0x77}},  // 7
		{2, []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},  // 2
		{1, []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},  // 1
		{4, []byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44}},  // 4
		{10, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}, // 10
		{11, []byte{0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc}}, // 11
		{12, []byte{0xbb, 0xbb, 0xbb, 0xbb, 0xbb, 0xbb, 0xbb, 0xbb}}, // 12
	}
	for i := range ins3 {
		r, err := pg.AddRecord(ins3[i].data)
		if err != nil {
			t.Errorf("something went wrong (ins3): %s", err)
		}
		recs.add(ins3[i].id, r)
	}
	// should contain: 5,3,7,6,8,2,1,9,4
	fmt.Printf("should contain: 5,3,7,6,8,2,1,9,4\nrecords: %s\nindex: %s\n\n", getSlotAndRecsFromPage(pg), recs)

	pg.clear()
	_ = pg
	runtime.GC()
}

func TestPage_SortAndSet(t *testing.T) {
	pg := NewPage(1)
	var err error
	_, err = pg.AddRecord([]byte("CCCCCCCC"))
	if err != nil {
		t.Errorf("adding record 1: %s", err)
	}
	_, err = pg.AddRecord([]byte("AAAAAAAA"))
	if err != nil {
		t.Errorf("adding record 2: %s", err)
	}
	_, err = pg.AddRecord([]byte("DDDDDDDD"))
	if err != nil {
		t.Errorf("adding record 3: %s", err)
	}
	_, err = pg.AddRecord([]byte("BBBBBBBB"))
	if err != nil {
		t.Errorf("adding record 4: %s", err)
	}
	cells := pg.getCellPtrs()
	for i := range cells {
		fmt.Printf("pos=%d, cell=%s, rec=%s\n", i, cells[i], pg.getRecForCell(cells[i]))
	}
}

const ()
