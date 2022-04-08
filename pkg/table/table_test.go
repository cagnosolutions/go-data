package table

import (
	"fmt"
	"testing"
)

func AssertExpected(t *testing.T, expected, got interface{}) {
	if expected != got {
		t.Errorf("expected=%v, got=%v\n", expected, got)
	}
}

func TestNewTable(t *testing.T) {
	var table *Table
	AssertExpected(t, (*Table)(nil), table)
	table = NewTable()
	AssertExpected(t, table != nil, table != nil)
}

func TestTable_FillTable(t *testing.T) {
	var table *Table
	AssertExpected(t, (*Table)(nil), table)
	table = NewTable()
	AssertExpected(t, true, table != nil)
	table.Fill(
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
	AssertExpected(t, 5, len(table.Fields))
}

func TestTable_MakeTable(t *testing.T) {
	var table *Table
	AssertExpected(t, (*Table)(nil), table)
	table = MakeTable(
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
	AssertExpected(t, true, table != nil)
	AssertExpected(t, 5, len(table.Fields))
}

func TestTable_SQLiteCreateTable(t *testing.T) {
	var table *Table
	AssertExpected(t, (*Table)(nil), table)
	table = MakeTable(
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
	AssertExpected(t, true, table != nil)
	AssertExpected(t, 5, len(table.Fields))

}

func BenchmarkStructToTable(b *testing.B) {
	b.ReportAllocs()
	tbl := NewTable()
	for i := 0; i < b.N; i++ {
		tbl.Fill(&Person{Name: "Darth Vader", Age: 56, IsActive: true})
		if len(tbl.Name) == 0 {
			b.Error("empty table")
		}
	}
}

type Moneybags struct {
	ID         int
	Name       string
	NetWorth   float64
	IsMarried  bool
	Occupation string
}

func BenchmarkTable_Size(b *testing.B) {
	b.ReportAllocs()

	tbl := MakeTable(
		&Moneybags{
			ID:         12345,
			Name:       "Money Bags",
			NetWorth:   789456.69,
			IsMarried:  true,
			Occupation: "Banker",
		},
	)
	fmt.Printf("%s\n", tbl)
	var size int
	for i := 0; i < b.N; i++ {
		size = tbl.Size()
		if size < 1 {
			b.Error("bad table size")
		}
	}
	res = size
	fmt.Println(res)
	grow1 := 32 + (32 * len(tbl.Fields)) + tbl.Size()
	fmt.Println("grow1:", grow1)
	grow2 := tbl.Size() + 125
	fmt.Println("grow2:", grow2)

	ctsize := `CREATE TABLE IF NOT EXISTS money_bags (
	id INTEGER NOT NULL PRIMARY KEY,
	name TEXT NOT NULL DEFAULT '',
	net_worth NUMERIC NOT NULL DEFAULT 0,
	is_married NUMERIC NOT NULL DEFAULT 0,
	occupation TEXT NOT NULL DEFAULT ''
);`
	fmt.Println(ctsize, len(ctsize))
}
