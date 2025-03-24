package bpt

//
import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func BenchmarkFillFactors(b *testing.B) {
	fillFactors := []int{10, 100, 1000, 10000, 1000000}

	var t *BPTree
	var err error

	for _, numItems := range fillFactors {
		b.Run(benchmarkName(numItems), func(b *testing.B) {
			b.ReportAllocs()
			rand.Seed(time.Now().UnixNano())

			// Benchmark fill factor
			b.ResetTimer()
			for i := 0; i < b.N; i++ {

				// create new bpt
				t, err = NewBPTree()
				if err != nil {
					b.Errorf("Error creating BPT: %v", err)
				}

				// fill with data
				for j := 0; j < numItems; j++ {
					data := rand.Uint32()
					t.Add(keyType{data}, valType{[]byte{byte(data), byte(data >> 8), byte(data >> 16), byte(data >> 2)}})
				}

				// close tree
				t.Close()
			}
		})
	}
}

func benchmarkName(numItems int) string {
	return "FillFactor_" + strconv.Itoa(numItems)
}
