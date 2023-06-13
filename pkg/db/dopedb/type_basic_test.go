package dopedb

import (
	"bytes"
	"reflect"
	"testing"
)

func TestTypeEncodersAndDecoders(t *testing.T) {

	errf := func(t *testing.T, name string, got any, want any) {
		t.Errorf("[%s] got=(%T) %#v, want=(%T) %#v\n", name, got, got, want, want)
	}

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			"test nil",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := any(nil)
				Primitive.EncNil(buf)
				v := Primitive.DecNil(buf)
				if v != d {
					errf(t, "nil", v, d)
				}
			},
		},
		{
			"test bool (true)",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := true
				Primitive.EncBool(buf, d)
				v := Primitive.DecBool(buf)
				if v != d {
					errf(t, "bool (true)", v, d)
				}
			},
		},
		{
			"test bool (false)",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := false
				Primitive.EncBool(buf, d)
				v := Primitive.DecBool(buf)
				if v != d {
					errf(t, "bool (false)", v, d)
				}
			},
		},
		{
			"test float32",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := float32(3.14159)
				Primitive.EncFloat32(buf, d)
				v := Primitive.DecFloat32(buf)
				if v != d {
					errf(t, "float32", v, d)
				}
			},
		},
		{
			"test float64",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := float64(3.14159)
				Primitive.EncFloat64(buf, d)
				v := Primitive.DecFloat64(buf)
				if v != d {
					errf(t, "float64", v, d)
				}
			},
		},
		{
			"test uint8",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := uint8(123)
				Primitive.EncUint8(buf, d)
				v := Primitive.DecUint8(buf)
				if v != d {
					errf(t, "uint8", v, d)
				}
			},
		},
		{
			"test uint16",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := uint16(12738)
				Primitive.EncUint16(buf, d)
				v := Primitive.DecUint16(buf)
				if v != d {
					errf(t, "uint16", v, d)
				}
			},
		},
		{
			"test uint32",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := uint32(849032192)
				Primitive.EncUint32(buf, d)
				v := Primitive.DecUint32(buf)
				if v != d {
					errf(t, "uint32", v, d)
				}
			},
		},
		{
			"test uint64",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := uint64(43893829489038001)
				Primitive.EncUint64(buf, d)
				v := Primitive.DecUint64(buf)
				if v != d {
					errf(t, "int32", v, d)
				}
			},
		},
		{
			"test int8",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := int8(125)
				Primitive.EncInt8(buf, d)
				v := Primitive.DecInt8(buf)
				if v != d {
					errf(t, "int8", v, d)
				}
			},
		},
		{
			"test int16",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := int16(12738)
				Primitive.EncInt16(buf, d)
				v := Primitive.DecInt16(buf)
				if v != d {
					errf(t, "int16", v, d)
				}
			},
		},
		{
			"test int32",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := int32(849032192)
				Primitive.EncInt32(buf, d)
				v := Primitive.DecInt32(buf)
				if v != d {
					errf(t, "int32", v, d)
				}
			},
		},
		{
			"test int64",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := int64(43893829489038001)
				Primitive.EncInt64(buf, d)
				v := Primitive.DecInt64(buf)
				if v != d {
					errf(t, "int64", v, d)
				}
			},
		},
		{
			"test uint",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := uint(43893829)
				Primitive.EncUint(buf, d)
				v := Primitive.DecUint(buf)
				if v != d {
					errf(t, "uint", v, d)
				}
			},
		},
		{
			"test int",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := int(43893829)
				Primitive.EncInt(buf, d)
				v := Primitive.DecInt(buf)
				if v != d {
					errf(t, "int", v, d)
				}
			},
		},
		{
			"test fix str",
			func(t *testing.T) {
				buf := make([]byte, 16)
				d := "foo bar"
				Strings.EncFixStr(buf, d)
				v := Strings.DecFixStr(buf)
				if v != d {
					errf(t, "fix str", v, d)
				}
			},
		},
		{
			"test str8",
			func(t *testing.T) {
				buf := make([]byte, 64)
				d := "this is a test, this is only a test."
				Strings.EncStr8(buf, d)
				v := Strings.DecStr8(buf)
				if v != d {
					errf(t, "str8", v, d)
				}
			},
		},
		{
			"test str16",
			func(t *testing.T) {
				buf := make([]byte, 512)
				d := `It is for us the living, rather, to be dedicated here to the
unfinished work which they who fought here have thus far so
nobly advanced.  It is rather for us to be here dedicated to
the great task remaining before us - that from these honored
dead we take increased devotion to that cause for which they
gave the last full measure of devotion`
				Strings.EncStr16(buf, d)
				v := Strings.DecStr16(buf)
				if v != d {
					errf(t, "str16", v, d)
				}
			},
		},
		{
			"test bin8",
			func(t *testing.T) {
				buf := make([]byte, 64)
				d := []byte("this is a test, this is only a test.")
				Strings.EncBin8(buf, d)
				v := Strings.DecBin8(buf)
				if !bytes.Equal(v, d) {
					errf(t, "bin8", v, d)
				}
			},
		},
		{
			"test bin16",
			func(t *testing.T) {
				buf := make([]byte, 512)
				d := []byte(`It is for us the living, rather, to be dedicated here to the
unfinished work which they who fought here have thus far so
nobly advanced.  It is rather for us to be here dedicated to
the great task remaining before us - that from these honored
dead we take increased devotion to that cause for which they
gave the last full measure of devotion`)
				Strings.EncBin16(buf, d)
				v := Strings.DecBin16(buf)
				if !bytes.Equal(v, d) {
					errf(t, "bin16", v, d)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func Test_primitiveTypes_DecFixInt(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecFixInt(tt.args.p); got != tt.want {
					t.Errorf("DecFixInt() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecFloat32(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecFloat32(tt.args.p); got != tt.want {
					t.Errorf("DecFloat32() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecFloat64(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecFloat64(tt.args.p); got != tt.want {
					t.Errorf("DecFloat64() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecInt(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecInt(tt.args.p); got != tt.want {
					t.Errorf("DecInt() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecInt16(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want int16
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecInt16(tt.args.p); got != tt.want {
					t.Errorf("DecInt16() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecInt32(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecInt32(tt.args.p); got != tt.want {
					t.Errorf("DecInt32() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecInt64(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecInt64(tt.args.p); got != tt.want {
					t.Errorf("DecInt64() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecInt8(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want int8
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecInt8(tt.args.p); got != tt.want {
					t.Errorf("DecInt8() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecNil(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecNil(tt.args.p); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DecNil() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecUint(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecUint(tt.args.p); got != tt.want {
					t.Errorf("DecUint() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecUint16(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecUint16(tt.args.p); got != tt.want {
					t.Errorf("DecUint16() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecUint32(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecUint32(tt.args.p); got != tt.want {
					t.Errorf("DecUint32() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecUint64(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecUint64(tt.args.p); got != tt.want {
					t.Errorf("DecUint64() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_DecUint8(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				if got := e.DecUint8(tt.args.p); got != tt.want {
					t.Errorf("DecUint8() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_primitiveTypes_EncBool(t *testing.T) {
	type args struct {
		p []byte
		v bool
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncBool(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncFixInt(t *testing.T) {
	type args struct {
		p []byte
		v int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncFixInt(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncFloat32(t *testing.T) {
	type args struct {
		p []byte
		v float32
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncFloat32(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncFloat64(t *testing.T) {
	type args struct {
		p []byte
		v float64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncFloat64(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncInt(t *testing.T) {
	type args struct {
		p []byte
		v int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncInt(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncInt16(t *testing.T) {
	type args struct {
		p []byte
		v int16
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncInt16(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncInt32(t *testing.T) {
	type args struct {
		p []byte
		v int32
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncInt32(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncInt64(t *testing.T) {
	type args struct {
		p []byte
		v int64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncInt64(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncInt8(t *testing.T) {
	type args struct {
		p []byte
		v int8
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncInt8(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncNil(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncNil(tt.args.p)
			},
		)
	}
}

func Test_primitiveTypes_EncUint(t *testing.T) {
	type args struct {
		p []byte
		v uint
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncUint(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncUint16(t *testing.T) {
	type args struct {
		p []byte
		v uint16
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncUint16(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncUint32(t *testing.T) {
	type args struct {
		p []byte
		v uint32
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncUint32(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncUint64(t *testing.T) {
	type args struct {
		p []byte
		v uint64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncUint64(tt.args.p, tt.args.v)
			},
		)
	}
}

func Test_primitiveTypes_EncUint8(t *testing.T) {
	type args struct {
		p []byte
		v uint8
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := primitiveTypes{}
				e.EncUint8(tt.args.p, tt.args.v)
			},
		)
	}
}
