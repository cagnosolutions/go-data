package main

import (
	"github.com/cagnosolutions/go-data/cmd/io/examples/reading"
	"github.com/cagnosolutions/go-data/cmd/io/examples/util"
)

const bufSize = 256

func main() {

	// run BasicRead
	reading.BasicRead(data, util.NewOutWriter(), bufSize)

	// run buffered Reader
	reading.BufferedRead(data, util.NewOutWriter(), bufSize)
}
