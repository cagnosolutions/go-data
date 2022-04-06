package table

import (
	"fmt"
	"reflect"
)

// Field represents a field in a table
type Field struct {
	Name  string
	Kind  reflect.Kind
	Value interface{}
}

// ValueString returns the interface value as a string
func (f *Field) ValueString() string {
	switch f.Value.(type) {
	case string:
		return fmt.Sprintf("'%s'", f.Value)
	default:
		return fmt.Sprintf("%v", f.Value)
	}
}

// String is the stringer method for a field
func (f *Field) String() string {
	return fmt.Sprintf("{ %q: %s (%s) }", f.Name, f.ValueString(), f.Kind)
}
