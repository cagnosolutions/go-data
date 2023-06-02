package tcp

import (
	"fmt"
)

func Error(w ResponseWriter, err error) {
	fmt.Fprintf(w, "got an error: %s\n", err)
	return
}
