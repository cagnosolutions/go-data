package puredb

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type queryStr struct {
	str    string
	passes bool
}

var queries = []queryStr{
	{str: "name=alex;", passes: false},
	{str: "name = 'alex';", passes: true},
	{str: "id = 2;", passes: true},
	{str: "id == 3;", passes: true},
	{str: "id != null;", passes: true},
	{str: "age >= 18 and age < 66;", passes: true},
	{str: "active = true;", passes: true},
	{str: "active != false;", passes: true},
	{str: "'somefield' == 'someotherfield';", passes: false},
	{str: "email = 'jdoe@gmail.com';", passes: true},
	{str: "email = foobar@example.net;", passes: true},
	{str: "name contains doe and age > 18;", passes: true},
	{str: "contains 'doe' or 'foobar';", passes: false},
	{str: "age >= 18 and age < 66 or age > 17;", passes: true},
	{str: "active != true;", passes: true},
	{str: "active is not true;", passes: false},
	{str: "active not true;", passes: false},
}

var ErrBadQuery = errors.New("bad query format")
var query = "age >= 18 and age < 66 or age > 17;"
var record = `{"_id":1,"age":24,"name":"john doe","email":"jdoe@example.com"}`

func ParseQuery(query string, record string) error {
	if !IsOK(query) {
		return ErrBadQuery
	}
	args := strings.Fields(strings.ReplaceAll(query, ";", ""))
	fmt.Printf("%+#v\n", args)
	return nil
}

func TestQuerySyntaxParse(t *testing.T) {
	err := ParseQuery(query, record)
	if err != nil {
		t.Error(err)
	}
}

func TestQueryPasses(t *testing.T) {
	for _, query := range queries {
		ok := IsOK(query.str)
		if ok != query.passes {
			t.Errorf("got=%v, expected=%v [query=%q]\n", ok, query.passes, query.str)
		}
	}
}

func TestQueryPrintSubs(t *testing.T) {
	for _, query := range queries {
		s := Expand(query.str)
		if s == nil {
			continue
		}
		fmt.Printf("submatch=%s, [query=%q]\n", s, query.str)
	}
}
