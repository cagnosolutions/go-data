package table

import (
	"fmt"
	"time"
)

func Mapper(args ...interface{}) *TableLite {
	if args == nil || len(args)%2 != 0 {
		return nil
	}
	s := &TableLite{
		Fields: make([]Entry, len(args)),
	}
	for i := 0; i < len(args); i += 2 {
		e := Entry{
			Key: fmt.Sprintf("%v", args[i]),
			Val: args[i+1],
			Typ: typeOf(args[i+1]),
		}
		s.Fields[i] = e
	}
	return s
}

const (
	INVALID byte = iota
	INTEGER
	TEXT
	BLOB
	REAL
	NUMERIC
	NULL
)

func typeOf(v interface{}) byte {
	switch v.(type) {
	case int, int8, int16, int32, int64:
		return INTEGER
	case uint, uint8, uint16, uint32, uint64:
		return INTEGER
	case float32, float64:
		return REAL
	case bool:
		return NUMERIC
	case []byte:
		return BLOB
	case string:
		return TEXT
	case time.Time:
		return NUMERIC
	case nil:
		return NULL
	default:
		return INVALID
	}
}

type Entry struct {
	Key string
	Val interface{}
	Typ byte
}

type TableLite struct {
	Name   string
	Fields []Entry
}
