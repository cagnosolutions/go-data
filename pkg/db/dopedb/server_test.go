package dopedb

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestTCPServer(t *testing.T) {

	iohandler := func(in *bytes.Buffer, out io.Writer) error {
		fmt.Printf("received: %q\nsent: OK\n", in.String())
		_, err := fmt.Fprintf(out, "OK\r\n")
		if err != nil {
			return err
		}
		return nil
	}

	err := ListenAndServe(":9999", iohandler)
	if err != nil {
		t.Fail()
	}
}
