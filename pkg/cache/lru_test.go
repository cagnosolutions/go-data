package cache

import (
	"fmt"
	"testing"
)

func TestLRU(t *testing.T) {
	lru := NewLRU[int, string](4)
	lru.Set(1, "one")
	lru.Set(2, "two")
	lru.Set(3, "three")
	lru.Set(4, "four")
	fmt.Println(lru)
	k, v := lru.Evict()
	fmt.Println(k, v)
	if k != 1 {
		t.Errorf("got: %v, expected: %v\n", k, 1)
	}
	fmt.Println(lru)
	lru.Set(5, "five")
	lru.Set(6, "six")
	lru.Set(7, "seven")
	fmt.Println(lru)
	lru.Evict()
	fmt.Println(lru)
	lru.SetEvicted(8, "eight")
	fmt.Println(lru)
	lru.Evict()
	fmt.Println(lru)
}
