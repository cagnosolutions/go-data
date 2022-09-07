package bytes

import (
	"log"
	"testing"
)

var (
	intSlice = []int{2, 10, 1, 14, 4, 3, 6, 11, 8, 13, 9, 5, 12, 15, 0, 16}
	strSlice = []string{"C", "C", "D", "D", "B", "B", "A", "A", "F", "F", "E", "E"}
)

func Move[T comparable](x []T, i, j int) []T {
	copy(x[i:], x[j:])
	log.Println(">>>", x)
	var zt T
	for k, n := len(x)-j+i, len(x); k < n; k++ {
		x[k] = zt // or the zero value of T
	}
	x = x[:len(x)-j+i]
	return x
}

func TestMove(t *testing.T) {
	// pos := sort.SearchInts(intSlice, 7)
	PrintSlice("before", strSlice)
	x := Move(strSlice, 4, 7)
	PrintSlice("after", strSlice)
	PrintSlice("after", x)
}

func TestFilterInPlace(t *testing.T) {
	PrintSlice("before", intSlice)

	FilterInPlace(
		intSlice, func(n int) bool {
			// if the value n, is an even number, keep it
			return n%2 == 0
		},
	)
	PrintSlice("after", intSlice)
}

func TestFilter(t *testing.T) {
	PrintSlice("before", intSlice)
	Filter(
		intSlice, func(n int) bool {
			// if the value n, is an even number, keep it
			return n%2 == 0
		},
	)
	PrintSlice("after", intSlice)
}

func TestCut(t *testing.T) {
	PrintSlice("before", intSlice)
	Cut(intSlice, 0, 5)
	PrintSlice("after", intSlice)
}

func TestDelete(t *testing.T) {
	PrintSlice("before", intSlice)
	Delete(intSlice, 5)
	PrintSlice("after", intSlice)
}
