package generic

import (
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func init() {
	// var seed int64 = 1519776033517775607
	seed := (time.Now().UnixNano())
	println("seed:", seed)
	rand.Seed(seed)
}

func TestHashDIB(t *testing.T) {
	var b bucket[string, interface{}]
	b.dib = 100
	b.hashkey = 90000
	if b.dib != 100 {
		t.Fatalf("expected %v, got %v", 100, b.dib)
	}
	if b.hashkey != 90000 {
		t.Fatalf("expected %v, got %v", 90000, b.hashkey)
	}
}

func TestBasicFuncs(t *testing.T) {

	N := 100

	m := NewMap[string, any](32)

	// setting data...
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("key-%.4d", i)
		_, updated := m.Set(key, i)
		if updated {
			t.Fatalf("should not have been a previous value to updated...\n")
		}
	}

	fmt.Println("[Added Data]\n", m.Details())

	var keys []string
	var ki int

	// ranging data...
	m.Range(
		func(key string, val any) bool {
			if key == "key-0000" {
				ki = len(keys)
			}
			keys = append(keys, key)
			// fmt.Printf("key=%q, val=%v\n", key, val)
			return true
		},
	)

	if len(keys) != m.Len() {
		t.Fatalf("ranged %d keys, but map says it has %d keys...\n", len(keys), m.Len())
	}

	// for i := 0; i < 25; i++ {
	// 	key := keys[ki]
	// 	hash := m.hash(key)
	// 	idx := hash & m.mask
	// 	buk := m.getBucket(idx)
	// 	fmt.Printf("key=%q, hash=%v, index=%d, bucket=%s\n", key, hash, idx, buk)
	// }
	_ = ki

	// getting data...
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("key-%.4d", i)
		val, found := m.Get(key)
		if !found {
			t.Fatalf("should have been able to find %q... (val=%v, found=%v)\n", key, val, found)
		}
		if val != i {
			t.Fatalf("incorrect value: got=%v, wanted=%v\n", val, i)
		}
	}

	// ranging data...
	// m.Range(
	// 	func(key string, val any) bool {
	// 		fmt.Printf("key=%q, val=%v\n", key, val)
	// 		return true
	// 	},
	// )

	fmt.Println("Deleting 1/2 of the data...")
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

	// ranging data again...
	// m.Range(
	// 	func(key string, val any) bool {
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

	m.Set("one", 1)
	m.Set("two", 2)
	m.Set("three", 3)

	fmt.Println("[Removed Remaining Data (except for 3 entries)]\n", m.Details())

	// ranging data again (there should only be three entries)...
	// m.Range(
	// 	func(key string, val any) bool {
	// 		fmt.Printf("key=%q, val=%v\n", key, val)
	// 		return true
	// 	},
	// )

}

