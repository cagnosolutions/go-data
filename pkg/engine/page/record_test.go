package page

import (
	"fmt"
	"testing"
)

func TestRecord_Flags(t *testing.T) {
	tests := []struct {
		want  uint8
		got   uint8
		valid bool
	}{
		// plain types
		{0x11, R_NUM_NUM, true},
		{0x12, R_NUM_STR, true},
		{0x21, R_STR_NUM, true},
		{0x22, R_STR_STR, true},

		// pointer types
		{0x14, R_NUM_PTR, true},
		{0x24, R_STR_PTR, true},
	}
	for i, tt := range tests {
		if tt.got != tt.want {
			fmt.Printf("[%d] WARN: got=0x%.4x, want=0x%.4x\n", i, tt.got, tt.want)
		}
		_, err := newRecordHeader(tt.got, 5, 26)
		if err != nil && tt.valid == true {
			t.Errorf("[%d] got error with flags (0x%.4x): %s\n", i, tt.got, err)
		}
	}
}

func makeValue(i byte) []byte {
	return []byte{
		't', 'h', 'i', 's',
		' ',
		'i', 's',
		' ',
		'r', 'e', 'c', 'o', 'r', 'd',
		' ',
		'n', 'u', 'm', 'b', 'e', 'r', ':',
		' ',
		i,
	}
}

func BenchmarkNewRecord(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 255; j++ {
			rec := NewRecord(R_NUM, R_STR, []byte{byte(j)}, makeValue(byte(j)))
			if rec == nil {
				b.Errorf("record is nil")
			}
			if rec.KeyType() != R_NUM&R_KEY {
				b.Errorf("bad key type, got=0x%.2x, want=0x%.2x\n", rec.KeyType(), R_NUM&R_KEY)
			}
			if rec.ValType() != R_STR&R_VAL {
				b.Errorf("bad val type, got=0x%.2x, want=0x%.2x\n", rec.ValType(), R_STR&R_VAL)
			}
			if len(rec.Key()) != 1 {
				b.Errorf("bad key size, got=%d, want>=24\n", len(rec.Key()))
			}
			if len(rec.Val()) < 24 {
				b.Errorf("bad val size, got=%d, want>=24\n", len(rec.Val()))
			}
		}
	}
}

type record struct {
	flags uint8
	key   []byte
	val   []byte
}

func newRecord2(keyType uint8, key []byte, valType uint8, val []byte) *record {
	return &record{
		flags: keyType | valType,
		key:   key,
		val:   val,
	}
}

func (r *record) KeyType() uint8 {
	return r.flags & R_KEY
}

func (r *record) ValType() uint8 {
	return r.flags & R_VAL
}

func (r *record) Key() []byte {
	return r.key
}

func (r *record) Val() []byte {
	return r.val
}

func (r *record) encode() []byte {
	b := make([]byte, 1+4+len(r.key)+len(r.val))
	var n int
	b[n] = r.flags
	n++
	b[n] = byte(len(r.key))
	n++
	b[n] = byte(len(r.key) >> 8)
	n++
	b[n] = byte(len(r.val))
	n++
	b[n] = byte(len(r.val) >> 8)
	n++
	n += copy(b[n:n+len(r.key)], r.key)
	n += copy(b[n:], r.val)
	_ = n
	return b
}

func (r *record) decode(b []byte) {
	var n int
	r.flags = b[n]
	n++
	klen := int(uint16(b[n]) | uint16(b[n+1])<<8)
	n += 2
	vlen := int(uint16(b[n]) | uint16(b[n+1])<<8)
	n += 2
	r.key = b[n : n+klen]
	n += len(r.key)
	r.val = b[n : n+vlen]
}

func BenchmarkNewRecord2(b *testing.B) {
	b.ReportAllocs()
	var rec *record
	var recB []byte
	for i := 0; i < b.N; i++ {
		for j := 0; j < 255; j++ {
			rec = newRecord2(R_KEY&R_NUM, []byte{byte(j)}, R_VAL&R_STR, makeValue(byte(j)))
			if rec == nil {
				b.Errorf("record is nil")
			}
			recB = rec.encode()
			if rec == nil {
				b.Errorf("record encode failed")
			}
			rec.flags = 0
			rec.decode(recB)
			if rec.flags == 0 {
				b.Errorf("record decode failed")
			}
			if rec.KeyType() != R_NUM&R_KEY {
				b.Errorf("bad key type, got=0x%.2x, want=0x%.2x\n", rec.KeyType(), R_NUM&R_KEY)
			}
			if rec.ValType() != R_STR&R_VAL {
				b.Errorf("bad val type, got=0x%.2x, want=0x%.2x\n", rec.ValType(), R_STR&R_VAL)
			}
			if len(rec.Key()) != 1 {
				b.Errorf("bad key size, got=%d, want>=24\n", len(rec.Key()))
			}
			if len(rec.Val()) < 24 {
				b.Errorf("bad val size, got=%d, want>=24\n", len(rec.Val()))
			}
		}
	}
}
