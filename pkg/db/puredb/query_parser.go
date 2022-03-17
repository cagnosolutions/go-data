package puredb

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// https://go.dev/play/p/wNjNXHI18_6

func runQueryParser() {

	query := `select * from users where age >= 18 and age < 66 or age > 17 or active = true and status = active;`
	qs := parseQuery(query)
	fmt.Printf("%s\n", qs)
	rec := []byte(`{"id":3,"fname":"john","lname":"doe","age":44,"active":true,"status":"inactive"}`)
	for _, fld := range getFields(rec) {
		ok := qs.Match(fld)
		fmt.Printf("field=%s, match=%v\n", fld, ok)
	}
}

func getFields(b []byte) [][]byte {
	return bytes.FieldsFunc(
		b, func(r rune) bool {
			return r == ',' || r == '{' || r == '}' || r == '\n'
		},
	)
}

func parseQuery(q string) *queryStmt {
	// first, we create statement strings from the supplied query
	sts := makeStatementStrings(q)
	// next, we parse each statement string into an actual statement
	qs := new(queryStmt)
	for _, st := range sts {
		qs.stmts = append(qs.stmts, parseStatement(st))
	}
	// and we return a query statement
	return qs
}

func makeStatementStrings(q string) []string {
	re := regexp.MustCompile(`(\s+or\s+)`)
	return re.Split(q, -1)
}

func parseStatement(s string) statement {
	re := regexp.MustCompile(`(\w+)\s+([!=><]{1,2}|like)\s+('?[a-z0-9_,.@%]+'?)`)
	sm := re.FindAllStringSubmatch(s, -1)
	var st statement
	for _, m := range sm {
		st.comps = append(st.comps, newCompTok(m[1], m[2], m[3]))
	}
	return st
}

func newCompTok(fld, op, val string) compTok {
	return compTok{
		// sm: getStrFmt(fld, val),
		fn:  getCompFn(op),
		fld: fld,
		val: val,
	}
}

func getStrFmt(fld, val string) string {
	re := regexp.MustCompile(`^(\d+|true|false)$`)
	ok := re.MatchString(val)
	if ok {
		return fmt.Sprintf("%q:%v", fld, val)
	}
	return fmt.Sprintf("%q:%q", fld, val)
}

func getCompFn(cs string) compFn {
	switch cs {
	case "=", "==":
		return eq
	case "!=", "<>":
		return ne
	case ">":
		return gt
	case ">=":
		return ge
	case "<":
		return lt
	case "<=":
		return le
	}
	return nil
}

type queryStmt struct {
	stmts []statement
}

func (qs *queryStmt) Match(data []byte) bool {
	for _, stmt := range qs.stmts {
		for _, comp := range stmt.comps {
			if !comp.fn(data, []byte(comp.matcher)) {
				return false
			}
		}
	}
	return true
}

func (qs *queryStmt) String() string {
	var ss string
	for _, ors := range qs.stmts {
		for _, ands := range ors.comps {
			ss += fmt.Sprintf("match: %s %s %s\n", ands.fld, ands.fn, ands.val)
		}
		ss += fmt.Sprintf("\tor\n")
	}
	return ss
}

type statement struct {
	comps []compTok
}

func (s *statement) String() string {
	var ss string
	for _, comp := range s.comps {
		ss += fmt.Sprintf("comp.fld=%s, comp.val=%v, comp.fn=%s\n", comp.fld, comp.val, comp.fn)
	}
	return ss
}

type compTok struct {
	// sm string
	fld     string
	val     string
	fn      compFn
	matcher string
}

// comparison function type
type compFn func(a, b []byte) bool

func (fn compFn) String() string {
	s := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	switch strings.ReplaceAll(s, "main.", "") {
	case "eq":
		return "="
	case "ne":
		return "!="
	case "gt":
		return ">"
	case "ge":
		return ">="
	case "lt":
		return "<"
	case "le":
		return "<="
	}
	return "compFn"
}

// '=' or '=='
func eq(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// '!=' or '<>'
func ne(a, b []byte) bool {
	return !bytes.Equal(a, b)
}

// '>'
func gt(a, b []byte) bool {
	return bytes.Compare(a, b) > 0
}

// '>='
func ge(a, b []byte) bool {
	return bytes.Compare(a, b) >= 0
}

// '<'
func lt(a, b []byte) bool {
	return bytes.Compare(a, b) < 0
}

// '<='
func le(a, b []byte) bool {
	return bytes.Compare(a, b) <= 0
}
