package reading

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/cagnosolutions/go-data/cmd/io/examples/util"
)

// BasicRead reads raw []byte data from input and writes it to output
func BasicRead(input []byte, output *util.OutWriter, buffsize int) {
	// get new bytes reader
	r := bytes.NewReader(input)
	// start reading
	for {
		// buf to read into
		buf := make([]byte, buffsize)
		_, err := r.Read(buf)
		if err != nil {
			// we have an error, but it might
			// just be the end of the file
			if err == io.EOF {
				break
			}
			panic(err) // if not EOF, panic
		}
		// let's write to our output
		_, err = output.Write(buf)
		if err != nil {
			panic(err)
		}
	}
	// when we are all finished, print the data
	fmt.Printf("BasicRead Read The Following:\n%s\n", output.Bytes())
}

// BufferedRead reads raw []byte data from input and writes it to output
func BufferedRead(input []byte, output *util.OutWriter, buffsize int) {
	// get new buffered reader
	r := bufio.NewReader(bytes.NewReader(input))
	// start reading
	for {
		// read from buffered reader, until end of line
		line, err := r.ReadBytes('\n')
		if err != nil {
			// we have an error, but it might
			// just be the end of the file
			if err == io.EOF {
				break
			}
			panic(err) // if not EOF, panic
		}
		// let's write to our output
		_, err = output.Write(line)
		if err != nil {
			panic(err)
		}
	}
	// when we are all finished, print the data
	fmt.Printf("Buffered Read The Following:\n%s\n", output.Bytes())
}
