package sqlite

import (
	"reflect"
)

const (
	INVALID byte = iota
	INTEGER
	TEXT
	BLOB
	REAL
	NUMERIC
	NULL
)

func typeOf(t reflect.Kind) byte {
	switch t {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return INTEGER
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return INTEGER
	case reflect.Float32, reflect.Float64:
		return REAL
	case reflect.Bool:
		return NUMERIC
	case reflect.Map, reflect.Slice, reflect.Struct:
		return BLOB
	case reflect.String:
		return TEXT
	// case time:
	//	return NUMERIC
	case reflect.Invalid:
		return NULL
	default:
		return INVALID
	}
}
