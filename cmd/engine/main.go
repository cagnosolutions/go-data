package main

import (
	"fmt"
	"strconv"

	"github.com/cagnosolutions/go-data/pkg/engine"
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
	pool := make(map[uint32]engine.Page, len(pages))
	for i := range pages {
		pool[uint32(i)] = pages[i]
	}

	createRecord := func(i int) engine.Record {
		return engine.NewRecord(
			0x12,
			[]byte(strconv.Itoa(i)),
			[]byte(fmt.Sprintf("this is record %d", i)),
		)
	}

	// Write data to a page
	for pid, page := range pool {
		if pid < 0 || page == nil {
			panic("should not happen")
		}
		for i := 0; i < 64; i++ {
			engine.AddRecord(&page, createRecord(i))
		}
	}
}

var createPages = func(numPages int) (pages []engine.Page, totalSize int) {
	for i := 0; i < numPages; i++ {
		pages = append(pages, engine.NewPage(uint32(i), engine.P_USED))
		totalSize += 16
	}
	return pages, totalSize
}