func TestRandomData(t *testing.T) {
	N := 10000
	start := time.Now()
	for time.Since(start) < time.Second*2 {
		nums := random(N, true)
		var m *Map[string, any]
		switch rand.Int() % 5 {
		default:
			log.Println("Going with default case...")
			m = NewMap[string, any](uint(N / ((rand.Int() % 3) + 1)))
		case 1:
			log.Println("Going with default case #1...")
			m = new(Map[string, any])
		case 2:
			log.Println("Going with default case #2...")
			m = NewMap[string, any](0)
		}
		v, ok := m.Get(k(999))
		if ok || v != nil {
			t.Fatalf("expected %v, got %v", nil, v)
		}
		v, ok = m.Del(k(999))
		if ok || v != nil {
			t.Fatalf("expected %v, got %v", nil, v)
		}
		if m.Len() != 0 {
			t.Fatalf("expected %v, got %v", 0, m.Len())
		}
		// set a bunch of items
		for i := 0; i < len(nums); i++ {
			v, ok := m.Set(nums[i], nums[i])
			if ok || v != nil {
				t.Fatalf("expected %v, got %v", nil, v)
			}
		}
		if m.Len() != N {
			t.Fatalf("expected %v, got %v", N, m.Len())
		}
		// retrieve all the items
		shuffle(nums)
		for i := 0; i < len(nums); i++ {
			v, ok := m.Get(nums[i])
			if !ok || v == nil || v != nums[i] {
				t.Fatalf("expected %v, got %v", nums[i], v)
			}
		}
		// replace all the items
		shuffle(nums)
		for i := 0; i < len(nums); i++ {
			v, ok := m.Set(nums[i], add(nums[i], 1))
			if !ok || v != nums[i] {
				t.Fatalf("expected %v, got %v", nums[i], v)
			}
		}
		if m.Len() != N {
			t.Fatalf("expected %v, got %v", N, m.Len())
		}
		// retrieve all the items
		shuffle(nums)
		for i := 0; i < len(nums); i++ {
			v, ok := m.Get(nums[i])
			if !ok || v != add(nums[i], 1) {
				t.Fatalf("expected %v, got %v", add(nums[i], 1), v)
			}
		}
		// remove half the items
		shuffle(nums)
		for i := 0; i < len(nums)/2; i++ {
			v, ok := m.Del(nums[i])
			if !ok || v != add(nums[i], 1) {
				t.Fatalf("expected %v, got %v", add(nums[i], 1), v)
			}
		}
		if m.Len() != N/2 {
			t.Fatalf("expected %v, got %v", N/2, m.Len())
		}
		// check to make sure that the items have been removed
		for i := 0; i < len(nums)/2; i++ {
			v, ok := m.Get(nums[i])
			if ok || v != nil {
				t.Fatalf("expected %v, got %v", nil, v)
			}
		}
		// check the second half of the items
		for i := len(nums) / 2; i < len(nums); i++ {
			v, ok := m.Get(nums[i])
			if !ok || v != add(nums[i], 1) {
				t.Fatalf("expected %v, got %v", add(nums[i], 1), v)
			}
		}
		// try to delete again, make sure they don't exist
		for i := 0; i < len(nums)/2; i++ {
			v, ok := m.Del(nums[i])
			if ok || v != nil {
				t.Fatalf("expected %v, got %v", nil, v)
			}
		}
		if m.Len() != N/2 {
			t.Fatalf("expected %v, got %v", N/2, m.Len())
		}
		m.Range(
			func(key keyT, value valueT) bool {
				if value != add(key, 1) {
					t.Fatalf("expected %v, got %v", add(key, 1), value)
				}
				return true
			},
		)
		var n int
		m.Range(
			func(key keyT, value valueT) bool {
				n++
				return false
			},
		)
		if n != 1 {
			t.Fatalf("expected %v, got %v", 1, n)
		}
		for i := len(nums) / 2; i < len(nums); i++ {
			v, ok := m.Del(nums[i])
			if !ok || v != add(nums[i], 1) {
				t.Fatalf("expected %v, got %v", add(nums[i], 1), v)
			}
		}
	}
}

func TestBench(t *testing.T) {
	// N, _ := strconv.ParseUint(os.Getenv("MAPBENCH"), 10, 64)
	// if N == 0 {
	// 	fmt.Printf("Enable benchmarks with MAPBENCH=1000000\n")
	// 	return
	// }
	const N = 1000000

	var pnums []int
	for i := 0; i < int(N); i++ {
		pnums = append(pnums, i)
	}

	{
		fmt.Printf("\n## STRING KEYS\n\n")
		nums := random(int(N), false)
		t.Run(
			"sac", func(t *testing.T) {
				testPerf(nums, pnums, "sac")
			},
		)
		t.Run(
			"Stdlib", func(t *testing.T) {
				testPerf(nums, pnums, "stdlib")
			},
		)
	}
	{
		fmt.Printf("\n## INT KEYS\n\n")
		nums := rand.Perm(int(N))
		t.Run(
			"sac", func(t *testing.T) {
				testPerf(nums, pnums, "sac")
			},
		)
		t.Run(
			"Stdlib", func(t *testing.T) {
				testPerf(nums, pnums, "stdlib")
			},
		)
	}

}

type keyT = string
type valueT = interface{}

func k(key int) keyT {
	return strconv.FormatInt(int64(key), 10)
}

func add(x keyT, delta int) int {
	i, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		panic(err)
	}
	return int(i + int64(delta))
}

func random(N int, perm bool) []keyT {
	nums := make([]keyT, N)
	if perm {
		for i, x := range rand.Perm(N) {
			nums[i] = k(x)
		}
	} else {
		m := make(map[keyT]bool)
		for len(m) < N {
			m[k(int(rand.Uint64()))] = true
		}
		var i int
		for k := range m {
			nums[i] = k
			i++
		}
	}
	return nums
}

func shuffle[K comparable](nums []K) {
	for i := range nums {
		j := rand.Intn(i + 1)
		nums[i], nums[j] = nums[j], nums[i]
	}
}

func printItem(s string, size int, dir int) {
	for len(s) < size {
		if dir == -1 {
			s += " "
		} else {
			s = " " + s
		}
	}
	fmt.Printf("%s ", s)
}

