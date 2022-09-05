package page

import (
	"fmt"
	"testing"
)

func TestCell_newCell(t *testing.T) {
	type args struct {
		id     uint16
		flags  uint16
		offset uint16
		length uint16
	}
	var tests = []struct {
		name string
		args args
		want cell
	}{
		{
			name: "new-cell-1",
			args: args{id: 65535, offset: 65535, length: 65535},
			want: cell(18446744069414780927),
		},
		{
			name: "new-cell-2",
			args: args{id: 1, offset: 4096, length: 255},
			want: cell(71793711247392769),
		},
		{
			name: "new-cell-3",
			args: args{id: 256, flags: 0xffff, offset: 16384, length: 8192},
			want: cell(2305913382252773632),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := newCell(tt.args.id, tt.args.offset, tt.args.length)
				if tt.name == "new-cell-3" {
					got.setFlags(0xffff)
				}
				if got != tt.want {
					t.Errorf("newCell() = %v, want %v", got, tt.want)
				}
				// fmt.Printf("%s --> %q\n", tt.name, got.String())
			},
		)
	}
}

var setAndGetTests = struct {
	name string
	t    []struct {
		name string
		c    cell
		set  uint16
		want uint16
	}
}{
	t: []struct {
		name string
		c    cell
		set  uint16
		want uint16
	}{
		{name: "cell-0", c: cell(0), set: 16, want: 16},
		{name: "cell-1", c: cell(0), set: 1, want: 1},
		{name: "cell-2", c: cell(0), set: 0x0100, want: 256},
		{name: "cell-3", c: cell(0), set: 32, want: 32},
		{name: "cell-4", c: cell(0), set: 0xff00, want: 65280},
		{name: "cell-5", c: cell(0), set: 0xffff, want: 65535},
		{name: "cell-6", c: cell(0), set: 255, want: 255},
		{name: "cell-7", c: cell(0), set: 0, want: 0},
		{name: "cell-8", c: cell(0), set: 9999, want: 9999},
		{name: "cell-9", c: cell(0), set: 0x1234, want: 4660},
	},
}

func TestCell_ID(t *testing.T) {
	tests := setAndGetTests
	tests.name = "(set/get)ID"
	for _, tt := range tests.t {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.c.setID(tt.set)
				got := tt.c.getID()
				if got != tt.want {
					t.Errorf("%s() = %v, want %v", tests.name, got, tt.want)
				}
				// fmt.Printf("%q --> %q\n", tt.name, tt.c.String())
			},
		)
	}
}

func TestCell_Flags(t *testing.T) {
	tests := setAndGetTests
	tests.name = "(set/get)Flags"
	for _, tt := range tests.t {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.c.setFlags(tt.set)
				got := tt.c.getFlags()
				if got != tt.want {
					t.Errorf("%s() = %v, want %v", tests.name, got, tt.want)
				}
				// fmt.Printf("%q --> %q\n", tt.name, tt.c.String())
			},
		)
	}
}

func TestCell_Offset(t *testing.T) {
	tests := setAndGetTests
	tests.name = "(set/get)Offset"
	for _, tt := range tests.t {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.c.setOffset(tt.set)
				got := tt.c.getOffset()
				if got != tt.want {
					t.Errorf("%s() = %v, want %v", tests.name, got, tt.want)
				}
				// fmt.Printf("%q --> %q\n", tt.name, tt.c.String())
			},
		)
	}
}

func TestCell_Length(t *testing.T) {
	tests := setAndGetTests
	tests.name = "(set/get)Length"
	for _, tt := range tests.t {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.c.setLength(tt.set)
				got := tt.c.getLength()
				if got != tt.want {
					t.Errorf("%s() = %v, want %v", tests.name, got, tt.want)
				}
				// fmt.Printf("%q --> %q\n", tt.name, tt.c.String())
			},
		)
	}
}

func TestCell_String(t *testing.T) {
	tests := setAndGetTests
	tests.name = "Stringer"
	for _, tt := range tests.t {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.c.setID(tt.set)
				tt.c.setFlags(tt.set)
				tt.c.setOffset(tt.set)
				tt.c.setLength(tt.set)
				want := fmt.Sprintf("id=%d, flags=0x%.4x, offset=%d, length=%d", tt.set, tt.set, tt.set, tt.set)
				got := tt.c.String()
				if got != want {
					t.Errorf("got%q, want %q", got, want)
				}
			},
		)
	}
}
