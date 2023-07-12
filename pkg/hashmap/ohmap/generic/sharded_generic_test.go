package generic

import (
	"fmt"
	"testing"
)

var data1 = []struct {
	key string
	val any
}{
	// random data
	{"one", 1},
	{"two", 2},
	{"three", 3},
	{"four", 4},
	{"five", 5},
	{"user", "some random user"},
	{"users", "a group of users"},
	{"order", "a random order"},
	{"prefix", "a standalone prefix"},

	// users
	{"users:001", "user 1"},
	{"users:002", "user 2"},
	{"users:003", "user 3"},
	{"users:004", "user 4"},

	// orders
	{"order:099", "order 99"},
	{"order:105", "order 105"},
	{"order:227", "order 227"},
	{"order:072", "order 72"},
	{"order:044", "order 44"},
	{"order:967", "order 967"},
	{"order:382", "order 382"},

	// custom prefix
	{"prefix:22", "prefix 22"},
	{"prefix:55", "prefix 55"},
	{"prefix:77", "prefix 77"},
}

func TestPrintDetails(t *testing.T) {
	sm := NewShardedMap[string, any](16)
	for _, ent := range data1 {
		sm.Set(ent.key, ent.val)
	}
	fmt.Println(sm.Details())
}

func TestGetCollection(t *testing.T) {
	sm := NewShardedMap[string, any](16)
	for _, ent := range data1 {
		sm.Set(ent.key, ent.val)
	}

	// get all users
	got, count := sm.GetCollection("users:")
	if count != 4 {
		t.Fatalf("got=%d, expected=4", count)
	}
	fmt.Println("users:\n", got)

	// get all orders
	got, count = sm.GetCollection("order:")
	if count != 7 {
		t.Fatalf("got=%d, expected=7", count)
	}
	fmt.Println("orders:\n", got)

	// get all custom prefix
	got, count = sm.GetCollection("prefix:")
	if count != 3 {
		t.Fatalf("got=%d, expected=3", count)
	}
	fmt.Println("prefix:\n", got)
}
