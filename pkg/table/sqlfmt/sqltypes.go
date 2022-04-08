package sqlfmt

import (
	"reflect"
)

var sqlTypeMap = map[int]string{
	sql_NONE:       "NONE",
	sql_INVALID:    "INVALID",
	sql_NULL:       "NULL",
	sql_INTEGER:    "INTEGER",
	sql_INT:        "INT",
	sql_TINYINT:    "TINYINT",
	sql_SMALLINT:   "SMALLINT",
	sql_MEDIUMINT:  "MEDIUMINT",
	sql_BIGINT:     "BIGINT",
	sql_INT2:       "INT2",
	sql_INT8:       "INT8",
	sql_BIT:        "BIT",
	sql_TEXT:       "TEXT",
	sql_CHAR:       "CHAR",
	sql_CHARACTER:  "CHARACTER",
	sql_VARCHAR:    "VARCHAR",
	sql_NCHAR:      "NCHAR",
	sql_NVARCHAR:   "NVARCHAR",
	sql_CLOB:       "CLOB",
	sql_BLOB:       "BLOB",
	sql_TINYBLOB:   "TINYBLOB",
	sql_TINYTEXT:   "TINYTEXT",
	sql_MEDIUMBLOB: "MEDIUMBLOB",
	sql_MEDIUMTEXT: "MEDIUMTEXT",
	sql_LONGBLOB:   "LONGBLOB",
	sql_LONGTEXT:   "LONGTEXT",
	sql_DATE:       "DATE",
	sql_DATETIME:   "DATETIME",
	sql_REAL:       "REAL",
	sql_DOUBLE:     "DOUBLE",
	sql_FLOAT:      "FLOAT",
	sql_NUMERIC:    "NUMERIC",
	sql_DECIMAL:    "DECIMAL",
	sql_BOOLEAN:    "BOOLEAN",
}

const (
	// type affinity sql_NONE
	sql_NONE = iota
	sql_INVALID
	sql_NULL

	// type affinity sql_INTEGER
	sql_INTEGER
	sql_INT
	sql_TINYINT
	sql_SMALLINT
	sql_MEDIUMINT
	sql_BIGINT
	sql_INT2
	sql_INT8
	sql_BIT

	// type affinity sql_TEXT
	sql_TEXT
	sql_CHAR
	sql_CHARACTER
	sql_VARCHAR
	sql_NCHAR
	sql_NVARCHAR
	sql_CLOB
	sql_BLOB
	sql_TINYBLOB
	sql_TINYTEXT
	sql_MEDIUMBLOB
	sql_MEDIUMTEXT
	sql_LONGBLOB
	sql_LONGTEXT
	sql_DATE
	sql_DATETIME

	// type affinity sql_REAL
	sql_REAL
	sql_DOUBLE
	sql_FLOAT

	// type affinity sql_NUMERIC
	sql_NUMERIC
	sql_DECIMAL
	sql_BOOLEAN
)

func sqlType(typ reflect.Kind) int {
	switch typ {
	case reflect.Invalid:
		return sql_NONE
	case
		reflect.Int,
		reflect.Uint:
		return sql_INT
	case
		reflect.Int8,
		reflect.Uint8:
		return sql_TINYINT
	case
		reflect.Int16,
		reflect.Uint16:
		return sql_SMALLINT
	case
		reflect.Int32,
		reflect.Uint32:
		return sql_MEDIUMINT
	case
		reflect.Int64,
		reflect.Uint64:
		return sql_BIGINT
	case
		reflect.Bool:
		return sql_NUMERIC
	case
		reflect.Float32:
		return sql_FLOAT
	case
		reflect.Float64:
		return sql_DOUBLE
	case
		reflect.Array,
		reflect.Map,
		reflect.Struct:
		return sql_BLOB
	case
		reflect.String:
		return sql_TEXT
	default:
		return sql_NONE
	}
}

func GetSQLType(typ reflect.Kind) string {
	s, found := sqlTypeMap[sqlType(typ)]
	if !found {
		return ""
	}
	return s
}

func GetSQLiteType(typ reflect.Kind) string {
	s, found := sqlTypeMap[sqlTypeAffinity(sqlType(typ))]
	if !found {
		return ""
	}
	return s
}

func sqlTypeAffinity(sqlTyp int) int {
	switch sqlTyp {
	case
		sql_NONE,
		sql_INVALID,
		sql_NULL:
		return sql_NONE
	case
		sql_INTEGER,
		sql_INT,
		sql_TINYINT,
		sql_SMALLINT,
		sql_MEDIUMINT,
		sql_BIGINT,
		sql_INT2,
		sql_INT8,
		sql_BIT:
		return sql_INTEGER
	case
		sql_TEXT,
		sql_CHAR,
		sql_CHARACTER,
		sql_VARCHAR,
		sql_NCHAR,
		sql_NVARCHAR,
		sql_CLOB,
		sql_BLOB,
		sql_TINYBLOB,
		sql_TINYTEXT,
		sql_MEDIUMBLOB,
		sql_MEDIUMTEXT,
		sql_LONGBLOB,
		sql_LONGTEXT,
		sql_DATE,
		sql_DATETIME:
		return sql_TEXT
	case
		sql_REAL,
		sql_DOUBLE,
		sql_FLOAT:
		return sql_REAL
	case
		sql_NUMERIC,
		sql_DECIMAL,
		sql_BOOLEAN:
		return sql_NUMERIC
	default:
		return sql_NONE
	}
}
