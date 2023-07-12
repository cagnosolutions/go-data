package ohmap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"math/rand"
	"strconv"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/bits"
	"github.com/cagnosolutions/go-data/pkg/util"
)

// 25 words
var words = []string{
	"reproducibility",
	"eruct",
	"acids",
	"flyspecks",
	"driveshafts",
	"volcanically",
	"discouraging",
	"acapnia",
	"phenazines",
	"hoarser",
	"abusing",
	"samara",
	"thromboses",
	"impolite",
	"drivennesses",
	"tenancy",
	"counterreaction",
	"kilted",
	"linty",
	"kistful",
	"biomarkers",
	"infusiblenesses",
	"capsulate",
	"reflowering",
	"heterophyllies",
	"Foo",
	"Bar",
	"foo",
	"bar",
	"FooBar",
	"foobar",
	"",
	" foo",
	"foo ",
	" foo ",
}

func Test_maphashHashFunc(t *testing.T) {
	set := make(map[uint64]string, len(words))
	var h maphash.Hash
	var hash uint64
	var coll int
	for _, word := range words {
		h.Write([]byte(word))
		hash = h.Sum64()
		if old, ok := set[hash]; !ok {
			set[hash] = word
		} else {
			coll++
			fmt.Printf(
				"collision: current word: %s, old word: %s, hash: %d\n", word, old, hash,
			)
		}
	}
	fmt.Printf("encountered %d collisions comparing %d words\n", coll, len(words))
}

func Test_defaultHashFunc(t *testing.T) {
	set := make(map[uint64]string, len(words))
	var hash uint64
	var coll int
	for _, word := range words {
		hash = defaultHashFunc(word)
		if old, ok := set[hash]; !ok {
			set[hash] = word
		} else {
			coll++
			fmt.Printf(
				"collision: current word: %s, old word: %s, hash: %d\n", word, old, hash,
			)
		}
	}
	fmt.Printf("encountered %d collisions comparing %d words\n", coll, len(words))
}

func Test_HashMap_Del(t *testing.T) {
	hm := NewHashMap(128)
	for i := 0; i < len(words); i++ {
		hm.Set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 35, hm.Len())
	count := hm.Len()
	var stop = hm.Len()
	for i := 0; i < stop; i++ {
		ret, ok := hm.Del(words[i])
		util.AssertExpected(t, true, ok)
		util.AssertExpected(t, []byte{0x69}, ret)
		count--
	}
	util.AssertExpected(t, 0, count)
	hm.Close()
}

func Test_HashMap_Get(t *testing.T) {
	hm := NewHashMap(128)
	for i := 0; i < len(words); i++ {
		hm.Set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 35, hm.Len())
	var count int
	for i := 0; i < hm.Len(); i++ {
		ret, ok := hm.Get(words[i])
		util.AssertExpected(t, true, ok)
		util.AssertExpected(t, []byte{0x69}, ret)
		count++
	}
	util.AssertExpected(t, 35, count)
	hm.Close()
}

func Test_HashMap_Len(t *testing.T) {
	hm := NewHashMap(128)
	for i := 0; i < len(words); i++ {
		hm.Set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 35, hm.Len())
	hm.Close()
}

func Test_HashMap_PercentFull(t *testing.T) {
	hm := NewHashMap(0)
	for i := 0; i < len(words)-10; i++ {
		hm.Set(words[i], []byte{0x69})
	}
	percent := fmt.Sprintf("%.2f", hm.PercentFull())
	util.AssertExpected(t, "0.78", percent)
	hm.Close()
}

func Test_HashMap_Set(t *testing.T) {
	hm := NewHashMap(128)
	for i := 0; i < len(words); i++ {
		hm.Set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 25, hm.Len())
	hm.Close()
}

func Test_HashMap_Range(t *testing.T) {
	hm := NewHashMap(128)
	for i := 0; i < len(words); i++ {
		hm.Set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 25, hm.Len())
	var counted int
	hm.Range(
		func(key string, value []byte) bool {
			if key != "" && bytes.Equal(value, []byte{0x69}) {
				counted++
				return true
			}
			return false
		},
	)
	util.AssertExpected(t, 25, counted)
	hm.Close()
}

var result interface{}

func BenchmarkHashMap_Set1(b *testing.B) {
	hm := NewHashMap(128)

	b.ResetTimer()
	b.ReportAllocs()

	var v []byte
	for n := 0; n < b.N; n++ {
		// try to get key/value "foo"
		v, ok := hm.Get("foo")
		if !ok {
			// if it doesn't exist, then initialize it
			hm.Set("foo", make([]byte, 32))
		} else {
			// if it does exist, then pick a random number between
			// 0 and 256--this will be our bit we try and set
			ri := uint(rand.Intn(128))
			if ok := bits.RawBytesHasBit(&v, ri); !ok {
				// we check the bit to see if it's already set, and
				// then we go ahead and set it if it is not set
				bits.RawBytesSetBit(&v, ri)
			}
			// after this, we make sure to save the bitset back to the hashmap
			if n < 64 {
				fmt.Printf("addr: %p, %+v\n", v, v)
				// PrintBits(v)
			}
			hm.Set("foo", v)
		}
	}
	result = v
}

