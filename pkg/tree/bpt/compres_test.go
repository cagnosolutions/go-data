package bpt

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"testing"
)

var strWords = []string{
	"art",
	"artwork",
	"banana",
	"cat",
	"catalog",
	"foreign",
	"forge",
	"google",
	"xyz",
	"zen",
	"zenith",
	"hello",
	"supercalifragilisticexpialidocious",
	"abcdefg",
	"This is a longer sentence with words!",
	"jdoe@example.com",
	"jondoe@example.com",
}

var bytWords = [][]byte{
	[]byte("art"),
	[]byte("artwork"),
	[]byte("banana"),
	[]byte("cat"),
	[]byte("catalog"),
	[]byte("foreign"),
	[]byte("forge"),
	[]byte("google"),
	[]byte("xyz"),
	[]byte("zen"),
	[]byte("zenith"),
	[]byte("hello"),
	[]byte("supercalifragilisticexpialidocious"),
	[]byte("abcdefg"),
	[]byte("This is a longer sentence with words!"),
	[]byte("jdoe@example.com"),
	[]byte("jondoe@example.com"),
}

var intWords = []int{
	1,
	2,
	3,
	4,
	5,
	6,
	7,
	8,
	9,
	10,
	11,
	12,
	23429,
	782,
	93432,
	182,
	993,
}

func BenchmarkBPTree_Compress(b *testing.B) {
	b.Run("CompressStr", BenchmarkBPTree_CompressStr)
	b.Run("CompressByt", BenchmarkBPTree_CompressByt)
	b.Run("CompressBytOpt", BenchmarkBPTree_CompressBytOpt)
	b.Run("CompressBytOpt2", BenchmarkBPTree_CompressBytOpt2)
	b.Run("CompressHyb", BenchmarkBPTree_CompressHyb)
	b.Run("CompressInt", BenchmarkBPTree_CompressInt)
}

func BenchmarkBPTree_CompressStr(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, word := range strWords {
			compressString(word, 8)
		}
	}
}

func BenchmarkBPTree_CompressByt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, word := range bytWords {
			compress(word, 8)
		}
	}
}

func BenchmarkBPTree_CompressBytOpt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 8)
		for _, word := range bytWords {
			compressOpt(word, &buf, 8)
		}
	}
}

func BenchmarkBPTree_CompressBytOpt2(b *testing.B) {
	clearbytes := func(buf *[8]byte) {
		for i := range buf {
			buf[i] = 0
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf [8]byte
		for _, word := range bytWords {
			compressOpt2(word, &buf)
			clearbytes(&buf)
		}
	}
}

func BenchmarkBPTree_CompressHyb(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, word := range strWords {
			compress([]byte(word), 8)
		}
	}
}

func BenchmarkBPTree_CompressInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 8)
		for _, word := range intWords {
			binary.LittleEndian.PutUint64(buf, uint64(word))
		}
	}
}

func TestBPTree_Compress(t *testing.T) {

	words := []string{
		"art",
		"artwork",
		"banana",
		"cat",
		"catalog",
		"foreign",
		"forge",
		"google",
		"xyz",
		"zen",
		"zenith",
		"hello",
		"supercalifragilisticexpialidocious",
		"abcdefg",
		"This is a longer sentence with words!",
		"jdoe@example.com",
		"jondoe@example.com",
	}

	var compressed []string

	n := 8 // Target compressed length

	// Sort words in dictionary order before insertion
	sort.Strings(words)

	// add compressed words
	fmt.Println("Compressed Words (Enhanced Uniqueness):")
	for _, word := range words {
		compressed = append(compressed, string(compress([]byte(word), n)))
	}
	// add a simple error check
	if len(compressed) != len(words) {
		t.Errorf("got: %v, expected: %v\n", len(compressed), len(words))
	}

	// sort compressed strings
	sort.Strings(compressed)

	// print out and compare
	cutp := (n / 2) - 1
	for i := 0; i < len(words); i++ {

		if !strings.HasPrefix(compressed[i], words[i][:cutp]) {
			t.Errorf("got: %v, expected: %v\n", compressed[i], words[i][:cutp])
		}
		fmt.Printf("Original: %-25s Compressed: %s\n", words[i], compressed[i])
	}

}

func TestBPTree_CompressOpt2(t *testing.T) {

	words := []string{
		"art",
		"artwork",
		"banana",
		"cat",
		"catalog",
		"foreign",
		"forge",
		"google",
		"xyz",
		"zen",
		"zenith",
		"hello",
		"supercalifragilisticexpialidocious",
		"abcdefg",
		"This is a longer sentence with words!",
		"jdoe@example.com",
		"jondoe@example.com",
	}

	var compressed []string

	var buf [8]byte

	// Sort words in dictionary order before insertion
	sort.Strings(words)

	// add compressed words
	fmt.Println("Compressed Words (Enhanced Uniqueness):")
	for _, word := range words {
		compressOpt2([]byte(word), &buf)
		compressed = append(compressed, string(buf[:]))
		for i := range buf {
			buf[i] = 0
		}
	}
	// add a simple error check
	if len(compressed) != len(words) {
		t.Errorf("got: %v, expected: %v\n", len(compressed), len(words))
	}

	// sort compressed strings
	sort.Strings(compressed)

	// print out and compare
	for i := 0; i < len(words); i++ {

		if !strings.HasPrefix(compressed[i], words[i][:3]) {
			t.Errorf("got: %v, expected: %v\n", compressed[i], words[i][:3])
		}
		fmt.Printf("Original: %-25s Compressed: %s\n", words[i], compressed[i])
	}

}
