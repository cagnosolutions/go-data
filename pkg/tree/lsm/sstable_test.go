package lsm

import (
	"fmt"
	"testing"
	"time"
)

func TestSSTable_Get(t *testing.T) {
	// Example usage:
	entries := []Entry{
		{Key: "a1", Val: "a1 value"},
		{Key: "a4", Val: "a4 value"},
		{Key: "b6", Val: "b6 value"},
		{Key: "b9", Val: "b9 value"},
		{Key: "c1", Val: "c1 value"},
		{Key: "c3", Val: "c3 value"},
		{Key: "c4", Val: "c4 is a bomb"},
		{Key: "c5", Val: "c5 is not a bomb"},
		{Key: "d7", Val: "d7 sounds like a vitamin"},
		{Key: "d8", Val: "d8 sounds like a rap group"},
		{Key: "d9", Val: "d9 is fun"},
		{Key: "e1", Val: "e1 reminds me of a lecture hall"},
		{Key: "e2", Val: "e2 is neat"},
		{Key: "f4", Val: "f4 sounds like a racecar race"},
		{Key: "f5", Val: "f5 is a button my keyboard"},
		{Key: "f6", Val: "f6 soulds like a faster race"},
		{Key: "f7", Val: "f7 should be the name of a jet"},
		{Key: "h1", Val: "h1 is an html tag"},
		{Key: "h2", Val: "h2, same as h1"},
		{Key: "j4", Val: "j4 should be the name of a little league pickup team"},
		{Key: "j6", Val: "j6 i dunno"},
		{Key: "k6", Val: "k6 sounds like a fun new shoe"},
		{Key: "k8", Val: "k8 sounds like a virus"},
		{Key: "l1", Val: "l1 is boring"},
		{Key: "l3", Val: "l3 is even more boring"},
		{Key: "l5", Val: "l5--I am getting sick of this"},
		{Key: "n2", Val: "n2, is something"},
		{Key: "n3", Val: "n3 is also something"},
		{Key: "n5", Val: "n5 is a thing I don't want to type"},
		{Key: "n7", Val: "n7 sounds like a TV rating"},
		{Key: "p3", Val: "p3 is close to ps3"},
		{Key: "p5", Val: "p5 is close to ps5"},
		{Key: "r1", Val: "r1 is a motorcycle"},
		{Key: "s3", Val: "s3 should be a motorcycle"},
	}

	_ = entries

	// Create the SSTable.
	sstable, err := OpenSSTable("sstable_test/")
	if err != nil {
		t.Errorf("error creating sstable: %s", err)
	}
	defer sstable.Close()

	// Write data to the table
	// err = sstable.WriteBatch(entries)
	// if err != nil {
	// 	t.Errorf("error writing batch: %s", err)
	// }

	// Get values from the SSTable.
	key := "j4"
	value, exists, err := sstable.GetBinary(key)
	if err != nil {
		fmt.Println("Error retrieving value from SSTable:", err)
		return
	}
	if exists {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	} else {
		fmt.Printf("Key not found: %s\n", key)
	}
}

var entries = []Entry{
	{Key: "b6", Val: "b6 value"},
	{Key: "d7", Val: "d7 sounds like a vitamin"},
	{Key: "f5", Val: "f5 is a button my keyboard"},
	{Key: "j4", Val: "j4 should be the name of a little league pickup team"},
	{Key: "k8", Val: "k8 sounds like a virus"},
	{Key: "l5", Val: "l5--I am getting sick of this"},
	{Key: "n7", Val: "n7 sounds like a TV rating"},
	{Key: "p5", Val: "p5 is close to ps5"},
	{Key: "r1", Val: "r1 is a motorcycle"},
	{Key: "s3", Val: "s3 should be a motorcycle"},
}

func get(b *testing.B) {
	// Open the SSTable.
	sstable, err := OpenSSTable("sstable_test/")
	if err != nil {
		b.Errorf("error creating sstable: %s", err)
	}
	defer sstable.Close()

	time.Sleep(500 * time.Millisecond)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := range entries {
			val, exists, err := sstable.Get(entries[j].Key)
			if err != nil {
				b.Errorf("got error: %s", err)
			}
			if !exists {
				b.Errorf("couldn't find key: %q\n", entries[j].Key)
			}
			if val != entries[j].Val {
				b.Errorf("values don'tmatch: got=%v, want=%v\n", val, entries[j].Val)
			}
		}
	}

}

func getBin(b *testing.B) {
	// Open the SSTable.
	sstable, err := OpenSSTable("sstable_test/")
	if err != nil {
		b.Errorf("error creating sstable: %s", err)
	}
	defer sstable.Close()

	time.Sleep(500 * time.Millisecond)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := range entries {
			val, exists, err := sstable.GetBinary(entries[j].Key)
			if err != nil {
				b.Errorf("got error: %s", err)
			}
			if !exists {
				b.Errorf("couldn't find key: %q\n", entries[j].Key)
			}
			if val != entries[j].Val {
				b.Errorf("values don'tmatch: got=%v, want=%v\n", val, entries[j].Val)
			}
		}
	}

}

func BenchmarkGetAndGetBinary(b *testing.B) {

	tests := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{
			"get",
			get,
		},
		{
			"getBin",
			getBin,
		},
	}

	for _, test := range tests {
		b.Run(test.name, test.fn)
	}

}
