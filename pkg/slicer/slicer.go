package slicer

import (
	"fmt"
	"sort"
)

// Cut cuts a set of elements from a slice at the provided beginning and ending
// indexes. It takes a slice of some type and returns the modified slice. **This
// function is GC friendly.
func Cut[T comparable](slice []T, beg, end int) []T {
	var zero T
	copy(slice[beg:], slice[end:])
	for k, n := len(slice)-end+beg, len(slice); k < n; k++ {
		slice[k] = zero // or the zero value of T
	}
	slice = slice[:len(slice)-end+beg]
	return slice
}

// CutPtr cuts a set of elements from a slice at the provided beginning and ending
// indexes. It takes a pointer to a slice of some type. *This function is GC friendly.
func CutPtr[T comparable](slice *[]T, beg, end int) {
	var zero T
	copy((*slice)[beg:], (*slice)[end:])
	for k, n := len(*slice)-end+beg, len(*slice); k < n; k++ {
		(*slice)[k] = zero // or the zero value of T
	}
	*slice = (*slice)[:len(*slice)-end+beg]
}

// Del removes an element from a slice at the provided index. It takes a slice of
// some type and returns the modified slice. *This function preserves the ordering
// of the slice. *This function is GC friendly.
func Del[T comparable](slice []T, at int) []T {
	var zero T
	if at < len(slice)-1 {
		copy(slice[at:], slice[at+1:])
	}
	slice[len(slice)-1] = zero // or the zero value of T
	slice = slice[:len(slice)-1]
	return slice
}

// DelPtr removes an element from a slice at the provided index. It takes a pointer
// to a slice of some type. *This function preserves the ordering of the slice. *This
// function is GC friendly.
func DelPtr[T comparable](slice *[]T, at int) {
	var zero T
	if at < len(*slice)-1 {
		copy((*slice)[at:], (*slice)[at+1:])
	}
	(*slice)[len(*slice)-1] = zero // or the zero value of T
	*slice = (*slice)[:len(*slice)-1]
}

func SlidingWindow(size int, input []int) [][]int {
	// returns the input slice as the first element
	if len(input) <= size {
		return [][]int{input}
	}
	// allocate slice at the precise size we need
	r := make([][]int, 0, len(input)-size+1)
	for i, j := 0, size; j <= len(input); i, j = i+1, j+1 {
		r = append(r, input[i:j])
	}
	return r
}

func Shift(slice []int, i, j int) []int {
	return nil
}

type slot struct {
	offset int
	length int
	used   bool
}

//
// func (s slot) bounds() (int, int) {
// 	return s.offset, s.offset + s.length
// }

// func _move(slice *[]int, sl *slot) {
// 	beg, end := sl.bounds()
// 	copy((*slice)[beg:], (*slice)[end:])
// 	for k, n := len(*slice)-end+beg, len(*slice); k < n; k++ {
// 		fmt.Println((*slice)[k])
// 		(*slice)[k] = -1 // or the zero value of T
// 	}
// 	// slice = slice[:len(slice)-end+beg]
// }

func Filter(slice []int, keep func(i int) bool) []int {
	n := 0
	for _, x := range slice {
		if keep(x) {
			slice[n] = x
			n++
		}
	}
	// slice = slice[:n]
	return slice
}

func insert[T comparable](slice []T, i int, v ...T) []T {
	// if cap not full, do in place
	if n := len(slice) + len(v); n <= cap(slice) {
		s2 := slice[:n]
		copy(s2[i+len(v):], slice[i:])
		copy(s2[i:], v)
		return s2
	}
	// otherwise, we have to grow the slice
	s2 := make([]T, len(slice)+len(v))
	copy(s2, slice[:i])
	copy(s2[i:], v)
	copy(s2[i+len(v):], slice[i:])
	return s2
}

func cut[T comparable](slice []T, i, j int) []T {
	var zero T
	copy(slice[i:], slice[j:])
	beg := len(slice) - j + i
	end := len(slice)
	for k, n := beg, end; k < n; k++ {
		slice[k] = zero // or the zero value of T
	}
	// adjust slice
	// slice = slice[:beg]
	// return
	return slice
}

func SubSlice[T comparable](slice []T, offset, length int) []T {
	if offset < 0 || offset+length > len(slice) {
		panic("sliced region is out of bounds")
	}
	return slice[offset : offset+length]
}

func Move[T comparable](slice []T, fromOffset, fromLength, toOffset, toLength int) []T {
	// the "cut" operation
	var zero T
	copy(slice[fromOffset:], slice[fromOffset+fromLength:])
	beg := len(slice) - (fromOffset + fromLength) + fromOffset
	end := len(slice)
	for k, n := beg, end; k < n; k++ {
		slice[k] = zero // or the zero value of T
	}
	// slice = slice[:beg]
	cutset := SubSlice(slice, toOffset, toLength)
	fmt.Println(">>>", cutset)
	// the "paste" (otherwise known as insert) operation
	if n := len(slice) + len(cutset); n <= cap(slice) {
		sub := slice[:n]
		copy(sub[toOffset+len(cutset):], slice[toOffset:])
		copy(sub[toOffset:], cutset)
		return sub
	}
	// otherwise, we have to grow the slice
	grown := make([]T, len(slice)+len(cutset))
	copy(grown, slice[:toOffset])
	copy(grown[toOffset:], cutset)
	copy(grown[toOffset+len(cutset):], slice[toOffset:])
	return grown
}

func swap(nums []int, i, j int) {
	nums[i] ^= nums[j]
	nums[j] ^= nums[i]
	nums[i] ^= nums[j]
}

// Ordered is a type that represents all the types that can be
// compared by <,> which is above and beyond a normal equality
// comparison.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

// Slice is a generic slice type that allows you to create a generic
// slice of any type that is Ordered
type Slice[T Ordered] []T

// Len fulfills the sort.Sort interface by returning the length of
// the slice s.
func (s Slice[T]) Len() int {
	return len(s)
}

// Less fulfills the sort.Sort interface by returning true if element
// found at index i is less than the element found at index j.
func (s Slice[T]) Less(i, j int) bool {
	return s[i] < s[j]
}

// Swap fulfills the sort.Sort interface by swapping the elements found
// at i and j.
func (s Slice[T]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Sort sorts the slice using the sort package. It utilizes the methods
// Len, Less and Swap that we have implemented for the generic Slice type.
func (s Slice[T]) Sort() {
	sort.Stable(s)
}

// Search performs a binary search on the Slice s.
func (s Slice[T]) Search(x T) (int, T) {
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	i, j := 0, s.Len()
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if s[i] < x {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i, s[i]
}
