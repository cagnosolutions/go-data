package scanner

import (
	"testing"
)

func BenchmarkFindIndex(t *testing.B) {
	data := []byte(`["hello","world"]`)

	for i := 0; i < t.N; i++ {
		data, err := FindIndex(data, 0, 1)
		if err != nil {
			t.FailNow()
			return
		}

		if string(data) != `"world"` {
			t.FailNow()
			return
		}
	}
}

func TestFindIndex(t *testing.T) {
	testCases := map[string]struct {
		In       string
		Index    int
		Expected string
		HasErr   bool
	}{
		"simple": {
			In:       `["hello","world"]`,
			Index:    1,
			Expected: `"world"`,
		},
		"spaced": {
			In:       ` [ "hello" , "world" ] `,
			Index:    1,
			Expected: `"world"`,
		},
		"all types": {
			In:       ` [ "hello" , 123, {"hello":"world"} ] `,
			Index:    2,
			Expected: `{"hello":"world"}`,
		},
	}

	for label, tc := range testCases {
		t.Run(
			label, func(t *testing.T) {
				data, err := FindIndex([]byte(tc.In), 0, tc.Index)
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