func commaize(n int) string {
	s1, s2 := fmt.Sprintf("%d", n), ""
	for i, j := len(s1)-1, 0; i >= 0; i, j = i-1, j+1 {
		if j%3 == 0 && j != 0 {
			s2 = "," + s2
		}
		s2 = string(s1[i]) + s2
	}
	return s2
}

func testPerf[K comparable, V any](nums []K, pnums []V, which string) {
	var ms1, ms2 runtime.MemStats
	initSize := 0 // len(nums) * 2
	defer func() {
		heapBytes := int(ms2.HeapAlloc - ms1.HeapAlloc)
		fmt.Printf(
			"memory %13s bytes %19s/entry \n",
			commaize(heapBytes), commaize(heapBytes/len(nums)),
		)
		fmt.Printf("\n")
	}()
	runtime.GC()
	time.Sleep(time.Millisecond * 100)
	runtime.ReadMemStats(&ms1)

	var setop, getop, delop func(int, int)
	var scnop func()
	switch which {
	case "stdlib":
		m := make(map[K]V, initSize)
		setop = func(i, _ int) { m[nums[i]] = pnums[i] }
		getop = func(i, _ int) { _ = m[nums[i]] }
		delop = func(i, _ int) { delete(m, nums[i]) }
		scnop = func() {
			for range m {
			}
		}
	case "sac":
		var m Map[K, V]
		setop = func(i, _ int) { m.Set(nums[i], pnums[i]) }
		getop = func(i, _ int) { m.Get(nums[i]) }
		delop = func(i, _ int) { m.Del(nums[i]) }
		scnop = func() {
			m.Range(
				func(key K, value V) bool {
					return true
				},
			)
		}
	}
	fmt.Printf("-- %s --", which)
	fmt.Printf("\n")

	ops := []func(int, int){setop, getop, setop, nil, delop}
	tags := []string{"set", "get", "reset", "scan", "delete"}
	for i := range ops {
		shuffle(nums)
		var na bool
		var n int
		start := time.Now()
		if tags[i] == "scan" {
			op := scnop
			if op == nil {
				na = true
			} else {
				n = 20
				for i := 0; i < n; i++ {
					op()
				}
			}
		} else {
			n = len(nums)
			for j := 0; j < n; j++ {
				ops[i](j, 1)
			}
		}
		dur := time.Since(start)
		if i == 0 {
			runtime.GC()
			time.Sleep(time.Millisecond * 100)
			runtime.ReadMemStats(&ms2)
		}
		printItem(tags[i], 9, -1)
		if na {
			printItem("-- unavailable --", 14, 1)
		} else {
			if n == -1 {
				printItem("unknown ops", 14, 1)
			} else {
				printItem(fmt.Sprintf("%s ops", commaize(n)), 14, 1)
			}
			printItem(fmt.Sprintf("%.0fms", dur.Seconds()*1000), 8, 1)
			if n != -1 {
				printItem(fmt.Sprintf("%s/sec", commaize(int(float64(n)/dur.Seconds()))), 18, 1)
			}
		}
		fmt.Printf("\n")
	}
}

func BenchmarkHashFuncs(b *testing.B) {
	tests := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{
			"default",
			func(b *testing.B) {

				m := NewMap[string, int](64)

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					for j := 0; j < 50; j++ {
						m.Set(strconv.Itoa(j), j)
					}
					for j := 0; j < 50; j++ {
						_, found := m.Get(strconv.Itoa(j))
						if !found {
							b.Errorf("key %q not found", strconv.Itoa(j))
						}
					}
					for j := 0; j < 50; j++ {
						k := strconv.Itoa(j)
						m.Del(k)
						_, found := m.Get(k)
						if found {
							b.Errorf("key %q found, but should be gone", k)
						}
					}
				}
			},
		},
		{
			"custom",
			func(b *testing.B) {

				m := NewMapWithHashFunc[string, int](64, NewHasher64[string](fnv.New64a()))

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					for j := 0; j < 50; j++ {
						m.Set(strconv.Itoa(j), j)
					}
					for j := 0; j < 50; j++ {
						_, found := m.Get(strconv.Itoa(j))
						if !found {
							b.Errorf("key %q not found", strconv.Itoa(j))
						}
					}
					for j := 0; j < 50; j++ {
						k := strconv.Itoa(j)
						m.Del(k)
						_, found := m.Get(k)
						if found {
							b.Errorf("key %q found, but should be gone", k)
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, tt.fn)
	}
}
