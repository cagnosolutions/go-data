package table

import (
	"fmt"
	"testing"
)

type Person struct {
	Name string
	Age  int
}

func TestStructToMap(t *testing.T) {
	m := StructToMap(&Person{Name: "Darth Vader", Age: 56})
	if m == nil {
		t.Error("got nil map")
	}
	fmt.Printf("%#v\n", m)
}

func TestMapToStruct(t *testing.T) {
	m := map[string]interface{}{"Age": 56, "Name": "Darth Vader"}
	var p Person
	MapToStruct(m, &p)
	if p == *new(Person) || &p == nil {
		t.Error("got empty struct")
	}
	fmt.Printf("%#v\n", p)
}
