package main

import (
	"log"
	"os"

	"github.com/cagnosolutions/go-data/cmd/gen/parser"
)

func main() {
	p, err := parser.NewParser(os.Getenv("GOFILE"))
	if err != nil {
		fail(err)
	}
	err = p.Parse()
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	log.Println(err)
	os.Exit(1)
}
