package spanindexer

import (
	"fmt"
	"testing"
)

var testData = []byte(`THE SONNETS
ALL’S WELL THAT ENDS WELL
THE TRAGEDY OF ANTONY AND CLEOPATRA
AS YOU LIKE IT
AAAA
AAAA
AAAA
AAAA
THE COMEDY OF ERRORS
THE TRAGEDY OF CORIOLANUS
AAAA
AAAA
CYMBELINE
THE TRAGEDY OF HAMLET, PRINCE OF DENMARK
THE FIRST PART OF KING HENRY THE FOURTH
AAAA
AAAA
AAAA
THE SECOND PART OF KING HENRY THE FOURTH
THE LIFE OF KING HENRY THE FIFTH
THE FIRST PART OF HENRY THE SIXTH
THE SECOND PART OF KING HENRY THE SIXTH
THE THIRD PART OF KING HENRY THE SIXTH
KING HENRY THE EIGHTH
KING JOHN
`)

func TestIndexer(t *testing.T) {
	ix := NewIndexer('\n')
	ix.Index(testData)
	for i := 0; i < ix.Len(); i++ {
		v := ix.GetSegment(i)
		if v == nil {
			continue
		}
		j, k := v.Beg, v.End
		if v.Status == 1 {
			fmt.Printf("line=%.2d, data=%q\n", i, testData[j:k])
		}
	}
	ix.PruneSegment(15)
	ix.PruneSegment(16)
	ix.PruneSegment(17)
	ix.Range(
		func(part, status, beg, end int) error {
			if status == 0 {
				return SkipSpan
			}
			fmt.Printf("line=%.2d, data=%q\n", part, testData[beg:end])
			return nil
		},
	)
	for i := 0; i < ix.Len(); i++ {
		v := ix.GetSegment(i)
		if v == nil {
			continue
		}
		j, k := v.Beg, v.End
		fmt.Printf("line=%.2d, data=%q\n", i, testData[j:k])
	}
}
