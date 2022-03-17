package puredb

import (
	"fmt"
	"regexp"
	"regexp/syntax"
)

const queryCheckerRegex = `^((\w+)\s([=!<>]{1,2}|contains)\s('?[a-z0-9_.@]+'?|'')(;$|\sand\s|\sor\s))+`
const queryCheckerRegex2 = `(?m)(?P<field>\w+)\s+(?P<operation>[=!<>]{1,2}|contains)\s+(?P<value>'?[a-z0-9_.@]+'?|'')\s+(?P<condition>and|or)\s+`
const queryCheckerRegex3 = `(?m)(?P<field>\w+)\s+(?P<operation>[=!<>]{1,
2}|contains)\s+(?P<value>'?[a-z0-9_.@]+'?|'')(;$)`
const queryCheckerRegex4 = `(?m)(?P<field>\w+)\s+(?P<op>[=!<>]{1,2}|contains)\s+(?P<value>'?[a-z0-9_.@]+'?|'')(?P<end>\s+(and|or)\s+|;)`

var queryChecker *regexp.Regexp
var queryChecker2 *regexp.Regexp
var queryChecker3 *regexp.Regexp
var queryChecker4 *regexp.Regexp

func init() {
	queryChecker = regexp.MustCompile(queryCheckerRegex)
	queryChecker2 = regexp.MustCompile(queryCheckerRegex2)
	queryChecker3 = regexp.MustCompile(queryCheckerRegex3)
	queryChecker4 = regexp.MustCompile(queryCheckerRegex4)
}

func IsOK(query string) bool {
	return queryChecker.MatchString(query)
}

func Expand(query string) []byte {
	// var split []string
	// if strings.Contains(query, "and") {
	// 	split = strings.Split(query, "and")
	// }
	// if strings.Contains(query, "or") {
	// 	split = strings.Split(query, "or")
	// }
	// for i := range split {
	// 	strings.Trim(split[i], "")
	// }
	pattern := queryChecker4
	if !pattern.MatchString(query) {
		return nil
	}
	template := `field="$field", op="$op", value="$value", end="$end"`
	var result []byte
	for _, submatches := range pattern.FindAllStringSubmatchIndex(query, -1) {
		result = pattern.ExpandString(result, template, query, submatches)
	}
	return result
}

type Query struct {
}

func parse(query string) string {
	re, err := syntax.Parse(queryCheckerRegex, syntax.FoldCase|syntax.Perl)
	if err != nil {
		panic(err)
	}
	ss := fmt.Sprintf("CapNames=%+v\n", re.CapNames())
	ss += re.String()
	return ss
}

var selectStmt = `SELECT $column_list FROM $table_name WHERE $search_condition;`
var searchCond = ``
