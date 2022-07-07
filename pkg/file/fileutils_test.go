package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cagnosolutions/go-data/pkg/util"
)

var data = util.LongLinesTextBytes

func getReader() *DelimReader {
	return NewDelimReader(bytes.NewReader(data), '\n')
}

func TestFileUtils_LineScanner(t *testing.T) {
	r := getReader()
	spans, err := r.LineScanner()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range spans {
		fmt.Printf("span={%s}\t\t%q\n", s, data[s.Beg:s.End])
	}
}

func TestFileUtils_LineReader(t *testing.T) {
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

func TestFileUtils_IndexSpans(t *testing.T) {
	r := getReader()
	offsets := `// 0-9
// 11-2498
// 2500-8643
// 8645-23217
// 23219-23222
// 23224-23225`
	fmt.Println(offsets)
	fmt.Println(util.Sizeof(r))
	spans, err := IndexSpans(r.r, '\n', 64)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3)
	fmt.Println(util.Sizeof(r))
	time.Sleep(3)

	for _, s := range spans {
		fmt.Printf("span={%s}\t\t%q\n", s, data[s.Beg:s.End])
	}
}

var data1 = `An Irishman's Jocular Tale

An Englishman, a Scotsman, and an Irishman all entered a
26-mile-long swimming race.  After 12 miles the Scottish 
man gets tired and drops out. Then after 16 miles the 
English man gets tired and drops out. After 25 miles the
Irish man decides he can't finish the race, so he turns
around and swims back to the start.

THE END
`

var data2 = `kinda weird;but;we;are;going;to;index;;each section using;a semi-colon;and it will do just fine;END;`

var data3 = "zero\r\none\r\ntwo\r\nthree\r\nfour\r\n\r\nfive\r\nsix\r\nseven\r\neight\r\nnine\r\nten\r\n"

func sampleTest() {

	// example 1
	r := strings.NewReader(data1)

	sp, err := IndexSpans(r, '\n', 16)
	if err != nil {
		panic(err)
	}

	for _, s := range sp {
		fmt.Printf("span={%s}, data=%q\n", s, data1[s.Beg:s.End])
	}

	fmt.Println()

	// example 2
	r = strings.NewReader(data2)
	sp, err = IndexSpans(r, ';', 16)
	if err != nil {
		panic(err)
	}
	for _, s := range sp {
		fmt.Printf("span={%s}, data=%q\n", s, data2[s.Beg:s.End])
	}

	// example 3
	r = strings.NewReader(data3)
	sp, err = IndexSpans(r, '\n', 16)
	if err != nil {
		panic(err)
	}
	for _, s := range sp {
		fmt.Printf("span={%s}, data=%q\n", s, data3[s.Beg:s.End])
	}
}

func TestFileUtils_IndexSpans2(t *testing.T) {
	sampleTest()
}

func readSlice(size int, delim byte) ([][]byte, error) {
	r := getReader()
	br := bufio.NewReaderSize(r.r, size)
	var res [][]byte
	for {
		data, err := br.ReadSlice(delim)
		if err != nil {
			if err == io.EOF {
				// fmt.Println("got EOF, breaking")
				break
			}
			if err == bufio.ErrBufferFull {
				// fmt.Println("got ErrBufferFull, continuing...")
				continue
			}
			return nil, err
		}
		res = append(res, data)
		// fmt.Printf("data=%q\n", data)
	}
	return res, nil
}

func TestFileUtils_ReadSlice(t *testing.T) {
	res, err := readSlice(64, '\n')
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range res {
		fmt.Printf("%q\n", e)
	}
}

func BenchmarkFileUtils_LineScanner(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := getReader()
		spans, err := r.LineScanner()
		if err != nil {
			b.Fatal(err)
		}
		if spans == nil || len(spans) < 1 {
			b.Fatal("bad return value")
		}
	}
}

func BenchmarkFileUtils_ReadSlice(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := readSlice(64, '\n')
		if err != nil {
			b.Fatal(err)
		}
		if res == nil || len(res) < 1 {
			b.Fatal("bad return value")
		}
	}
}

func BenchmarkFileUtils_IndexSpans(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := getReader()
		spans, err := IndexSpans(r.r, '\n', 64)
		if err != nil {
			b.Fatal(err)
		}
		if spans == nil || len(spans) < 1 {
			b.Fatal("bad return value")
		}
	}
}

func BenchmarkFileUtils_FieldsFunc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dat, err := os.ReadFile("../util/long-lines.txt")
		if err != nil {
			b.Fatal(err)
		}
		res := bytes.FieldsFunc(
			dat, func(r rune) bool {
				return r == '\n'
			},
		)
		if res == nil || len(res) < 1 {
			b.Fatal("bad return value")
		}
	}
}
