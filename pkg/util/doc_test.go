package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

var d1 = []byte("This is a line.\nThis is another line.\nThis is a third line.\nAnd a fourth.\n")
var d2 = []byte(`{'doc':1,'id':23,'name':'jon doe'}
{'doc':2,'id':13,'name':'felix doe'}
{'doc":3,'id':8,'name':'mary doe'}
{'doc":4,"id':20,'name':'ron doe'}
{'doc':5,'id':94,'name':'jack doe'}
`)

func TestSpansDelim(t *testing.T) {
	spans := spansDelim(d1, '\n')
	if len(spans) != 4 {
		t.Errorf("expected len(spans) == 4, got len(spans) == %d\n", len(spans))
	}
	for i, span := range spans {
		fmt.Printf(
			"Span %d begins at char %d, and ends at char %d.\nData: %q\n\n",
			i+1,
			span.beg,
			span.end,
			d1[span.beg:span.end],
		)
		if span.end-span.beg <= 0 {
			t.Errorf("bad span offset")
		}
	}
}

func TestSpansFunc(t *testing.T) {
	spans := spansFunc(
		d2, func(b byte) bool {
			switch b {
			case ',', '{', '}', '\n':
				return true
			}
			return false
		},
	)
	if len(spans) != 19 {
		t.Errorf("expected len(spans) == 19, got len(spans) == %d\n", len(spans))
	}
	for i, span := range spans {
		fmt.Printf(
			"Span %d begins at char %d, and ends at char %d.\nData: %q\n\n",
			i+1,
			span.beg,
			span.end,
			d2[span.beg:span.end],
		)
		if span.end-span.beg <= 0 {
			t.Errorf("bad span offset")
		}
	}
}

func TestNewDocument(t *testing.T) {
	tf := makeTestFile()
	writeData(tf)
	defer closeTestFile(tf)
	doc, err := NewDocument(tf)
	if err != nil {
		t.Errorf("error initializing new document: %s\n", err)
	}
	for _, span := range doc.spans {
		fmt.Printf("%+v\n", span)
	}
}

func BenchmarkIndexAll(b *testing.B) {
	var spans []span
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		spans = IndexAll(d2, []byte(`id`), 8)
		if spans == nil {
			b.Errorf("error, no spans found\n")
		}
	}
	b.StopTimer()
	for i, sp := range spans {
		fmt.Printf(
			"span %d, is at %v and contains %q\n",
			i, sp, d2[sp.beg:sp.end],
		)
	}
}

func BenchmarkBytesCut(b *testing.B) {
	type span_ struct {
		before []byte
		after  []byte
		found  bool
	}
	spans := make([]span_, 0, 32)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bef, aft, ok := bytes.Cut(d2, []byte(`id`))
		spans = append(spans, span_{before: bef, after: aft, found: ok})
		if spans == nil {
			b.Errorf("error, no spans found\n")
		}
	}
	b.StopTimer()
	// for _, sp := range spans {
	// 	fmt.Printf(
	// 		"span %v\n", sp,
	// 	)
	// }
}

func makeTestFile() *os.File {
	tf, err := os.CreateTemp("", "data.txt")
	if err != nil {
		log.Panicf("error creating temp file: %s\n", err)
	}
	return tf
}

func writeData(fp *os.File) {
	_, err := fp.Write(d1)
	if err != nil {
		log.Panicf("error writing to temp file: %s\n", err)
	}
	_, err = fp.Write(d2)
	if err != nil {
		log.Panicf("error writing to temp file: %s\n", err)
	}
	_, err = fp.Seek(0, io.SeekStart)
	if err != nil {
		log.Panicf("error seeking in temp file: %s\n", err)
	}
}

func closeTestFile(fp *os.File) {
	err := fp.Close()
	if err != nil {
		log.Panicf("error closing temp file: %s\n", err)
	}
}
