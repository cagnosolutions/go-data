package files

import (
	"strconv"
)

// Three is a struct
type Three struct {
	Num int
}

// NewThree is a constructor for Three
func NewThree(n int) *Three {
	return &Three{
		Num: n,
	}
}

// String is a stringer method for Two
func (t *Three) String() string {
	return strconv.Itoa(t.Num)
}
