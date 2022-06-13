package pager

import (
	"fmt"
	"os"
	"testing"
	"text/tabwriter"
)

type User struct {
	Name   string
	Email  string
	Age    int
	Active bool
}

func TestTable_MakeTable(t *testing.T) {
	tbl := makeTable(
		"users",
		C{"name", kindString, []byte("Jon Doe")},
		C{"email", kindString, []byte("jdoe@example.com")},
		C{"age", kindInt, []byte("35")},
		C{"active", kindBool, []byte("true")},
	)
	tbl.insertStruct(
		&User{
			Name:   "Jon Doe",
			Email:  "jdoe@example.com",
			Age:    35,
			Active: true,
		},
	)
	tbl.insertStruct(
		&User{
			Name:   "Jane Mar",
			Email:  "jmar@example.com",
			Age:    37,
			Active: true,
		},
	)
	fmt.Println(tbl)
}

func TestTabWriter(t *testing.T) {
	// initialize tabwriter
	w := new(tabwriter.Writer)

	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)

	defer w.Flush()

	fmt.Fprintf(w, "\b%s\t%s\t%s\t\n", "Col1", "Col2", "Col3")
	fmt.Fprintf(w, "\b%s\t%s\t%s\t\n", "----", "----", "----")

	for i := 0; i < 5; i++ {
		fmt.Fprintf(w, "\n %d\t%d\t%d\t\n", i, i+1, i+2)
	}

	//	Col1	Col2	Col3
	//	----	----	----
	//	0       1       2
	//	1       2       3
	//	2       3       4
	//	3       4       5
	//  4       5       6
}

func TestTabWriter2(t *testing.T) {
	// Observe how the b's and the d's, despite appearing in the
	// second cell of each line, belong to different columns.
	w := tabwriter.NewWriter(os.Stdout, 16, 8, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "|a\tb\tc\t")
	fmt.Fprintln(w, "|aa\tbb\tcc\t")
	fmt.Fprintln(w, "|aaa\t\t\t") // trailing tab
	fmt.Fprintln(w, "|aaaa\tdddd\teeee\t")
	w.Flush()
}
