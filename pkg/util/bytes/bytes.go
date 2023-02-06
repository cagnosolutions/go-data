package bytes

import (
	"fmt"
)

// Sources for code found below will be listed below.
// [https://cs.opensource.google/go/go/+/master:src/internal/bytealg/bytealg.go]

func PrintSlice[T comparable](prefix string, x []T) {
	fmt.Printf("%-8s: %v (len=%d, cap=%d)\n", prefix, x, len(x), cap(x))
}

// SlideN slides the beginning of the slice down by the amount provided in the n argument.
// It updates the slice every time, decreasing the length and capacity with every slice.
func Slide[T comparable](x []T, n int) []T {
	if n < 0 || n > len(x) {
		panic(fmt.Sprintf("impossible slide: len:%v n:%v", len(x), n))
	}
	x = x[n:]
	return x
}

// does not modify the origional
func Filter[T comparable](x []T, fn func(T) bool) []T {
	y := x[:0]
	for _, n := range x {
		if fn(n) {
			y = append(y, n)
		}
	}
	return y
}

// modifies the origional slice
func FilterInPlace[T comparable](x []T, fn func(T) bool) []T {
	i := 0
	for _, n := range x {
		if fn(n) {
			x[i] = n
			i++
		}
	}
	x = x[:i]
	return x
}

func Cut[T comparable](x []T, i, j int) {
	var zt T
	copy(x[i:], x[j:])
	for k, n := len(x)-j+i, len(x); k < n; k++ {
		x[k] = zt // or the zero value of T
	}
	x = x[:len(x)-j+i]
}

func Delete[T comparable](x []T, i int) {
	var zt T
	copy(x[i:], x[i+1:])
	x[len(x)-1] = zt // or the zero value of T
	x = x[:len(x)-1]
}

// insert, vector (no control over len/cap)
// a = append(a[:i], append(b, a[i:]...)...)

// insert, vector and expand to pos j (control over len/cap)
// a = append(a[:i], append(make([]T, j), a[i:]...)...)

// insert, single value (non-vector) (no control over len/cap)
// a = append(a[:i], append([]T{x}, a[i:]...)...)
