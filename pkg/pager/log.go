package pager

import (
	"log"
	"os"
)

var debug = log.New(os.Stdout, "::[DEBUG] >> ", log.Lshortfile|log.Lmsgprefix)
