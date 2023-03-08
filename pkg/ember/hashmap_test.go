package ember

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/hash/murmur3"
	"github.com/cagnosolutions/go-data/pkg/util"
)

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
	" ",
	" foo",
	"foo ",
	" foo ",
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
	hm := newHashMap(128, nil)
	for i := 0; i < len(words); i++ {
		hm.set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 35, hm.Len())
	count := hm.Len()
	var stop = hm.Len()
	for i := 0; i < stop; i++ {
		ret, ok := hm.del(words[i])
		util.AssertExpected(t, true, ok)
		util.AssertExpected(t, []byte{0x69}, ret)
		count--
	}
	util.AssertExpected(t, 0, count)
	hm.close()
}

func Test_HashMap_Get(t *testing.T) {
	hm := newHashMap(
		128, func(key string) uint64 {
			return murmur3.Sum64([]byte(key))
		},
	)
	for i := 0; i < len(words); i++ {
		hm.set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 35, hm.Len())
	var count int
	for i := 0; i < hm.Len(); i++ {
		ret, ok := hm.get(words[i])
		util.AssertExpected(t, true, ok)
		util.AssertExpected(t, []byte{0x69}, ret)
		count++
	}
	util.AssertExpected(t, 35, count)
	hm.close()
}

func Test_HashMap_Len(t *testing.T) {
	hm := newHashMap(128, nil)
	for i := 0; i < len(words); i++ {
		hm.set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 35, hm.Len())
	hm.close()
}

func Test_HashMap_PercentFull(t *testing.T) {
	hm := newHashMap(0, nil)
	for i := 0; i < len(words)-10; i++ {
		hm.set(words[i], []byte{0x69})
	}
	percent := fmt.Sprintf("%.2f", hm.percentFull())
	util.AssertExpected(t, "0.78", percent)
	hm.close()
}

func Test_HashMap_Set(t *testing.T) {
	hm := newHashMap(128, nil)
	for i := 0; i < len(words); i++ {
		hm.set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 35, hm.Len())
	hm.close()
}

func Test_HashMap_Range(t *testing.T) {
	hm := newHashMap(128, nil)
	for i := 0; i < len(words); i++ {
		hm.set(words[i], []byte{0x69})
	}
	util.AssertExpected(t, 35, hm.Len())
	var counted int
	hm.Range(
		func(key string, value []byte) bool {
			fmt.Printf("key=%s, value=%q\n", key, value)
			if key != "" && bytes.Equal(value, []byte{0x69}) {
				counted++
				return true
			}
			return false
		},
	)
	util.AssertExpected(t, 35, counted)
	hm.close()
}

var result interface{}

func BenchmarkHashMap_Set1(b *testing.B) {
	hm := newHashMap(128, nil)

	b.ResetTimer()
	b.ReportAllocs()

	var v []byte
	for n := 0; n < b.N; n++ {
		// try to get key/value "foo"
		v, ok := hm.get("foo")
		if !ok {
			// if it doesn't exist, then initialize it
			hm.set("foo", make([]byte, 32))
		} else {
			// if it does exist, then pick a random number between
			// 0 and 256--this will be our bit we try and set
			ri := uint(rand.Intn(128))
			if ok := bitsetHas(&v, ri); !ok {
				// we check the bit to see if it's already set, and
				// then we go ahead and set it if it is not set
				bitsetSet(&v, ri)
			}
			// after this, we make sure to save the bitset back to the hashmap
			if n < 64 {
				fmt.Printf("addr: %p, %+v\n", v, v)
				// PrintBits(v)
			}
			hm.set("foo", v)
		}
	}
	result = v
}

func BenchmarkHashMap_Set2(b *testing.B) {
	hm := newHashMap(128, nil)

	b.ResetTimer()
	b.ReportAllocs()

	var v []byte
	for n := 0; n < b.N; n++ {
		// try to get key/value "foo"
		v, ok := hm.get("foo")
		if !ok {
			// if it doesn't exist, then initialize it
			hm.set("foo", make([]byte, 32))
		} else {
			v = append(v, []byte{byte(n >> 8)}...)
			// after this, we make sure to save the bitset back to the hashmap
			if n < 64 {
				fmt.Printf("addr: %p, %+v\n", v, v)
				// PrintBits(v)
			}
			hm.set("foo", v)
		}
	}
	result = v
}

func TestHashMapMillionEntriesSize(t *testing.T) {
	count := 1000000
	hm := newHashMap(512, nil)
	for i := 0; i < count; i++ {
		_, ok := hm.set(strconv.Itoa(i), nil)
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
		_, ok := hm.get(strconv.Itoa(i))
		if !ok {
			t.Errorf("error: could not located value for key: %q\n", strconv.Itoa(i))
		}
	}
	for i := 0; i < count; i++ {
		_, ok := hm.del(strconv.Itoa(i))
		if !ok {
			t.Errorf("error: could not remove value for key: %q\n", strconv.Itoa(i))
		}
	}
	if hm.Len() != count-count {
		t.Errorf("error: incorrect count of entries\n")
	}
	hm.close()
}
