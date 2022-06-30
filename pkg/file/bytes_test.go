package file

import (
	"bytes"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func indexBytePortable(s []byte, c byte) int {
	for i, b := range s {
		if b == c {
			return i
		}
	}
	return -1
}

var ch = byte('\n')
var text = util.LongLinesTextBytes

func BenchmarkBytes_IndexByte(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		n := bytes.IndexByte(data, ch)
		if n == -1 {
			b.Fatal("could not locate byte")
		}
	}
}

func BenchmarkBytes_IndexBytePortable(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		n := indexBytePortable(data, ch)
		if n == -1 {
			b.Fatal("could not locate byte")
		}
	}
}

func BenchmarkBytes_Index(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		n := bytes.Index(data, []byte{ch})
		if n == -1 {
			b.Fatal("could not locate byte")
		}
	}
}
