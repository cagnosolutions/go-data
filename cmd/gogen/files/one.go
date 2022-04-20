package files

import (
	"strconv"
)

// One is a struct
type One struct {
	Num int
}

// NewOne is a constructor for One
func NewOne(n int) *One {
	return &One{
		Num: n,
	}
}

// String is a stringer method for One
func (o *One) String() string {
	return strconv.Itoa(o.Num)
}
