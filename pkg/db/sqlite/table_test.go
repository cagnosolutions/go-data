package sqlite

import (
	"fmt"
	"testing"
	"time"
)

func AssertExpected(t *testing.T, expected, got interface{}) {
	if expected != got {
		t.Errorf("expected=%v, got=%v\n", expected, got)
	}
}

func TestTable_NewTable(t *testing.T) {
	var tb *Table
	AssertExpected(t, (*Table)(nil), tb)
	tb = NewTable(
		&struct {
			ID         int
			Name       string
			NetWorth   float64
			IsMarried  bool
			Occupation string
		}{
			ID:         12345,
			Name:       "Money Bags",
			NetWorth:   789456.69,
			IsMarried:  true,
			Occupation: "Banker",
		},
	)
	AssertExpected(t, true, tb != nil)
	AssertExpected(t, 5, len(tb.Fields))
	fmt.Printf("%#v\n", tb)
}

func BenchmarkTable_NewTable(b *testing.B) {
	b.ReportAllocs()
	var tb *Table
	for i := 0; i < b.N; i++ {
		tb = NewTable(
			&struct {
				ID         int
				Name       string
				NetWorth   float64
				IsMarried  bool
				Occupation string
			}{
				ID:         12345,
				Name:       "Money Bags",
				NetWorth:   789456.69,
				IsMarried:  true,
				Occupation: "Banker",
			},
		)
	}
	res = tb
}

type UserTable1 struct {
	ID           int       `db:"id"`
	FName        string    `db:"f_name"`
	LName        string    `db:"l_name"`
	FullName     string    `db:"full_name"`
	EmailAddress string    `db:"email_address"`
	Age          int       `db:"age"`
	RegisteredOn time.Time `db:"registered_on"`
	IsActive     bool      `db:"is_active"`
}

func TestTable_Create(t *testing.T) {
	var tb *Table
	AssertExpected(t, (*Table)(nil), tb)
	tb = NewTable(&UserTable1{})
	AssertExpected(t, true, tb != nil)
	var stmt string
	stmt = tb.Create()
	if stmt == "" {
		t.Error("bad statement")
	}
	fmt.Println(stmt)
}

func TestTable_Drop(t *testing.T) {
	var tb *Table
	AssertExpected(t, (*Table)(nil), tb)
	tb = NewTable(&UserTable1{})
	AssertExpected(t, true, tb != nil)
	var stmt string
	stmt = tb.Drop()
	if stmt == "" {
		t.Error("bad statement")
	}
	fmt.Println(stmt)
}

func TestTable_Select(t *testing.T) {
	var tb *Table
	AssertExpected(t, (*Table)(nil), tb)
	tb = NewTable(&UserTable1{})
	AssertExpected(t, true, tb != nil)
	var stmt string
	stmt = tb.Select(
		"is_active=$active and age < $age", map[string]interface{}{
			"active": true,
			"age":    3,
		},
	)
	if stmt == "" {
		t.Error("bad statement")
	}
	fmt.Println(stmt)
}

func TestWithTimeOut(t *testing.T) {
	timeout := time.After(10 * time.Second)
	done := make(chan bool)
	go func() {
		t.Run("TestTable_Create", TestTable_Create)
		time.Sleep(5 * time.Second)
		done <- true
	}()
	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}
}

func BenchmarkWithTimeOut(b *testing.B) {
	timeout := time.After(10 * time.Second)
	done := make(chan bool)
	go func() {
		b.Run("BenchmarkTable_Create", BenchmarkTable_Create)
		time.Sleep(5 * time.Second)
		done <- true
	}()
	select {
	case <-timeout:
		b.Fatal("Test didn't finish in time")
	case <-done:
	}
}

var res interface{}

func BenchmarkTable_Create(b *testing.B) {
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
	tb := NewTable(user)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stmt = tb.Create()
		if stmt == "" {
			b.Error("bad statement")
		}
	}
	res = stmt
}
