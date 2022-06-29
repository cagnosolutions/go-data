package file

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

var data = util.LongLinesTextBytes

func getReader() *DelimReader {
	return NewDelimReader(bytes.NewReader(data), '\n')
}

func TestDelimReader_Bufio_IndexData(t *testing.T) {
	r := getReader()
	spans, err := r.IndexData()
	if err != nil {
		t.Fatal(err)
	}
	for _, sp := range spans {
		fmt.Printf("id=%d, beg=%d, end=%d, data=%q\n", sp.ID, sp.Beg, sp.End, data[sp.Beg:sp.End])
	}
}

func TestDelimReader_Bufio_IndexData2(t *testing.T) {
	r := getReader()
	spans, err := r.IndexData2()
	if err != nil {
		fmt.Println("got an error and it was", err)
		t.Fatal(err)
	}
	for _, sp := range spans {
		fmt.Printf("id=%d, beg=%d, end=%d, data=%q\n", sp.ID, sp.Beg, sp.End, data[sp.Beg:sp.End])
	}
	fmt.Println(spans)
}

func BenchmarkDelimReader_Bufio_IndexData(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := getReader()
		spans, err := r.IndexData()
		if err != nil {
			b.Fatal(err)
		}
		if len(spans) < 1 {
			b.Fatal("bad read maaaan...")
		}
	}
}
