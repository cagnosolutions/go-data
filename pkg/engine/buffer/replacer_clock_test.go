package buffer

import (
	"fmt"
	"testing"
)

func TestClockReplacer_Size(t *testing.T) {
	cr := NewClockReplacer(10)
	fmt.Println(cr.Size())
	cr.Unpin(1)
	fmt.Println(cr.Size())

}

func TestClockReplacer_All(t *testing.T) {
	cr := NewClockReplacer(10)
	cr.Unpin(1)
	cr.Unpin(2)
	cr.Unpin(3)
	cr.Unpin(4)
	cr.Unpin(5)
	cr.Unpin(6)
	cr.Unpin(1)

	ans := cr.Size()
	if ans != 6 {
		t.Errorf("got %d, want %d", ans, 6)
	}

	val := cr.Victim()
	if *val != 1 {
		t.Errorf("got %d, want %d", val, 1)
	}

	val = cr.Victim()
	if *val != 2 {
		t.Errorf("got %d, want %d", val, 2)
	}

	val = cr.Victim()
	if *val != 3 {
		t.Errorf("got %d, want %d", val, 3)
	}

	cr.Pin(3)
	cr.Pin(4)
	ans = cr.Size()
	if ans != 2 {
		t.Errorf("got %d, want %d", ans, 2)
	}

	cr.Unpin(4)

	val = cr.Victim()
	if *val != 5 {
		t.Errorf("got %d, want %d", val, 5)
	}

	val = cr.Victim()
	if *val != 6 {
		t.Errorf("got %d, want %d", val, 6)
	}

	val = cr.Victim()
	if *val != 4 {
		t.Errorf("got %d, want %d", val, 4)
	}
}

func TestClockReplacer_CircularList(t *testing.T) {

	// create list
	list := newCircularList[FrameID, bool](10)

	// test print
	// fmt.Println("list:", list)

	// check fileSize
	if sz := list.size; sz != 0 {
		t.Errorf("got %d, wanted %d\n", sz, 0)
	}

	// insert three even numbers
	list.insert(0, true)
	list.insert(2, true)
	list.insert(4, true)

	// test print
	// fmt.Println("list:", list)

	// check has key
	for i := uint16(0); i <= list.size; i++ {
		keyFound := list.hasKey(FrameID(i))
		if i%2 == 0 && !keyFound {
			t.Errorf("got key=%d (%v), wanted key=%d (%v)\n", i, keyFound, i, !keyFound)
		}
		if i%2 != 0 && keyFound {
			t.Errorf("got key=%d (%v), wanted key=%d (%v)\n", i, !keyFound, i, keyFound)
		}
	}

	// insert some more
	list.insert(0, true)
	list.insert(0, true)
	list.insert(0, true)
	list.insert(3, true)
	list.insert(5, true)

	// expecting fileSize=5, because 0 has been inserted before
	if sz := list.size; sz != 5 {
		t.Errorf("got %d, wanted %d\n", sz, 5)
	}

	// test print
	// fmt.Println("list:", list)

	// test scan
	iter := func(n *dllNode[FrameID, bool]) bool {
		// fmt.Println(n)
		return n != nil
	}
	list.scan(iter)

	// remove
	list.remove(0)
	// fmt.Println(list)
	list.scan(iter)

	// remove another
	list.remove(4)
	// fmt.Println(list)
	list.scan(iter)

	// remove another
	list.remove(5)
	// fmt.Println(list)
	list.scan(iter)

	// remove another
	list.remove(2)
	//	fmt.Println(list)
	list.scan(iter)
}
