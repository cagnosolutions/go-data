package bytes

import (
	"fmt"
)

// SlideN slides the beginning of the slice down by the amount provided in the n argument.
// It updates the slice every time, decreasing the length and capacity with every slice.
func SlideN(bz *[]byte, count *int, n int) bool {
	if n < 0 || n > len(*bz) {
		panic(fmt.Sprintf("impossible slide: len:%v n:%v", len(*bz), n))
	}
	*bz = (*bz)[n:]
	if count != nil {
		*count += n
	}
	return true
}
