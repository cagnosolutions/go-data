package main

import (
	"errors"

	"github.com/cagnosolutions/go-data/pkg/util"
)

var tracer = util.NewTracer(3)

func main() {
	err := testingA()
	if err != nil {
		tracer.Log(nil)
	}
}

func testingA() error {
	err := testingB()
	if err != nil {
		tracer.Panic(err)
	}
	return nil
}

func testingB() error {
	return errors.New("something went wrong")
}
