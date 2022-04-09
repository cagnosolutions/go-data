package sqlite

import (
	"log"
	"reflect"
	"strings"
	"unsafe"

	"github.com/cagnosolutions/go-data/pkg/format"
)

type Field struct {
	Name  string
	Value interface{}
	Type  byte
}

type Table struct {
	Name   string
	Fields []Field
	Size   int
}

// NewTable fills out a table using a *struct as input
func NewTable(ptr interface{}) *Table {
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
	t := &Table{
		Name:   format.ToSnakeCase(typ.Name()),
		Fields: make([]Field, 0, val.NumField()),
	}
	t.Size = len(t.Name) + len(t.Fields)
	t.Name = format.ToSnakeCase(typ.Name())
	t.Fields = make([]Field, 0, val.NumField())
	t.Size = len(t.Name)
	// loop over each field in the struct
	for i := 0; i < val.NumField(); i++ {
		// return an instance of struct field at this position
		sf := typ.Field(i)
		// check to make sure the field is exported
		if !sf.IsExported() {
			// if it's not, skip it
			continue
		}
		// get the name
		name := sf.Tag.Get("db")
		if name == "" {
			name = format.ToSnakeCase(sf.Name)
		}
		// otherwise, fill out the table field
		t.Fields = append(
			t.Fields, Field{
				Name:  name,
				Value: val.Field(i).Interface(),
				Type:  typeOf(sf.Type.Kind()),
			},
		)
	}
	// update size
	for i := range t.Fields {
		t.Size += len(t.Fields[i].Name) + int(unsafe.Sizeof(t.Fields[i].Value))
	}
	// return table
	return t
}

func (t *Table) Create() string {
	var sb strings.Builder
	var hasPK bool
	sb.Grow(t.Size + 120)
	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(t.Name)
	sb.WriteString(" (\n")
	for i := range t.Fields {
		sb.WriteString("\t")
		sb.WriteString(t.Fields[i].Name)
		sb.WriteString(" ")
		sb.WriteString(t.Fields[i].Kind())
		if t.Fields[i].Name == "id" && !hasPK {
			sb.WriteString(" NOT NULL PRIMARY KEY,\n")
			hasPK = true
			continue
		}
		if i < len(t.Fields)-1 {
			sb.WriteString(" NOT NULL,\n")
			continue
		}
		sb.WriteString(" NOT NULL\n")
	}
	sb.WriteString(");")
	return sb.String()
}

func (t *Table) Drop() string {
	var sb strings.Builder
	sb.Grow(len(t.Name) + 23)
	sb.WriteString("DROP TABLE IF EXISTS ")
	sb.WriteString(t.Name)
	sb.WriteString(";")
	return sb.String()
}

func (t *Table) Select(where string, data map[string]interface{}) string {
	var sb strings.Builder
	var hasWhere bool
	var whereStr string
	if len(where) > 0 && data != nil {
		hasWhere = true
		whereStr = format.Mprintf(where, data)
	}
	sb.Grow(t.Size + 32 + len(whereStr))
	sb.WriteString("SELECT * FROM ")
	sb.WriteString(t.Name)
	if !hasWhere {
		sb.WriteString(";")
		return sb.String()
	}
	sb.WriteString(" WHERE ")
	sb.WriteString(whereStr)
	sb.WriteString(";")
	return sb.String()
}

func (t *Table) Insert(args ...string) string {
	return ""
}

func (t *Table) Update(args ...string) string {
	return ""
}

func (t *Table) Delete(args ...string) string {
	return ""
}

func (f *Field) Kind() string {
	switch f.Type {
	case INTEGER:
		return "INTEGER"
	case TEXT:
		return "TEXT"
	case BLOB:
		return "BLOB"
	case REAL:
		return "REAL"
	case NUMERIC:
		return "NUMERIC"
	case NULL:
		return "NULL"
	default:
		return "INVALID"
	}
}
