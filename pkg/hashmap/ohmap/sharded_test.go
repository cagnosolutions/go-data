package ohmap

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestShardedHashMap(t *testing.T) {
	count := 1000000
	hm := NewShardedHashMap(128)
	for i := 0; i < count; i++ {
		_, ok := hm.Set(strconv.Itoa(i), nil)
		if ok {
			t.Errorf("error: could not located value for key: %q\n", strconv.Itoa(i))
		}
	}
	if hm.Len() != count {
		t.Errorf("error: incorrect count of entries\n")
	}
	fmt.Printf(
		"v2.hashmap containing %d entries is taking %d bytes (%.2f kb, %.2f mb)\n",
		count, util.Sizeof(hm), float64(util.Sizeof(hm)/1024), float64(util.Sizeof(hm)/1024/1024),
	)
	for i := 0; i < count; i++ {
		_, ok := hm.Get(strconv.Itoa(i))
		if !ok {
			t.Errorf("error: could not located value for key: %q\n", strconv.Itoa(i))
		}
	}
	for i := 0; i < count; i++ {
		_, ok := hm.Del(strconv.Itoa(i))
		if !ok {
			t.Errorf("error: could not remove value for key: %q\n", strconv.Itoa(i))
		}
	}
	if hm.Len() != count-count {
		t.Errorf("error: incorrect count of entries\n")
	}
	hm.Close()
}

func TestShardedHashMap_SetAndGetBit(t *testing.T) {
	hm := NewShardedHashMap(128)
	hm.SetBit("mykey", 24, 1)
	hm.SetBit("mykey", 3, 1)
	hm.SetBit("mykey", 4, 1)
	b, ok := hm.GetBit("mykey", 3)
	fmt.Printf("hm.GetBit('mykey', 3)=%.32b (%v)\n", b, ok)
	hm.SetBit("mykey", 3, 0)
	b, ok = hm.GetBit("mykey", 3)
	fmt.Printf("hm.GetBit('mykey', 3)=%.32b (%v)\n", b, ok)
	hm.SetBit("mykey", 24, 1)
	b, ok = hm.GetBit("mykey", 24)
	fmt.Printf("hm.GetBit('mykey', 24)=%.32b (%v)\n", b, ok)
	hm.Close()
}

func TestShardedHashMap_SetAndGetUint(t *testing.T) {
	hm := NewShardedHashMap(128)

	hm.SetUint("counter", 1)
	n, ok := hm.GetUint("counter")
	fmt.Println(n, ok)

	n++
	hm.SetUint("counter", n)
	n, ok = hm.GetUint("counter")
	fmt.Println(n, ok)

	n += 8
	hm.SetUint("counter", n)
	n, ok = hm.GetUint("counter")
	fmt.Println(n, ok)

	hm.Close()
}

func PrintBits(b []byte) {
	// var res string = "16" // set this to the "bit resolution" you'd like to see
	var res = strconv.Itoa(8)
	log.Printf("%."+res+"b (%s bits)", b, res)
}

func Benchmark_ShardedHashMap(b *testing.B) {

	// number of entries
	const count = 100000

	fmt.Printf("Benchmarking %d entries...\n", count)

	// initialize our data
	data := [count]struct {
		key string
		val []byte
	}{}

	// fill out our data to insert
	for j := 0; j < count; j++ {
		data[j] = struct {
			key string
			val []byte
		}{
			key: fmt.Sprintf("key:%4d", j),
			val: []byte(fmt.Sprintf("value-%.6d", j)),
		}
	}

	// create a new hashmap
	hm := NewShardedHashMap(64)

	// load up hashmap with some our data
	for _, entry := range data {
		_, overwrote := hm.Set(entry.key, entry.val)
		if overwrote {
			b.Errorf("overwrote existing value, something went wrong...")
		}
	}

	// benchmark hashing and locating speed
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, entry := range data {
			val, found := hm.Get(entry.key)
			if !found {
				b.Errorf("could not find entry: %q\n", entry.key)
			}
			if !bytes.Equal(val, entry.val) {
				b.Errorf("value does not match: got=%q, wanted=%q\n", val, entry.val)
			}
		}
	}

	// delete all the entries in the hashmap...
	// for j := 0; j < count; j++ {
	// 	key := fmt.Sprintf("key:%.4d", j)
	// 	val := []byte(fmt.Sprintf("value-%.6d", j))
	// 	_, overwrote := hm.Set(key, val)
	// 	if overwrote {
	// 		fmt.Errorf("overwrote existing value, something went wrong...")
	// 	}
	// }

	// range the hashmap and print the data...
	// hm.Range(func(key string, value []byte) bool {
	// 	fmt.Printf("key=%q, value=%q\n", key, value)
	// 	return true
	// })

	// close the hashmap

	fmt.Printf("size: %s", hm.Size())

	hm.Close()
}