func BenchmarkHashMap_Set2(b *testing.B) {
	hm := NewHashMap(128)

	b.ResetTimer()
	b.ReportAllocs()

	var v []byte
	for n := 0; n < b.N; n++ {
		// try to get key/value "foo"
		v, ok := hm.Get("foo")
		if !ok {
			// if it doesn't exist, then initialize it
			hm.Set("foo", make([]byte, 32))
		} else {
			v = append(v, []byte{byte(n >> 8)}...)
			// after this, we make sure to save the bitset back to the hashmap
			if n < 64 {
				fmt.Printf("addr: %p, %+v\n", v, v)
				// PrintBits(v)
			}
			hm.Set("foo", v)
		}
	}
	result = v
}

func TestHashMapMillionEntriesSize(t *testing.T) {
	count := 1000000
	hm := NewHashMap(512)
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
		"hashmap containing %d entries is taking %d bytes (%.2f kb, %.2f mb)\n",
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

func Benchmark_HashMapVsStdMap(b *testing.B) {

	const count = 10000

	data := [count]struct {
		key string
		val uint32
	}{}

	for i := range data {
		data[i] = struct {
			key string
			val uint32
		}{
			key: fmt.Sprintf("key:%.4d", i),
			val: uint32(i),
		}
	}

	mapSize := func(m map[string][]byte) string {
		format := "map containing %d entries is using %d bytes (%.2f kb, %.2f mb) of ram\n"
		sz := Sizeof(m)
		kb := float64(sz / 1024)
		mb := float64(sz / 1024 / 1024)
		return fmt.Sprintf(format, len(m), sz, kb, mb)
	}

	tests := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{
			"stdMapBench",
			func(b *testing.B) {

				// initialize
				m := make(map[string][]byte, 64)
				var buf [4]byte

				// bench
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, entry := range data {
						binary.LittleEndian.PutUint32(buf[:], entry.val)
						m[entry.key] = nil // buf[:]
					}

				}
				fmt.Printf(mapSize(m))
			},
		},
		{
			"rhhMapBench",
			func(b *testing.B) {

				// initialize
				m := NewHashMap(64)
				var buf [4]byte

				// bench
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, entry := range data {
						binary.LittleEndian.PutUint32(buf[:], entry.val)
						m.Set(entry.key, nil) // buf[:])
					}

				}
				fmt.Println(m.Size())
			},
		},
	}

	for _, test := range tests {
		b.Run(test.name, test.fn)
	}
}

func TestBasicFuncs(t *testing.T) {

	N := 100

	m := NewHashMap(32)

	// setting data...
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("key-%.4d", i)
		_, updated := m.Set(key, nil)
		if updated {
			t.Fatalf("should not have been a previous value to updated...\n")
		}
	}

	fmt.Println("Map Details...\n", m.Details())

	var keys []string
	var ki int

	// ranging data...
	m.Range(
		func(key string, val []byte) bool {
			if key == "key-0000" {
				ki = len(keys)
			}
			keys = append(keys, key)
			fmt.Printf("key=%q, val=%v\n", key, val)
			return true
		},
	)

	if len(keys) != m.Len() {
		t.Fatalf("ranged %d keys, but map says it has %d keys...\n", len(keys), m.Len())
	}

	for i := 0; i < 25; i++ {
		key := keys[ki]
		hashkey := m.hash(key)
		idx := hashkey & m.mask
		buk := m.getBucket(idx)
		hash := buk.getHash()
		fmt.Printf("key=%q, hashkey=%v, hash=%v, index=%d, bucket=%s\n", key, hashkey, hash, idx, buk)
	}

	// getting data...
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("key-%.4d", i)
		val, found := m.Get(key)
		if !found {
			t.Fatalf("should have been able to find %q... (val=%v, found=%v)\n", key, val, found)
		}
		if val != nil {
			t.Fatalf("incorrect value: got=%v, wanted=%v\n", val, i)
		}
	}

	// ranging data...
	m.Range(
		func(key string, val []byte) bool {
			fmt.Printf("key=%q, val=%v\n", key, val)
			return true
		},
	)

	// deleting 1/2 the data...
	for i := 0; i < N; i++ {
		// just delete the even entries
		if i%2 == 0 {
			_, removed := m.Del(fmt.Sprintf("key-%.4d", i))
			if !removed {
				t.Fatalf("should have been able to find and remove...\n")
			}
		}
	}

	fmt.Println("[Removed 1/2 of the Data]\n", m.Details())

	// // ranging data again...
	// m.Range(
	// 	func(key string, val []byte) bool {
	// 		fmt.Printf("key=%q, val=%v\n", key, val)
	// 		return true
	// 	},
	// )

	// deleting the remaining 1/2 of the data...
	for i := 0; i < N; i++ {
		// just delete the odd entries
		if i%2 != 0 {
			_, removed := m.Del(fmt.Sprintf("key-%.4d", i))
			if !removed {
				t.Fatalf("should have been able to find and remove...\n")
			}
		}
	}

	m.Set("one", []byte{1})
	m.Set("two", []byte{2})
	m.Set("three", []byte{3})

	// ranging data again (there should only be three entries)...
	// m.Range(
	// 	func(key string, val []byte) bool {
	// 		fmt.Printf("key=%q, val=%v\n", key, val)
	// 		return true
	// 	},
	// )

	fmt.Println("[Removed Remaining Data]\n", m.Details())

}
