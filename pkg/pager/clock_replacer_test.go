package pager

import (
	"fmt"
	"testing"
)

func TestClockReplacer(t *testing.T) {
	cr := newClockReplacer[frameID, *frame](10)
	cr.unpin(1, nil)
	cr.unpin(2, nil)
	cr.unpin(3, nil)
	cr.unpin(4, nil)
	cr.unpin(5, nil)
	cr.unpin(6, nil)
	cr.unpin(1, nil)

	ans := cr.size()
	if ans != 6 {
		t.Errorf("got %d, want %d", ans, 6)
	}

	val, _ := cr.victim()
	if val != 1 {
		t.Errorf("got %d, want %d", val, 1)
	}

	val, _ = cr.victim()
	if val != 2 {
		t.Errorf("got %d, want %d", val, 2)
	}

	val, _ = cr.victim()
	if val != 3 {
		t.Errorf("got %d, want %d", val, 3)
	}

	cr.pin(3)
	cr.pin(4)
	ans = cr.size()
	if ans != 2 {
		t.Errorf("got %d, want %d", ans, 2)
	}

	cr.unpin(4, nil)

	val, _ = cr.victim()
	if val != 5 {
		t.Errorf("got %d, want %d", val, 5)
	}

	val, _ = cr.victim()
	if val != 6 {
		t.Errorf("got %d, want %d", val, 6)
	}

	val, _ = cr.victim()
	if val != 4 {
		t.Errorf("got %d, want %d", val, 4)
	}
}

func TestCircularList(t *testing.T) {

	// create list
	list := newCircularList[int, bool](10)

	// test print
	fmt.Println("list:", list)

	// check size
	if sz := list.size; sz != 0 {
		t.Errorf("got %d, wanted %d\n", sz, 0)
	}

	// insert three even numbers
	list.insert(0, true)
	list.insert(2, true)
	list.insert(4, true)

	// test print
	fmt.Println("list:", list)

	// check has key
	for i := 0; i <= list.size; i++ {
		keyFound := list.hasKey(i)
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

	// expecting size=5, because 0 has been inserted before
	if sz := list.size; sz != 5 {
		t.Errorf("got %d, wanted %d\n", sz, 5)
	}

	// test print
	fmt.Println("list:", list)

	// test scan
	iter := func(n *node[int, bool]) bool {
		fmt.Println(n)
		return n != nil
	}
	list.scan(iter)

	// remove
	list.remove(0)
	fmt.Println(list)
	list.scan(iter)

	// remove another
	list.remove(4)
	fmt.Println(list)
	list.scan(iter)

	// remove another
	list.remove(5)
	fmt.Println(list)
	list.scan(iter)

	// remove another
	list.remove(2)
	fmt.Println(list)
	list.scan(iter)
}
