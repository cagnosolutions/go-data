package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

var data = util.LongLinesTextBytes

func getReader() *DelimReader {
	return NewDelimReader(bytes.NewReader(data), '\n')
}

func TestDelimReader_LineScanner(t *testing.T) {
	r := getReader()
	spans, err := r.LineScanner()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range spans {
		fmt.Printf("span={%s}\t\t%q\n", s, data[s.Beg:s.End])
	}
}

func TestDelimReader_LineReader(t *testing.T) {
	r := getReader()
	offsets := `// 0-9
// 11-2498
// 2500-8643
// 8645-23217
// 23219-23222
// 23224-23225`
	fmt.Println(offsets)
	spans, err := r.LineReader()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range spans {
		fmt.Printf("span={%s}\t\t%q\n", s, data[s.Beg:s.End])
	}
}

func Test_IndexSpans(t *testing.T) {
	r := getReader()
	offsets := `// 0-9
// 11-2498
// 2500-8643
// 8645-23217
// 23219-23222
// 23224-23225`
	fmt.Println(offsets)
	spans, err := IndexSpans(r.r, '\n', 64)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range spans {
		fmt.Printf("span={%s}\t\t%q\n", s, data[s.Beg:s.End])
	}
}

func TestBytesReadSlice(t *testing.T) {
	r := getReader()
	br := bufio.NewReaderSize(r.r, 64)
	for {
		data, err := br.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("got EOF, breaking")
				break
			}
			if err == bufio.ErrBufferFull {
				// fmt.Println("got ErrBufferFull, continuing...")
				continue
			}
			t.Error(err)
		}
		fmt.Printf("data=%q\n", data)
	}
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

func TestDelimReader_Bufio_ReadLines(t *testing.T) {
	r := getReader()
	br := bufio.NewReader(r.r)
	var line []byte
	for line == nil {
		line, err := readLine(br)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("got an error and it was", err)
			t.Fatal(err)
		}
		fmt.Printf("%s\n", line)
	}
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
