package page

import (
	"fmt"
	"testing"
	"time"

	"github.com/cagnosolutions/go-data/pkg/dbms/errs"
)

const (
	recCount = 88
	recSize  = 16
)

func BenchmarkPage_AddRecords_2K(b *testing.B) {
	pgSize := uint16(2 << 10)
	var rCount int

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pp := []page{
			newPageSize(1, pgSize),
			newPageSize(2, pgSize),
			newPageSize(3, pgSize),
			newPageSize(4, pgSize),
			newPageSize(5, pgSize),
			newPageSize(6, pgSize),
			newPageSize(7, pgSize),
			newPageSize(8, pgSize),
			newPageSize(9, pgSize),
			newPageSize(10, pgSize),
			newPageSize(11, pgSize),
			newPageSize(12, pgSize),
			newPageSize(13, pgSize),
			newPageSize(14, pgSize),
			newPageSize(15, pgSize),
			newPageSize(16, pgSize),
		}
		b.StartTimer()
		rCount = 0
		ts := time.Now()
		for _, p := range pp {
			for {
				_, err := p.addRecord([]byte(fmt.Sprintf(fmt.Sprintf("r-%%.%dd", recSize-1), i)))
				if err != nil {
					if err == errs.ErrNoRoom {
						break
					}
					b.Error(err)
				}
				rCount++
			}
		}
		b.ReportMetric(float64(time.Since(ts)), "ns/op")
		b.ReportMetric(float64(rCount), "recs/op")
		b.SetBytes(int64(pgSize))
	}
}

func BenchmarkPage_AddRecords_4K(b *testing.B) {
	pgSize := uint16(4 << 10)
	var rCount int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pp := []page{
			newPageSize(1, pgSize),
			newPageSize(2, pgSize),
			newPageSize(3, pgSize),
			newPageSize(4, pgSize),
			newPageSize(5, pgSize),
			newPageSize(6, pgSize),
			newPageSize(7, pgSize),
			newPageSize(8, pgSize),
		}
		b.StartTimer()
		rCount = 0
		for _, p := range pp {
			b.ReportAllocs()
			for {
				_, err := p.addRecord([]byte(fmt.Sprintf(fmt.Sprintf("r-%%.%dd", recSize-1), i)))
				if err != nil {
					if err == errs.ErrNoRoom {
						break
					}
					b.Error(err)
				}
				rCount++
			}
		}
		b.ReportMetric(float64(rCount), "recs/op")
		b.SetBytes(int64(pgSize))
	}
}

func BenchmarkPage_AddRecords_8K(b *testing.B) {
	pgSize := uint16(8 << 10)
	var rCount int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pp := []page{
			newPageSize(1, pgSize),
			newPageSize(2, pgSize),
			newPageSize(3, pgSize),
			newPageSize(4, pgSize),
		}
		b.StartTimer()
		rCount = 0
		for _, p := range pp {
			b.ReportAllocs()
			for {
				_, err := p.addRecord([]byte(fmt.Sprintf(fmt.Sprintf("r-%%.%dd", recSize-1), i)))
				if err != nil {
					if err == errs.ErrNoRoom {
						break
					}
					b.Error(err)
				}
				rCount++
			}
		}
		b.ReportMetric(float64(rCount), "recs/op")
		b.SetBytes(int64(pgSize))
	}
}

func BenchmarkPage_AddRecords_16K(b *testing.B) {
	pgSize := uint16(16 << 10)
	var rCount int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pp := []page{
			newPageSize(1, pgSize),
			newPageSize(2, pgSize),
		}
		b.StartTimer()
		rCount = 0
		for _, p := range pp {
			b.ReportAllocs()
			for {
				_, err := p.addRecord([]byte(fmt.Sprintf(fmt.Sprintf("r-%%.%dd", recSize-1), i)))
				if err != nil {
					if err == errs.ErrNoRoom {
						break
					}
					b.Error(err)
				}
				rCount++
			}
		}
		b.ReportMetric(float64(rCount), "recs/op")
		b.SetBytes(int64(pgSize))
	}
}

func BenchmarkPage_AddRecords_32K(b *testing.B) {
	pgSize := uint16(32 << 10)
	var rCount int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pp := []page{
			newPageSize(1, pgSize),
		}
		b.StartTimer()
		rCount = 0
		for _, p := range pp {
			b.ReportAllocs()
			for {
				_, err := p.addRecord([]byte(fmt.Sprintf(fmt.Sprintf("r-%%.%dd", recSize-1), i)))
				if err != nil {
					if err == errs.ErrNoRoom {
						break
					}
					b.Error(err)
				}
				rCount++
			}
		}
		b.ReportMetric(float64(rCount), "recs/op")
		b.SetBytes(int64(pgSize))
	}
}

const (
	kb = 1 << 10
)

var add = func(p page, recCount, recSize int) (int, error) {
	fmt.Printf(">> adding: %d records to go!\n", recCount)
	for i := 0; i < recCount; i++ {
		rec := fmt.Sprintf(fmt.Sprintf("r-%%.%dd", recSize-1), i)
		_, err := p.addRecord([]byte(rec))
		if err != nil {
			return i, err
		}
	}
	return recCount, nil
}
