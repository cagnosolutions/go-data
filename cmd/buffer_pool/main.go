package main

import (
	"fmt"

	"github.com/cagnosolutions/go-data/pkg/pager"
)

func main() {
	bp := pager.NewBufferPool(10, nil)
	_ = bp
	fmt.Println("Checking page sizes:")
	checkPageSizes()
	fmt.Println("Checking buffer sizes:")
	checkBufferSizes()
}

type pbsize struct {
	pageSize   int
	bufferSize int
}

func checkPageSizes() {
	sizes := []pbsize{
		{pageSize: 0},
		{pageSize: -1},
		{pageSize: 1},
		{pageSize: 52},
		{pageSize: 64},
		{pageSize: 1024},
		{pageSize: 2000},
		{pageSize: 2050},
		{pageSize: 6500},
		{pageSize: 9000},
		{pageSize: 890329},
	}
	for _, s := range sizes {
		after := pager.CalcPageSize(s.pageSize)
		fmt.Printf("pagesize.before=%d\tpagesize.after=%d\t", s.pageSize, after)
		fmt.Printf("buffsize.min=%d\tbuffsize.max=%d\n", after*8, after*32)
	}
}

func checkBufferSizes() {
	kb16 := 16 << 10
	kb64 := 64 << 10
	kb256 := 256 << 10
	mb1 := 1 << 20
	mb16 := 16 << 20
	sizes := []pbsize{
		{pageSize: 0, bufferSize: kb16},
		{pageSize: -1, bufferSize: -1},
		{pageSize: 1, bufferSize: kb64},
		{pageSize: 52, bufferSize: mb1},
		{pageSize: 64, bufferSize: 255},
		{pageSize: 1024, bufferSize: mb16},
		{pageSize: 2000, bufferSize: kb16},
		{pageSize: 2050, bufferSize: kb64},
		{pageSize: 6500, bufferSize: kb16},
		{pageSize: 9000, bufferSize: kb256},
		{pageSize: 890329, bufferSize: mb1},
	}
	for _, s := range sizes {
		after := pager.CalcBufferSize(s.pageSize, s.bufferSize)
		pgsiz := pager.CalcPageSize(s.pageSize)
		fmt.Printf(
			"buffsize.before=%d\tbuffsize.after=%d\tpagesize.after=%d (%d pages in buffer)\n",
			s.bufferSize, after, pgsiz, after/pgsiz,
		)
	}
}
