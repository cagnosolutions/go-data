package scanner

import (
	"testing"
)

func BenchmarkFindKey(t *testing.B) {
	data := []byte(dat1)
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		out, err := FindKey(data, 0, []byte("first"))
		if err != nil {
			t.FailNow()
			return
		}

		if string(out) != `"got it"` {
			t.FailNow()
			return
		}
	}
}

func TestFindKey(t *testing.T) {
	testCases := map[string]struct {
		In       string
		Key      string
		Expected string
		HasErr   bool
	}{
		"simple": {
			In:       `{"hello":"world"}`,
			Key:      "hello",
			Expected: `"world"`,
		},
		"spaced": {
			In:       ` { "hello" : "world" } `,
			Key:      "hello",
			Expected: `"world"`,
		},
	}

	for label, tc := range testCases {
		t.Run(
			label, func(t *testing.T) {
				data, err := FindKey([]byte(tc.In), 0, []byte(tc.Key))
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

var dat1 = `{
  "id":234,
  "age":35,
  "email":"test@example.com",
  "name":[
    "john",
    "doe"
  ],
  "nested_object":{
	"first":"got it",
    "more":{
      "somekey":"somevalue"
    }
  },
  "active":true,
  "hello":"world"
}`
