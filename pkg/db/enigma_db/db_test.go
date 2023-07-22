package enigma_db

import (
	"fmt"
	"strconv"
	"testing"
)

type DB1 struct {
	index      map[[2]string]int
	recordSets map[string]RecordSet[any]
}

func NewDB1() *DB1 {
	return &DB1{
		index:      make(map[[2]string]int),
		recordSets: make(map[string]RecordSet[any]),
	}
}

func (db1 *DB1) MakeRS(name string) {
	_, found := db1.recordSets[name]
	if found {
		return
	}
	set := NewRecordSet[any](name)
	db1.recordSets[name] = *set
}

func (db1 *DB1) Set(rs string, key string, val any) int {
	set, found := db1.recordSets[rs]
	if !found {
		return -1
	}
	set.Records = append(
		set.Records, Record[any]{
			Key:   key,
			Value: val,
		},
	)
	off := len(set.Records) - 1
	db1.index[[2]string{rs, key}] = off
	db1.recordSets[rs] = set
	return off
}

func (db1 *DB1) Get(rs string, key string) (any, bool) {
	set, found := db1.recordSets[rs]
	if !found {
		return nil, false
	}
	off, found := db1.index[[2]string{rs, key}]
	if !found {
		return nil, false
	}
	return set.Records[off].Value.(any), true
}

func BenchmarkDB1andDB2(b *testing.B) {
	tests := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{
			"db1",
			func(b *testing.B) {

				db := NewDB1()

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					db.MakeRS("users")
					db.MakeRS("orders")
					db.MakeRS("invoices")

					// add some users
					db.Set("users", "1", "I am user 1")
					db.Set("users", "2", "I am user 2")
					db.Set("users", "3", "I am user 3")
					db.Set("users", "4", "I am user 4")
					db.Set("users", "5", "I am user 5")

					// add some orders
					db.Set("orders", "1", "I am order 1")
					db.Set("orders", "2", "I am order 2")
					db.Set("orders", "3", "I am order 3")
					db.Set("orders", "4", "I am order 4")
					db.Set("orders", "5", "I am order 5")
					db.Set("orders", "6", "I am order 6")
					db.Set("orders", "7", "I am order 7")
					db.Set("orders", "8", "I am order 8")
					db.Set("orders", "9", "I am order 9")
					db.Set("orders", "10", "I am order 10")

					// add some invoices
					db.Set("invoices", "1", "I am invoice 1")
					db.Set("invoices", "2", "I am invoice 2")
					db.Set("invoices", "3", "I am invoice 3")

					// get a user, an order and an invoice
					u, found := db.Get("users", "3")
					if !found {
						b.Fatalf("could not found user")
					}
					o, found := db.Get("orders", "7")
					if !found {
						b.Fatalf("could not found order")
					}
					inv, found := db.Get("invoices", "1")
					if !found {
						b.Fatalf("could not found invoice")
					}
					if u == "" || o == "" || inv == "" {
						b.Fatalf("returned values are empty")
					}
				}
			},
		},
		{
			"db2",
			func(b *testing.B) {

				db := NewDB2()

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					db.MakeRS("users")
					db.MakeRS("orders")
					db.MakeRS("invoices")

					// add some users
					db.Set("users", "1", "I am user 1")
					db.Set("users", "2", "I am user 2")
					db.Set("users", "3", "I am user 3")
					db.Set("users", "4", "I am user 4")
					db.Set("users", "5", "I am user 5")

					// add some orders
					db.Set("orders", "1", "I am order 1")
					db.Set("orders", "2", "I am order 2")
					db.Set("orders", "3", "I am order 3")
					db.Set("orders", "4", "I am order 4")
					db.Set("orders", "5", "I am order 5")
					db.Set("orders", "6", "I am order 6")
					db.Set("orders", "7", "I am order 7")
					db.Set("orders", "8", "I am order 8")
					db.Set("orders", "9", "I am order 9")
					db.Set("orders", "10", "I am order 10")

					// add some invoices
					db.Set("invoices", "1", "I am invoice 1")
					db.Set("invoices", "2", "I am invoice 2")
					db.Set("invoices", "3", "I am invoice 3")

					// get a user, an order and an invoice
					u, found := db.Get("users", "3")
					if !found {
						b.Fatalf("could not found user")
					}
					o, found := db.Get("orders", "7")
					if !found {
						b.Fatalf("could not found order")
					}
					inv, found := db.Get("invoices", "1")
					if !found {
						b.Fatalf("could not found invoice")
					}
					if u == "" || o == "" || inv == "" {
						b.Fatalf("returned values are empty")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, tt.fn)
	}
}

func TestDB1(t *testing.T) {
	db1 := NewDB1()
	db1.MakeRS("users")
	db1.MakeRS("orders")
	db1.MakeRS("invoices")

	// add some users
	db1.Set("users", "1", "I am user 1")
	db1.Set("users", "2", "I am user 2")
	db1.Set("users", "3", "I am user 3")
	db1.Set("users", "4", "I am user 4")
	db1.Set("users", "5", "I am user 5")

	// add some orders
	db1.Set("orders", "1", "I am order 1")
	db1.Set("orders", "2", "I am order 2")
	db1.Set("orders", "3", "I am order 3")
	db1.Set("orders", "4", "I am order 4")
	db1.Set("orders", "5", "I am order 5")
	db1.Set("orders", "6", "I am order 6")
	db1.Set("orders", "7", "I am order 7")
	db1.Set("orders", "8", "I am order 8")
	db1.Set("orders", "9", "I am order 9")
	db1.Set("orders", "10", "I am order 10")

	// add some invoices
	db1.Set("invoices", "1", "I am invoice 1")
	db1.Set("invoices", "2", "I am invoice 2")
	db1.Set("invoices", "3", "I am invoice 3")

	// get a user, an order and an invoice
	u, found := db1.Get("users", "3")
	if !found {
		t.Fatalf("could not found user")
	}
	o, found := db1.Get("orders", "7")
	if !found {
		t.Fatalf("could not found order")
	}
	i, found := db1.Get("invoices", "1")
	if !found {
		t.Fatalf("could not found invoice")
	}
	fmt.Println("user", u)
	fmt.Println("order", o)
	fmt.Println("invoice", i)
}

