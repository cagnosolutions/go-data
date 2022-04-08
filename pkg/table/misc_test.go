package table

import (
	"fmt"
	"testing"
)

type Person struct {
	Name     string
	foo      int
	Age      int
	IsActive bool
}

func TestStructToMap(t *testing.T) {
	m := StructToMap(&Person{Name: "Darth Vader", Age: 56, IsActive: true})
	if m == nil {
		t.Error("got nil map")
	}
	fmt.Printf("%#v\n", m)
}

func TestMapToStruct(t *testing.T) {
	m := map[string]interface{}{"Age": 56, "foo": 234, "Name": "Darth Vader", "Active": true}
	var p Person
	MapToStruct(m, &p)
	if p == *new(Person) || &p == nil {
		t.Error("got empty struct")
	}
	fmt.Printf("%#v\n", p)
}

func TestStructToTable(t *testing.T) {
	tbl := NewTable()
	tbl.Fill(&Person{Name: "Darth Vader", Age: 56, IsActive: true})
	if len(tbl.Name) == 0 {
		t.Error("empty table")
	}
	fmt.Printf("%#v\n", tbl)
}

var res interface{}

func BenchmarkStructToMap(b *testing.B) {
	b.ReportAllocs()
	p := &Person{Name: "Darth Vader", Age: 56, IsActive: true}
	var m map[string]interface{}
	for i := 0; i < b.N; i++ {
		m = StructToMap(p)
		if m == nil {
			b.Error("got nil map")
		}
	}
	res = m
}

func BenchmarkMapToStruct(b *testing.B) {
	b.ReportAllocs()
	m := map[string]interface{}{"Name": "Darth Vader", "Age": 56, "Active": true}
	var p Person
	for i := 0; i < b.N; i++ {
		MapToStruct(m, &p)
		if p == *new(Person) || &p == nil {
			b.Error("got empty struct")
		}
	}
	res = p
}

func TestSprintAny(t *testing.T) {
	var out string
	out = SprintAny(
		"My name is $name, I am $age and I am $is_active",
		&Person{Name: "Darth Vader", Age: 56, IsActive: true},
	)
	fmt.Println(out)
	res = out
}

func BenchmarkSprintAny(b *testing.B) {
	b.ReportAllocs()
	var out string
	for i := 0; i < b.N; i++ {
		out = SprintAny(
			"My name is $name, I am $age and I am $is_active",
			&Person{Name: "Darth Vader", Age: 56, IsActive: true},
		)
		if len(out) == 0 {
			b.Error("fail")
		}
	}
	res = out
}
