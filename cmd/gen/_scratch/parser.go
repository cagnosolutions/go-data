package _scratch

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

type Parser struct {
	ModelsByName map[string]interface{}
}

func NewModel(v interface{}) interface{} {
	return nil
}

func (p *Parser) Parse() error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, os.Getenv("GOFILE"), nil, 0)
	if err != nil {
		return err
	}
	ast.Inspect(
		f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				err = p.parseType(x)
				if err != nil {
					return false
				}
			}
			return true
		},
	)
	return p.generate()
}

func (p *Parser) parseType(st *ast.TypeSpec) error {
	if strings.HasSuffix(st.Name.Name, "Model") {
		model := strings.Replace(st.Name.Name, "Model", "", -1)
		if p.ModelsByName[model] == nil {
			p.ModelsByName[model] = NewModel(model)
		}
	}
	return nil
}

func (p *Parser) generate() error {
	return nil
}

func (p *Parser) generateCode(tmpl *template.Template, fn string) error {
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, p)
	if err != nil {
		return err
	}
	res, err := imports.Process(fn, buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return os.WriteFile(fn, res, 0666)
}
