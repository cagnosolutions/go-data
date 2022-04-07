package format

import (
	"testing"
)

var toCamelCases = []struct {
	args string
	want string
}{
	{"", ""},
	{"camel_case", "CamelCase"},
	{"pascal_case", "PascalCase"},
	{"snake_case", "SnakeCase"},
	{"pascal_snake", "PascalSnake"},
	{"a", "A"},
	{"aa", "Aa"},
	{"aaa", "Aaa"},
	{"aaaa", "Aaaa"},
	{"aa_aa", "AaAa"},
	{"http_request", "HttpRequest"},
	{"battery_life_value", "BatteryLifeValue"},
	{"id0_value", "Id0Value"},
	{"my_life_is_awesom_e", "MyLifeIsAwesomE"},
	{"japan125_canada130_australia150", "Japan125Canada130Australia150"},
	{"japan_canada_australia", "JapanCanadaAustralia"},
	{"id", "Id"},
	{"_id", "_Id"},
	{"is_active", "IsActive"},
	{"email_address", "EmailAddress"},
	{"sweden125_nevada130_mexico150", "Sweden125Nevada130Mexico150"},
	{"id123", "Id123"},
	{"column1", "Column1"},
	{"foo123_bar456_baz789", "Foo123Bar456Baz789"},
}

func BenchmarkToCamelCase(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, s := range toCamelCases {
			out := ToCamelCase(s.args)
			if out != s.want {
				b.Errorf("out=%q, wanted=%q\n", out, s.want)
			}
		}
	}
}
