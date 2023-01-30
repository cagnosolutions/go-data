package main

import (
	"fmt"
	"runtime"

	"github.com/cagnosolutions/go-data/pkg/engine"
	"github.com/cagnosolutions/go-data/pkg/web/utils"
)

func main() {
	// Create some pages...
	var pages []engine.Page
	var size int
	pages, size = createPages(2)
	fmt.Printf("Created %d pages totaling %dKB\n", len(pages), size)

	// Add them to a map, then watch the gc stats
	pool := make(map[uint32]engine.Page, len(pages))
	for i := range pages {
		pool[uint32(i)] = pages[i]
	}

	utils.HandleSignalInterrupt("")
	fmt.Scanln()
	runtime.GC()
}

var createPages = func(numPages int) (pages []engine.Page, totalSize int) {
	for i := 0; i < numPages; i++ {
		pages = append(pages, engine.NewPage(uint32(i), engine.P_USED))
		totalSize += 16
	}
	return pages, totalSize
}
