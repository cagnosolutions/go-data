package matcher

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/json/ndjson"
)

func BenchmarkHandRoll(b *testing.B) {
	input := []byte(`{"_id":5,"f_name":"Maggie","l_name":"Smith","age":35,"email":"msmith@gmail.com","active":true}`)
	_ = input
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ndjson.Match(input, []byte(`age:35`), 1)
	}
}

func BenchmarkLexerNextToken(b *testing.B) {
	input := `{"_id":5,"f_name":"Maggie","l_name":"Smith","age":35,"email":"msmith@gmail.com","active":true}`
	// fmt.Printf("input=%s\n", input)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		var tok Token
		for tok.typ != tokEOF {
			tok = l.NextToken()
			_ = tok
			// fmt.Printf("%s\n", tok)
		}
	}
}

type User struct {
	ID     int    `json:"_id"`
	FName  string `json:"f_name"`
	LName  string `json:"l_name"`
	Age    int    `json:"age"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

func BenchmarkJSONDecoderNextToken(b *testing.B) {
	input := []byte(`{"_id":5,"f_name":"Maggie","l_name":"Smith","age":35,"email":"msmith@gmail.com","active":true}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := json.NewDecoder(bytes.NewReader(input))
		// read open bracket
		// _, err := dec.Token()
		// if err != nil {
		//	log.Fatal("1:" + err.Error())
		// }
		// fmt.Printf("%T: %v\n", tok, tok)
		// while the array contains values
		for dec.More() {
			var u User
			// decode an array value (Message)
			err := dec.Decode(&u)
			if err != nil {
				log.Fatal("2:" + err.Error())
			}
			// fmt.Printf("%v: %v\n", m.Name, m.Text)
		}
		// read closing bracket
		// _, err = dec.Token()
		// if err != nil {
		//	log.Fatal("3:" + err.Error())
		// }
		// fmt.Printf("%T: %v\n", tok, tok)
	}

}
