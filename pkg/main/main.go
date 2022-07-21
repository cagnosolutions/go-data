package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

func main() {

	PrintMemUsage()

	fd, err := os.OpenFile("pkg/util/large-binary-data.txt", os.O_RDWR|os.O_SYNC, 0666)
	if err != nil {
		panic(err)
	}
	br := bufio.NewReaderSize(fd, 512<<10)
	data := make([]byte, 512<<10)
	var n int
	for {
		n, err = br.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		fmt.Printf("read %d bytes of data...\n", n)
		PrintMemUsage()
		// _, err = fmt.Scanln()
		// if err != nil {
		// 	panic(err)
		// }
		time.Sleep(3 * time.Second)
	}
	PrintMemUsage()
	err = fd.Close()
	if err != nil {
		panic(err)
	}
}

func writeLargeFile() {
	fd, err := os.OpenFile("pkg/util/large-binary-data.txt", os.O_CREATE|os.O_SYNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 4096*16; i++ {
		_, err = fd.Write(record)
		if err != nil {
			panic(err)
		}
	}
	err = fd.Close()
	if err != nil {
		panic(err)
	}
}

var record = []byte{
	'a', 'a', 'a', 'a',
	'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b',
	'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c', 'c',
	'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd',
	'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd', 'd',
	'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e',
	'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e',
	'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e',
	'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e',
	'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f',
	'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f',
	'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f',
	'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f',
	'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f',
	'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f',
	'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f',
	'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f', 'f',
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
