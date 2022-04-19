package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	this := filepath.Base(os.Args[0])
	file := os.Getenv("GOFILE")
	arg1 := os.Args[1]
	fmt.Printf("Running %q go on %q, (%T) %#v\n", this, file, arg1, arg1)
}
