package scanner

import (
	"testing"
)

func BenchmarkAny(t *testing.B) {
	data := []byte(`"Hello, 世界 - 生日快乐"`)

	for i := 0; i < t.N; i++ {
		end, err := Any(data, 0)
		if err != nil {
			t.FailNow()
			return
		}

		if end == 0 {
			t.FailNow()
			return
		}
	}
}

func TestAny(t *testing.T) {
	testCases := map[string]struct {
		In  string
		Out string
	}{
		"string": {
			In:  `"hello"`,
			Out: `"hello"`,
		},
		"array": {
			In:  `["a","b","c"]`,
			Out: `["a","b","c"]`,
		},
		"object": {
			In:  `{"a":"b"}`,
			Out: `{"a":"b"}`,
		},
		"number": {
			In:  `1.234e+10`,
			Out: `1.234e+10`,
		},
	}

	for label, tc := range testCases {
		t.Run(
			label, func(t *testing.T) {
				end, err := Any([]byte(tc.In), 0)
				if err != nil {
					t.FailNow()
				}
				data := tc.In[0:end]
				if string(data) != tc.Out {
					t.FailNow()
				}
			},
		)
	}
}
