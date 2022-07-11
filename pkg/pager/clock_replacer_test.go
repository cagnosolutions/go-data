package pager

import (
	"testing"
)

func TestClockReplacer_All(t *testing.T) {
	cr := newClockReplacer(10)
	cr.unpin(1)
	cr.unpin(2)
	cr.unpin(3)
	cr.unpin(4)
	cr.unpin(5)
	cr.unpin(6)
	cr.unpin(1)

	ans := cr.size()
	if ans != 6 {
		t.Errorf("got %d, want %d", ans, 6)
	}

	val := cr.victim()
	if *val != 1 {
		t.Errorf("got %d, want %d", val, 1)
	}

	val = cr.victim()
	if *val != 2 {
		t.Errorf("got %d, want %d", val, 2)
	}

	val = cr.victim()
	if *val != 3 {
		t.Errorf("got %d, want %d", val, 3)
	}

	cr.pin(3)
	cr.pin(4)
	ans = cr.size()
	if ans != 2 {
		t.Errorf("got %d, want %d", ans, 2)
	}

	cr.unpin(4)

	val = cr.victim()
	if *val != 5 {
		t.Errorf("got %d, want %d", val, 5)
	}

	val = cr.victim()
	if *val != 6 {
		t.Errorf("got %d, want %d", val, 6)
	}

	val = cr.victim()
	if *val != 4 {
		t.Errorf("got %d, want %d", val, 4)
	}
}

func TestClockReplacer_CircularList(t *testing.T) {

	// create list
	list := newCircularList(10)

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
		keyFound := list.hasKey(frameID(i))
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
	iter := func(n *node) bool {
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
