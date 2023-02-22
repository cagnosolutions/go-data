package slicer

import (
	"fmt"
	"sort"
	"testing"
)

var res interface{}

func BenchmarkInsertVectorInts(b *testing.B) {
	b.ReportAllocs()
	nn := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := 0; i < b.N; i++ {
		_ = insert(nn, 4, 33, 33, 33)
	}
	res = nn
}

func TestCutInts(t *testing.T) {
	nn := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Println("before cut:", nn)
	fmt.Println("cutting 3-7")
	nn = Cut(nn, 3, 7)
	fmt.Println("after cut:", nn)
}

func TestCutPtrInts(t *testing.T) {
	nn := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Println("before cut:", nn)
	fmt.Println("cutting 3-7")
	CutPtr(&nn, 3, 7)
	fmt.Println("after cut:", nn)
}

func TestDelInts(t *testing.T) {
	nn := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Println("before delete:", nn)
	fmt.Println("deleting 4")
	nn = Del(nn, 4)
	fmt.Println("after delete:", nn)
}

func TestDelPtrInts(t *testing.T) {
	nn := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Println("before delete:", nn)
	fmt.Println("deleting 4")
	DelPtr(&nn, 4)
	fmt.Println("after delete:", nn)
}

func TestCutStrings(t *testing.T) {
	nn := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	fmt.Println("before cut:", nn)
	fmt.Println("cutting 3-7")
	nn = Cut(nn, 3, 7)
	fmt.Println("after cut:", nn)
}

func TestCutPtrStrings(t *testing.T) {
	nn := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	fmt.Println("before cut:", nn)
	fmt.Println("cutting 3-7")
	CutPtr(&nn, 3, 7)
	fmt.Println("after cut:", nn)
}

func TestCutBytes(t *testing.T) {
	nn := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	fmt.Println("before cut:", nn)
	fmt.Println("cutting 3-7")
	nn = Cut(nn, 3, 7)
	fmt.Println("after cut:", nn)
}

func TestCutPtrBytes(t *testing.T) {
	nn := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	fmt.Println("before cut:", nn)
	fmt.Println("cutting 3-7")
	CutPtr(&nn, 3, 7)
	fmt.Println("after cut:", nn)
}

type A struct {
	Num int
}

func TestCutStructs(t *testing.T) {
	nn := []A{{Num: 0}, {Num: 1}, {Num: 2}, {Num: 3}, {Num: 4}, {Num: 5}, {Num: 6}, {Num: 7}, {Num: 8}, {Num: 9}}
	fmt.Println("before cut:", nn)
	fmt.Println("cutting 3-7")
	nn = Cut(nn, 3, 7)
	fmt.Println("after cut:", nn)
}

func TestCutPtrStructs(t *testing.T) {
	nn := []A{{Num: 0}, {Num: 1}, {Num: 2}, {Num: 3}, {Num: 4}, {Num: 5}, {Num: 6}, {Num: 7}, {Num: 8}, {Num: 9}}
	fmt.Println("before cut:", nn)
	fmt.Println("cutting 3-7")
	CutPtr(&nn, 3, 7)
	fmt.Println("after cut:", nn)
}

func TestFilter(t *testing.T) {
	nn := []int{
		33, 33, 33, 0, 0, 0, 0, 0, 0, 0,
		1, 1, // 10-12
		0, 0, 0, // 2-5
		2, 2, 2, // 5-8
		3, 3, // 8-10
		0, 0, // 10-12
		4, 4, 4, // 12-15
		5, 5, // 15-17
		0,       // 17-18
		6,       // 18-19
		0, 0, 0, // 19-22
		7, 7, 7, // 22-25
		8, 8, 8, 8, // 25-29
		0, 0, 0, 0, // 29-33
		9, 9, // 33-35
	}
	fmt.Println(nn)
	nn = Filter(
		nn, func(i int) bool {
			return i > 0
		},
	)
	fmt.Println(nn)
}