type DB2 struct {
	recordSets []RecordSet[any]
}

func NewDB2() *DB2 {
	return &DB2{
		recordSets: make([]RecordSet[any], 0),
	}
}

func (db2 *DB2) MakeRS(name string) {
	for _, set := range db2.recordSets {
		if set.Name == name {
			return
		}
	}
	set := NewRecordSet[any](name)
	db2.recordSets = append(db2.recordSets, *set)
}

func (db2 *DB2) Set(rs string, key string, val any) {
	var at int = -1
	for i, set := range db2.recordSets {
		if set.Name == rs {
			at = i
			break
		}
	}
	if at == -1 {
		return
	}
	db2.recordSets[at].AddRecord(key, val)
}

func (db2 *DB2) Get(rs string, key string) (any, bool) {
	var at int = -1
	for i, set := range db2.recordSets {
		if set.Name == rs {
			at = i
			break
		}
	}
	if at == -1 {
		return nil, false
	}
	return db2.recordSets[at].GetRecord(key).(any), true
}

func TestDB2(t *testing.T) {
	db2 := NewDB2()
	db2.MakeRS("users")
	db2.MakeRS("orders")
	db2.MakeRS("invoices")

	// add some users
	db2.Set("users", "1", "I am user 1")
	db2.Set("users", "2", "I am user 2")
	db2.Set("users", "3", "I am user 3")
	db2.Set("users", "4", "I am user 4")
	db2.Set("users", "5", "I am user 5")

	// add some orders
	db2.Set("orders", "1", "I am order 1")
	db2.Set("orders", "2", "I am order 2")
	db2.Set("orders", "3", "I am order 3")
	db2.Set("orders", "4", "I am order 4")
	db2.Set("orders", "5", "I am order 5")
	db2.Set("orders", "6", "I am order 6")
	db2.Set("orders", "7", "I am order 7")
	db2.Set("orders", "8", "I am order 8")
	db2.Set("orders", "9", "I am order 9")
	db2.Set("orders", "10", "I am order 10")

	// add some invoices
	db2.Set("invoices", "1", "I am invoice 1")
	db2.Set("invoices", "2", "I am invoice 2")
	db2.Set("invoices", "3", "I am invoice 3")

	// get a user, an order and an invoice
	u, found := db2.Get("users", "3")
	if !found {
		t.Fatalf("could not found user")
	}
	o, found := db2.Get("orders", "7")
	if !found {
		t.Fatalf("could not found order")
	}
	i, found := db2.Get("invoices", "1")
	if !found {
		t.Fatalf("could not found invoice")
	}
	fmt.Println("user", u)
	fmt.Println("order", o)
	fmt.Println("invoice", i)
}

type keyStr string
type keyStc struct {
	k1 string
	k2 string
}
type keyArr [2]string
type keyInf interface {
	Kind() int
	Key() any
}
type k1 string

func (k k1) Kind() int { return 1 }
func (k k1) Key() any  { return k }

type k2 [2]string

func (k k2) Kind() int { return 2 }
func (k k2) Key() any  { return k }

func BenchmarkMaps(b *testing.B) {
	tests := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{
			name: "regular map",
			fn: func(b *testing.B) {

				m := make(map[keyStr]int, 0)

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					for j := 0; j < 1000; j++ {
						k := keyStr(strconv.Itoa(j))
						m[k] = j
						_, found := m[k]
						if !found {
							b.Errorf("could not find key: %v\n", k)
						}
					}

				}

			},
		},
		{
			name: "map with compound keys (struct)",
			fn: func(b *testing.B) {

				m := make(map[keyStc]int, 0)

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					for j := 0; j < 1000; j++ {
						k := keyStc{
							strconv.Itoa(j),
							strconv.Itoa(j + 1),
						}
						m[k] = j
						_, found := m[k]
						if !found {
							b.Errorf("could not find key: %v\n", k)
						}
					}

				}
			},
		},
		{
			name: "map with compound keys (array)",
			fn: func(b *testing.B) {

				m := make(map[keyArr]int, 0)

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					for j := 0; j < 1000; j++ {
						k := keyArr{
							strconv.Itoa(j),
							strconv.Itoa(j + 1),
						}
						m[k] = j
						_, found := m[k]
						if !found {
							b.Errorf("could not find key: %v\n", k)
						}
					}
				}
			},
		},
		{
			name: "map with compound keys (interface)",
			fn: func(b *testing.B) {

				m := make(map[keyInf]int, 0)

				b.ResetTimer()
				b.ReportAllocs()

				var k keyInf
				for i := 0; i < b.N; i++ {
					for j := 0; j < 1000; j++ {
						if j^1 == j+1 {
							// even
							k = k1(strconv.Itoa(j))
						} else {
							// odd
							k = k2{
								strconv.Itoa(j),
								strconv.Itoa(j + 1),
							}
						}
						m[k] = j
						_, found := m[k]
						if !found {
							b.Errorf("could not find key: %v\n", k)
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
