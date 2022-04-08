package sqlfmt

import (
	"fmt"
	"testing"
	"time"

	"github.com/cagnosolutions/go-data/pkg/table"
)

type UserTable1 struct {
	ID           int
	FName        string
	LName        string
	FullName     string
	EmailAddress string
	Age          int
	RegisteredOn time.Time
	IsActive     bool
}

func TestSQLBuilder_CreateStmt(t *testing.T) {
	var sqlb *SQLBuilder
	var stmt string
	// user := &UserTable1{
	// 	ID:           23,
	// 	FName:        "Jon",
	// 	LName:        "Doe",
	// 	FullName:     "Jon Doe",
	// 	EmailAddress: "jdoe@example.com",
	// 	Age:          48,
	// 	RegisteredOn: time.Now(),
	// 	IsActive:     true,
	// }

	stmt = sqlb.CreateStmt(table.MakeTable(&UserTable1{}))
	if stmt == "" {
		t.Error("bad statement")
	}
	fmt.Println(stmt)
}

var res interface{}

func BenchmarkSQLBuilder_CreateStmt(b *testing.B) {
	var sqlb *SQLBuilder
	var stmt string
	user := &UserTable1{
		ID:           23,
		FName:        "Jon",
		LName:        "Doe",
		FullName:     "Jon Doe",
		EmailAddress: "jdoe@example.com",
		Age:          48,
		RegisteredOn: time.Now(),
		IsActive:     true,
	}
	tbl := table.MakeTable(user)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stmt = sqlb.CreateStmt(tbl)
		if stmt == "" {
			b.Error("bad statement")
		}
	}
	res = stmt
}
