package log

import (
	"log"
)

func debug(msg string, args ...any) {
	format := ">> DEBUG: " + msg
	log.Printf(format, args...)
}
