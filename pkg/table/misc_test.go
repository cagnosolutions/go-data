package table

import (
	"fmt"
	"testing"
)

type Person struct {
	Name   string
	foo    int
	Age    int
	Active bool
}

func TestStructToMap(t *testing.T) {
	m := StructToMap(&Person{Name: "Darth Vader", Age: 56, Active: true})
	if m == nil {
		t.Error("got nil map")
	}
	fmt.Printf("%#v\n", m)
}

func TestMapToStruct(t *testing.T) {
	m := map[string]interface{}{"Age": 56, "Name": "Darth Vader", "Active": true}
	var p Person
	MapToStruct(m, &p)
	if p == *new(Person) || &p == nil {
		t.Error("got empty struct")
	}
	fmt.Printf("%#v\n", p)
}

func TestStructToTable(t *testing.T) {
	tbl := NewTable()
	tbl.FillTable(&Person{Name: "Darth Vader", Age: 56, Active: true})
	if len(tbl.Name) == 0 {
		t.Error("empty table")
	}
	fmt.Printf("%#v\n", tbl)
}

func BenchmarkStructToMap(b *testing.B) {
	b.ReportAllocs()
	p := &Person{Name: "Darth Vader", Age: 56, Active: true}
	for i := 0; i < b.N; i++ {
		m := StructToMap(p)
		if m == nil {
			b.Error("got nil map")
		}
	}
}

func BenchmarkMapToStruct(b *testing.B) {
	b.ReportAllocs()
	m := map[string]interface{}{"Name": "Darth Vader", "Age": 56, "Active": true}
	for i := 0; i < b.N; i++ {
		var p Person
		MapToStruct(m, &p)
		if p == *new(Person) || &p == nil {
			b.Error("got empty struct")
		}
	}
}

func BenchmarkStructToTable(b *testing.B) {
	b.ReportAllocs()
	tbl := NewTable()
	for i := 0; i < b.N; i++ {
		tbl.FillTable(&Person{Name: "Darth Vader", Age: 56, Active: true})
		if len(tbl.Name) == 0 {
			b.Error("empty table")
		}
	}
}
