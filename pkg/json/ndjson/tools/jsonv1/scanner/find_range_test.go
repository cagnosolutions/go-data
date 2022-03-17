package scanner

import (
	"testing"
)

func BenchmarkFindRange(t *testing.B) {
	data := []byte(`["a","b","c","d","e"]`)

	for i := 0; i < t.N; i++ {
		out, err := FindRange(data, 0, 1, 2)
		if err != nil {
			t.FailNow()
			return
		}

		if string(out) != `["b","c"]` {
			t.FailNow()
			return
		}
	}
}

func TestFindRange(t *testing.T) {
	testCases := map[string]struct {
		In       string
		From     int
		To       int
		Expected string
		HasErr   bool
	}{
		"simple": {
			In:       `["a","b","c","d","e"]`,
			From:     1,
			To:       2,
			Expected: `["b","c"]`,
		},
		"single": {
			In:       `["a","b","c","d","e"]`,
			From:     1,
			To:       1,
			Expected: `["b"]`,
		},
		"mixed": {
			In:       `["a",{"hello":"world"},"c","d","e"]`,
			From:     1,
			To:       1,
			Expected: `[{"hello":"world"}]`,
		},
		"ordering": {
			In:     `["a",{"hello":"world"},"c","d","e"]`,
			From:   1,
			To:     0,
			HasErr: true,
		},
		"out of bounds": {
			In:     `["a",{"hello":"world"},"c","d","e"]`,
			From:   1,
			To:     20,
			HasErr: true,
		},
	}

	for label, tc := range testCases {
		t.Run(
			label, func(t *testing.T) {
				data, err := FindRange([]byte(tc.In), 0, tc.From, tc.To)
				if tc.HasErr {
					if err == nil {
						t.FailNow()
					}
				} else {
					if string(data) != tc.Expected {
						t.FailNow()
					}
					if err != nil {
						t.FailNow()
					}
				}
			},
		)
	}
}
