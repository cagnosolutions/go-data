package files

import (
	"strconv"
)

// Two is a struct
type Two struct {
	Num int
}

// NewTwo is a constructor for Two
func NewTwo(n int) *Two {
	return &Two{
		Num: n,
	}
}

// String is a stringer method for Two
func (t *Two) String() string {
	return strconv.Itoa(t.Num)
}