func TestMoveV2(t *testing.T) {
	nn := []int{
		-1, -1, -1, -1, -1, -1, // 0,6
		0, 0, // 6,2
		1, 1, 1, // 8,3
		2, 2, // 11,2
		0, 0, 0, 0, 0, // 13,5
		3, 3, 3, // 18,3
		4, 4, 4, 4, // 21,4
		0, 0, // 25,2
		5, 5, 5, // 27,3
		6, 6, // 30,2
		0, 0, 0, // 32,3
		7, 7, 7, // 35,3
		8, 8, 8, 8, // 38,4
		9, 9, // 42,2
	}
	// index
	index := map[int]slot{
		-1: {0, 6, false}, // free space
		1:  {6, 2, false},
		2:  {8, 3, true},
		3:  {11, 2, true},
		4:  {13, 5, false},
		5:  {18, 3, true},
		6:  {21, 4, true},
		7:  {25, 2, false},
		8:  {27, 3, true},
		9:  {30, 2, true},
		10: {32, 3, false},
		11: {35, 3, true},
		12: {38, 4, true},
		13: {42, 2, true},
	}

	// populate
	var sls slots
	for i := 0; i < len(index); i++ {
		sl := index[i]
		sls = append(sls, &sl)
	}
	fmt.Println(sls)

	// before
	fmt.Println("BEFORE: ", nn)
	for i := range sls {
		sub := SubSlice(nn, sls[i].offset, sls[i].length)
		fmt.Println(sub)
	}

	sort.Stable(sls)

	// after
	fmt.Println(" AFTER: ", nn)
	for i := range sls {
		sub := SubSlice(nn, sls[i].offset, sls[i].length)
		fmt.Println(sub)
	}
}

func (s slots) Len() int {
	return len(s)
}

func (s slots) Less(i, j int) bool {
	if (*s[i]).used == false && (*s[j]).used == true {
		return true
	}
	return false
}

func (s slots) Swap(i, j int) {
	*s[i], *s[j] = *s[j], *s[i]
}

func (s *slot) bounds() (int, int) {
	return s.offset, s.offset + s.length
}

func compaction(nn []int, index map[int]slot) []int {
	// get free space slice
	free := SubSlice(nn, index[-1].offset, index[-1].length)
	// create a current slot variable
	var prev *slot
	// loop over index and check each slot
	for i := len(index) - 1; i > 0; i-- {
		// get slot
		sl := index[i]
		// if the slot is not used (it is free)
		if prev != nil && !prev.used {
			pv := SubSlice(nn, prev.offset, prev.length)
			cv := SubSlice(nn, sl.offset, sl.length)
			fmt.Printf("prev=%v, current=%v\n", pv, cv)
		}
		// set the prev slot for the next iteration
		prev = &sl
	}
	_ = free
	return nn
}

func TestMove(t *testing.T) {
	nn := []byte{
		33, 33, 33, 0, 0, 0, 0, 0, 0, 0, // 0-10
		1, 1, // 10-12
		0, 0, 0, // 12-15
		2, 2, 2, // 15-18
		3, 3, // 18-20
		0, 0, // 20-22
		4, 4, 4, // 22-25
		5, 5, // 25-27
		0,    // 27-28
		6,    // 28-29
		0, 0, // 29-31
		7, 7, 7, // 31-34
		8, 8, 8, 8, // 34-38
		0, 0, 0, // 38-41
		9, 9, // 41-43
	}

	ss := []*slot{
		{41, 2, true},
		{38, 3, false},
		{34, 4, true},
		{31, 3, true},
		{29, 2, false},
		{28, 1, true},
		{27, 1, false},
		{25, 2, true},
		{22, 3, true},
		{20, 2, false},
		{18, 2, true},
		{15, 3, true},
		{12, 3, false},
		{10, 2, true},
		// {0, 10, false},
	}
	lo, hi := 3, 10

	fmt.Println("BEFORE:")
	slots(ss).PrintSlots(nn)
	fmt.Println(nn[lo:hi], nn[hi:])

	slots(ss).GC(&nn)

	fmt.Println("AFTER:")
	slots(ss).PrintSlots(nn)
	fmt.Println(nn[lo:hi], nn[hi:])

}

