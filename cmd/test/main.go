package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

var exampleSrc = `package main

// foo is a function
func foo() string {
	return "bar"
}`

func main() {
	Parse("", exampleSrc)
}

func Parse(filename string, src interface{}) {
	files := token.NewFileSet()
	f, err := parser.ParseFile(files, filename, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	err = ast.Print(files, f)
	if err != nil {
		log.Fatal(err)
	}
}
