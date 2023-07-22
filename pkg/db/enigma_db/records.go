package enigma_db

import (
	"sort"
	"strings"
)

type Record[T any] struct {
	Key   string
	Value T
}

type RecordSet[T any] struct {
	Name    string
	Records []Record[T]
}

func (r RecordSet[T]) Len() int {
	return len(r.Records)
}

func (r RecordSet[T]) Less(i, j int) bool {
	return r.Records[i].Key < r.Records[j].Key
}

func (r RecordSet[T]) Swap(i, j int) {
	r.Records[i], r.Records[j] = r.Records[j], r.Records[i]
}

func NewRecordSet[T any](name string) *RecordSet[T] {
	return &RecordSet[T]{
		Name:    name,
		Records: make([]Record[T], 0),
	}
}

func (rs *RecordSet[T]) AddRecord(key string, val T) {
	// create and insert new record
	rs.Records = append(
		rs.Records, Record[T]{
			Key:   key,
			Value: val,
		},
	)
	// make sure they are sorted
	sort.Stable(rs)
}

func (rs *RecordSet[T]) GetRecord(key string) (val T) {
	// perform binary search on records
	i, found := sort.Find(
		rs.Len(), func(i int) int {
			return strings.Compare(key, rs.Records[i].Key)
		},
	)
	if found {
		val = rs.Records[i].Value
	}
	return val
}