func (sls slots) PrintSlots(data []byte) {
	fmt.Printf("%+v\n", sls)
	for i, sl := range sls {
		beg, end := sl.bounds()
		fmt.Printf("  slot[%d]=%d\n", i, data[beg:end])
	}
}

type slots []*slot

func (sls slots) GC(data *[]byte) {
	// var prev *slot
	var prevID int
	for i, sl := range sls {
		if sl.used {
			// set pointer for prev slot so we can update it later
			// prev = sls[i]
			prevID = i
			continue
		}
		beg, end := sl.bounds()
		// fmt.Printf("beg=%d, end=%d, span=%d", beg, end, (*data)[beg:end])
		copy((*data)[beg:], (*data)[end:])
		// var nb, ne int
		for k, n := len(*data)-end+beg, len(*data); k < n; k++ {
			// fmt.Printf(", k=%d, n=%d\n", k, n)
			(*data)[k] = 0 // or the zero value of T
			// nb, ne = k-1, n
		}
		fmt.Println(">>> beg", beg)
		sls[prevID].offset = beg

		// prev.offset = beg
		// fmt.Printf("updating previous slot, was=%+v, now=%+v\n", sls[prevID], prev)
		// fmt.Printf(
		//	"slot[%d]=%v\nwas=data[%d:%d] (len=%d, cap=%d), now=data[%d:%d] (len=%d, cap=%d)\n\n",
		//	i, sl, nb, ne, len(*data), cap(*data), beg, end, len(*data), cap(*data),
		// )

		// previous (used) slot offset is: newBeg, newEnd := beg, beg+(ne-nb)
		// fmt.Printf("was=%d:%d, now=%d:%d\n", nb, ne, beg, beg+(ne-nb))
		// *data = (*data)[:len(*data)-end+beg]
	}
}

func (sls slots) GCold(data *[]byte, lo, hi int) {
	// get our slice of usable free space which is
	// always going to be between our lower and our
	// upper boundary markers
	// free := (*data)[lo:hi]
	// create our upper pointer to use
	var ptr int
	// start iterating our slots
	for _, sl := range sls {
		// handle collect empty record
		if !sl.used {
			// gets the bounds of the "empty record"
			// so we can set the empty pointer
			_, ptr = sl.bounds()
			fmt.Println("got pointer of empty record", ptr)
			continue
			// check that the free space has room
			// for the empty record, else panic
			// if len(free) < end-beg {
			//	panic("no more room")
			// }
			// copy the record to the free space
			// copy(free, (*data)[beg:end])
			// fmt.Println(beg, end, (*data)[beg:end])
		}
		if sl.used {
			// check to make sure we have a
			// pointer to use for our copy
			if ptr <= 0 {
				// if we don't have a copy pointer
				// then we just continue
				continue
			}
			// we must have a copy pointer, so we
			// should get the bounds of the record
			beg, end := sl.bounds()
			// now we can copy our record
			// copy((*data)[:ptr], (*data)[:end])
			a := (*data)[ptr-(end-beg) : ptr]
			b := (*data)[beg:end]
			copy(a, b)
			// *data = (*data)[:end]
			fmt.Printf("a=%d\n", a)
			fmt.Printf("b=%d\n", b)
			ptr = end - 1
			// fmt.Printf("c=%d\n", c)
		}
	}
}

func printSlices(nn []int) {
	fmt.Println(nn[0:2])
	fmt.Println(nn[2:5])
	fmt.Println(nn[5:8])
	fmt.Println(nn[8:10])
	fmt.Println(nn[10:12])
	fmt.Println(nn[12:15])
	fmt.Println(nn[15:17])
	fmt.Println(nn[17:18])
	fmt.Println(nn[18:19])
	fmt.Println(nn[19:22])
	fmt.Println(nn[22:25])
	fmt.Println(nn[25:29])
	fmt.Println(nn[29:33])
	fmt.Println(nn[33:35])
}
