package ngram

import (
	"fmt"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestNgramScanner(t *testing.T) {
	data := []byte(util.GettysburgAddress)
	// see how many n-grams we can find
	ngs := make(map[string]int)
	// make our ngram function
	fn := func(g []byte, beg, end int) bool {
		// if _, match := bigram[string(g)]; match {
		ngs[string(g)]++
		// }
		return true
	}
	err := NgramScanner(4, data, fn)
	if err != nil {
		t.Error(err)
	}
	// see what matches we have (occurred, occurrence)
	for gram, occurrence := range ngs {
		fmt.Printf("Ngram %q occurred %d times\n", gram, occurrence)
	}
}

func TestNNgramScanner(t *testing.T) {
	data := []byte(util.GettysburgAddress)
	// see how many n-grams we can find
	ngs := make(map[string]int)
	// make our ngram function
	fn := func(g []byte, beg, end int) bool {
		// if _, match := bigram[string(g)]; match {
		ngs[string(g)]++
		// }
		return true
	}
	err := NNgramScanner(2, 4, data, fn)
	if err != nil {
		t.Error(err)
	}
	// see what matches we have (occurred, occurrence)
	for gram, occurrence := range ngs {
		fmt.Printf("Ngram %q occurred %d times\n", gram, occurrence)
	}
}
