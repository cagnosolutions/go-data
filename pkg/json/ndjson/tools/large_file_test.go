package tools

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/db/matcher"
	"github.com/cagnosolutions/go-data/pkg/json/ndjson"
	"github.com/cagnosolutions/go-data/pkg/json/ndjson/tools/jsonv1/scanner"
)

var name = "small_file.json"

func Benchmark_StdLib_Json_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		data := LoadFileData(name)
		var result struct{}
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_Tools_JsonV1_FindKey(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		data := LoadFileData(name)
		bfound, err := scanner.FindKey(data, 0, []byte("hello"))
		if err != nil {
			b.Error(err)
		}
		if string(bfound) != `"world"` {
			b.Errorf("expected=%v, got=%s\n", "world", bfound)
		}
	}
}

func Benchmark_Tools_JsonV1_FindKeyV2(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		data := LoadFileData(name)
		bfound, err := scanner.FindKey2(data, []byte("hello"))
		if err != nil {
			b.Error(err)
		}
		if string(bfound) != `"world"` {
			b.Errorf("expected=%v, got=%s\n", "world", bfound)
		}
	}
}

func Benchmark_Tools_PureDB_Matcher(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		data := LoadFileData(name)
		found := ndjson.Match(data, []byte(`"hello": "world"`), 1)
		if !found {
			b.Error("no match found")
		}
	}
}

func Benchmark_Tools_PureDB_LexerToken(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		data := LoadFileData(name)
		l := matcher.NewLexer(string(data))
		var tok matcher.Token
		for tok.GetType() != matcher.TokEOF {
			tok = l.NextToken()
			_ = tok
		}
	}
}

func Benchmark_StdLib_Json_Decoder(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		data := LoadFileData(name)
		var result struct{}
		dec := json.NewDecoder(bytes.NewReader(data))
		err := dec.Decode(&result)
		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_StdLib_Json_Decode_Token(b *testing.B) {
	b.ReportAllocs()
	searchkey, searchval, foundkey := "hello", "world", false
	for n := 0; n < b.N; n++ {
		data := LoadFileData(name)
		dec := json.NewDecoder(bytes.NewReader(data))
		for dec.More() {
			tok, err := dec.Token()
			if err != nil {
				if err != nil {
					b.Error(err)
				}
			}
			s, ok := tok.(string)
			if !ok {
				continue
			}
			if s == searchkey && !foundkey {
				foundkey = true
			}
			if s == searchval && foundkey {
				goto foundit
			}
		}
	foundit:
	}
}
