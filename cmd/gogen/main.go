package main

import (
	"github.com/cagnosolutions/go-data/pkg/gogen/parser"
)

func main() {
	p, err := parser.NewParser("_hello.go")
	if err != nil {
		panic(err)
	}
	err = p.ParseStruct()
	if err != nil {
		panic(err)
	}

	p.Find(p.File())
}
