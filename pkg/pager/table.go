package pager

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cagnosolutions/go-data/pkg/format"
)

type colType byte

const (
	kindNIL colType = iota
	kindString
	kindBinary
	kindInt
	kindUint
	kindFloat
	kindBool
	kindBit
	kindJSON
	kindTime
)

func getColType(t reflect.Kind) colType {
	switch t {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return kindInt
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return kindUint
	case reflect.Float32, reflect.Float64:
		return kindFloat
	case reflect.Bool:
		return kindBool
	case reflect.Map, reflect.Array, reflect.Struct:
		return kindJSON
	case reflect.String:
		return kindString
	case reflect.TypeOf([]byte{}).Kind():
		return kindBinary
	case reflect.TypeOf(byte(0)).Kind():
		return kindBit
	case reflect.TypeOf(time.Time{}).Kind():
		return kindTime
	default:
		return kindNIL
	}
}

type R = row
type C = column

type row struct {
	id      int
	columns []column
}

type column struct {
	name  string
	kind  colType
	value any
}

const (
	minWidth     = 32
	tabWidth     = 4
	padWidth     = 0
	padChar      = ' '
	padCharDebug = '.'
)

func (r *row) rowHeader() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, minWidth, tabWidth, padWidth, padChar, tabwriter.DiscardEmptyColumns)
	var err error
	var bound string
	_, err = fmt.Fprintf(w, "%s\t", "id")
	if err != nil {
		panic(err)
	}
	for _, col := range r.columns {
		_, err = fmt.Fprintf(w, "%s\t", col.name)
		if err != nil {
			panic(err)
		}
	}
	_, err = fmt.Fprintf(w, "\n")
	if err != nil {
		panic(err)
	}
	bound += "\n"
	err = w.Flush()
	if err != nil {
		panic(err)
	}
	ss := buf.String()
	sb := strings.Repeat("-", len(ss))
	return sb + "\n" + ss + sb + "\n"
}

func (r *row) rowString() string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, minWidth, tabWidth, padWidth, padChar, 0)
	_, err := fmt.Fprintf(w, "%d\t", r.id)
	if err != nil {
		panic(err)
	}
	for _, col := range r.columns {
		var format string
		switch col.kind {
		case kindString, kindBinary:
			format = "%q\t"
		case kindInt, kindUint:
			format = "%d\t"
		case kindFloat:
			format = "%f\t"
		default:
			format = "%v\t"
		}
		_, err = fmt.Fprintf(w, format, col.value)
		if err != nil {
			panic(err)
		}
	}
	_, err = fmt.Fprintf(w, "\n")
	if err != nil {
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}
	return sb.String() + strings.Repeat("-", sb.Len()) + "\n"
}

type table struct {
	name   string
	header row
	rows   []row
	count  int
}

func makeTable(name string, hcol ...column) *table {
	if hcol == nil || len(hcol) < 1 {
		return nil
	}
	return &table{
		name: name,
		header: row{
			columns: hcol,
		},
		rows:  make([]row, 0),
		count: 0,
	}
}

func (t *table) insertStruct(ptr any) {
	t.rows = append(
		t.rows, row{
			id:      t.count + 1,
			columns: structToCols(ptr),
		},
	)
	t.count++
}

func (t *table) String() string {
	ss := t.header.rowHeader()
	for _, r := range t.rows {
		ss += r.rowString()
	}
	return ss
}

// structToCols fills out a table using a *struct as input
func structToCols(ptr interface{}) []column {
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
	cols := make([]column, 0, val.NumField())
	// loop over each field in the struct
	for i := 0; i < val.NumField(); i++ {
		// return an instance of struct field at this position
		sf := typ.Field(i)
		// check to make sure the field is exported
		if !sf.IsExported() {
			// if it's not, skip it
			continue
		}
		cols = append(
			cols, column{
				name:  format.ToSnakeCase(sf.Name),
				kind:  getColType(sf.Type.Kind()),
				value: val.Field(i).Interface(),
			},
		)
	}
	// return columns
	return cols
}
