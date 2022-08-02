package util

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type TableWriter struct {
	*tabwriter.Writer
}

func NewTableWriter(hdr ...string) *TableWriter {
	w := tabwriter.NewWriter(os.Stdout, 8, 4, 8, '\t', tabwriter.AlignRight)
	fmt.Fprintf(w, strings.Join(hdr, "\t"))
	fmt.Fprintf(w, "\t\n")
	return &TableWriter{w}
}

func (tw *TableWriter) WriteRow(row ...any) {
	for i := range row {
		fmt.Fprintf(tw, fmt.Sprintf("%v\t", row[i]))
	}
	fmt.Fprintf(tw, "\n")
}

func (tw *TableWriter) Flush() {
	tw.Writer.Flush()
}
