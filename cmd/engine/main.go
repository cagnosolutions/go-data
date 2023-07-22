package main

import (
	"fmt"
	"strconv"

	"github.com/cagnosolutions/go-data/pkg/engine/page"
)

func main() {
	testPageV1()
	// go func() {
	// 	for {
	// 		time.Sleep(1 * time.Second)
	// 		runtime.GC()
	// 	}
	// }()
	//
	// time.Sleep(30 * time.Second)
}

func testPageV1() {
	// Create some pages...
	pages, size := createPages(16)
	fmt.Printf("Created %d pages totaling %dKB\n", len(pages), size)

	// Add them to a map, then watch the gc stats
	pool := make(map[uint32]page.Page, len(pages))
	for i := range pages {
		pool[uint32(i)] = pages[i]
	}

	createRecord := func(i int) page.Record {
		return page.NewRecord(
			page.R_NUM,
			page.R_STR,
			[]byte(strconv.Itoa(i)),
			[]byte(fmt.Sprintf("this is record %d", i)),
		)
	}

	// Write data to a page
	for pid, pg := range pool {
		if pid < 0 || pg == nil {
			panic("should not happen")
		}
		for i := 0; i < 64; i++ {
			pg.AddRecord(createRecord(i))
		}
	}
}

var createPages = func(numPages int) (pages []page.Page, totalSize int) {
	for i := 0; i < numPages; i++ {
		pages = append(pages, page.NewPage(uint32(i), page.P_USED))
		totalSize += 16
	}
	return pages, totalSize
}
