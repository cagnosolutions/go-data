package bytes

import (
	"strings"
	"testing"
)

type prefixTest struct {
	str string // string to test
	pre string // prefix to test
	res bool   // result should be
}

func (v *prefixTest) err(b *testing.B, got bool) {
	b.Errorf("str=%q, pre=%q, res=%v (got=%v)\n", v.str, v.pre, v.res, got)
}

var prefixTests = []prefixTest{
	{"foo", "fo", true},
	{"foo", "f", true},
	{"foo", "fan", false},
	{"food", "foo", true},
	{"food", "foam", false},
	{"food", "fun", false},
	{"plan", "plan", true},
	{"plan", "pl", true},
	{"plan", "plz", false},
	{"planet", "plan", true},
	{"planet", "plane", true},
	{"planet", "plum", false},
	{"moo", "mo", true},
	{"moo", "mon", false},
	{"mood", "m", true},
	{"moon", "mom", false},
	{"nope", "no", true},
	{"nope", "nop", true},
	{"nope", "nan", false},
	{"noob", "noon", false},
	{"noob", "no", true},
	{"noob", "nooz", false},
	{"random", "rand", true},
	// {"random", "randomly", false},
	{"abcdefghijklmnopqrstuvwxyz", "abcdefghijklmnopqrstuvwxy", true},
	{"abcdefghijklmnopqrstuvwxyz", "abcdefghijkl", true},
	{"abcdefghijklmnopqrstuvwxyz", "bcdefghijklmnopqrstuvwxyz", false},
	{"abcdefghijklmnopqrstuvwxyz", "abcdeghijklmnopqrstuvxyz", false},
	{"abcdefghijklmnopqrstuvwxyz", "a", true},
	{"abcdefghijklmnopqrstuvwxyz", "abcd", true},
	{"0123456789", "012", true},
	{"0123456789", "0123456781", false},
	{"tomorrow", "tom", true},
	{"tomorrow", "tommy", false},
	{"tomorrow", "to", true},
	{"tomorrow", "x", false},
	{"the month of august", "the monkeys go wild", false},
	{"the month of august", "the man", false},
	{"the month of august", "the month", true},
	{"philadelphia", "phil", true},
	{"philadelphia", "philly", false},
}

func slicePrefix_v1(s string, pre string) bool {
	if s[:len(pre)] == pre {
		return true
	}
	return false
}

func slicePrefix_v2(s string, pre string) bool {
	return s[:len(pre)] == pre
}

func hasPrefixFast(s string, pre string) bool {
	if len(pre) > len(s) {
		return false
	}
	for i := len(pre) - 1; i > 0; i-- {
		if s[i] != pre[i] {
			return false
		}
	}
	return true
}

func BenchmarkPrefix_StdLib(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, v := range prefixTests {
			got := strings.HasPrefix(v.str, v.pre)
			if got != v.res {
				v.err(b, got)
			}
		}
	}
}

func BenchmarkPrefix_SlicePrefixV1(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, v := range prefixTests {
			got := slicePrefix_v1(v.str, v.pre)
			if got != v.res {
				v.err(b, got)
			}
		}
	}
}

func BenchmarkPrefix_SlicePrefixV2(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, v := range prefixTests {
			got := slicePrefix_v2(v.str, v.pre)
			if got != v.res {
				v.err(b, got)
			}
		}
	}
}

func BenchmarkPrefix_Fast(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, v := range prefixTests {
			got := hasPrefixFast(v.str, v.pre)
			if got != v.res {
				v.err(b, got)
			}
		}
	}
}
