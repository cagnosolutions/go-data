package format

import (
	"regexp"
	"strings"
	"testing"
)

var toSnakeCases = []struct {
	args string
	want string
}{
	{"", ""},
	{"camelCase", "camel_case"},
	{"PascalCase", "pascal_case"},
	{"snake_case", "snake_case"},
	{"Pascal_Snake", "pascal_snake"},
	{"SCREAMING_SNAKE", "screaming_snake"},
	{"kebab-case", "kebab_case"},
	{"Pascal-Kebab", "pascal_kebab"},
	{"SCREAMING-KEBAB", "screaming_kebab"},
	{"A", "a"},
	{"AA", "aa"},
	{"AAA", "aaa"},
	{"AAAA", "aaaa"},
	{"AaAa", "aa_aa"},
	{"HTTPRequest", "http_request"},
	{"BatteryLifeValue", "battery_life_value"},
	{"Id0Value", "id0_value"},
	{"ID0Value", "id0_value"},
	{"MyLIFEIsAwesomE", "my_life_is_awesom_e"},
	{"Japan125Canada130Australia150", "japan125_canada130_australia150"},
	{"JapanCanadaAustralia", "japan_canada_australia"},
	{"ID", "id"},
	{"Id", "id"},
	{"_ID", "_id"},
	{"IsActive", "is_active"},
	{"EmailAddress", "email_address"},
	{"Sweden125Nevada130Mexico150", "sweden125_nevada130_mexico150"},
	{"ID123", "id123"},
	{"Column1", "column1"},
	{"Foo123Bar456Baz789", "foo123_bar456_baz789"},
}

func BenchmarkToSnakeCase(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, s := range toSnakeCases {
			out := ToSnakeCase(s.args)
			if out != s.want {
				b.Errorf("out=%q, wanted=%q\n", out, s.want)
			}
		}
	}
}

func BenchmarkToSnakeCaseCopy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, s := range toSnakeCases {
			out := _ToSnakeCaseCopy(s.args)
			if out != s.want {
				b.Errorf("out=%q, wanted=%q\n", out, s.want)
			}
		}
	}
}

func _BenchmarkToSnakeCaseRegex(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, s := range toSnakeCases {
			out := ToSnakeCaseRegex(s.args)
			if out != s.want {
				b.Errorf("out=%q, wanted=%q\n", out, s.want)
			}
		}
	}
}

var matchFirstCap = regexp.MustCompile("([ ])([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([- a-z0-9])([A-Z])")

func ToSnakeCaseRegex(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
