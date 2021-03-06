package table

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"unsafe"
)

// Table represents a generic table structure
type Table struct {
	Name   string
	Fields []Field
}

func (t *Table) Size() int {
	if t.Name == "" || t.Fields == nil || len(t.Fields) < 1 {
		return 0
	}
	var size int
	size += len(t.Name)
	size += len(t.Fields)
	for i := range t.Fields {
		size += len(t.Fields[i].Name)
		size += int(unsafe.Sizeof(t.Fields[i].Value))
	}
	return size
}

// NewTable creates and returns a new empty *Table
func NewTable() *Table {
	return new(Table)
}

// String is the stringer method for a table
func (t *Table) String() string {
	var ss []string
	ss = append(ss, fmt.Sprintf("Name: %q", t.Name))
	for _, fld := range t.Fields {
		ss = append(ss, fmt.Sprintf("%v", fld))
	}
	return strings.Join(ss, ", ")
}

// Fill fills out a table using a *struct as input
func (t *Table) Fill(ptr interface{}) {
	// get the value of the struct
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		// dereference the pointer
		val = val.Elem()
	}
	// check the type to ensure it is actually a struct
	if val.Kind() != reflect.Struct {
		log.Panicf("%v type can't have attributes inspected\n", val.Kind())
	}
	// get the underlying type from the reflected value
	typ := val.Type()
	// fill out the table name, and instantiate the fields
	t.Name = typ.Name()
	t.Fields = make([]Field, 0, val.NumField())
	// loop over each field in the struct
	for i := 0; i < val.NumField(); i++ {
		// return an instance of struct field at this position
		sf := typ.Field(i)
		// check to make sure the field is exported
		if !sf.IsExported() {
			// if it's not, skip it
			continue
		}
		// otherwise, fill out the table field
		t.Fields = append(
			t.Fields, Field{
				Name:  sf.Name,
				Kind:  sf.Type.Kind(),
				Value: val.Field(i).Interface(),
			},
		)
	}
}

// MakeTable creates a new table, fills it out using the provided
// struct pointer and returns the newly created *Table.
func MakeTable(ptr interface{}) *Table {
	t := NewTable()
	t.Fill(ptr)
	return t
}
