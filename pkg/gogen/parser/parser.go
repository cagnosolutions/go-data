package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"text/template"
)

type Parser struct {
	file *ast.File
	tmpl *template.Template
	fset *token.FileSet
}

func NewParser(filename string) (*Parser, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return &Parser{
		file: f,
		fset: fset,
	}, nil
}

func NewParserWithTemplate(filename string, templatePattern string) (*Parser, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return &Parser{
		file: f,
		tmpl: template.Must(template.New("").ParseGlob(templatePattern)),
		fset: fset,
	}, nil
}

func (p *Parser) File() *ast.File {
	return p.file
}

func (p *Parser) ParseStructSimple() error {
	out := collectStructs(p.file, p.fset)
	for pos, str := range out {
		fmt.Printf("%d: %s\n", pos, str)
	}
	return nil
}

func (p *Parser) getLine(pos token.Pos) int {
	return p.fset.Position(pos).Line
}

func (p *Parser) Find(n ast.Node) {
	switch x := n.(type) {
	case *ast.Package:
		fmt.Printf(
			"%d:%T: name=%q, files=%d\n",
			p.getLine(x.Pos()),
			x,
			x.Name,
			len(x.Files),
		)
	case *ast.File:
		fmt.Printf(
			"%d:%T: name=%q\n",
			p.getLine(x.Pos()),
			x,
			x.Name,
		)
	case *ast.TypeSpec:
		fmt.Printf(
			"%d:%T: name=%q\n",
			p.getLine(x.Pos()),
			x,
			x.Name,
		)
	}
}

func (p *Parser) ParseStruct() error {
	st := new(structType)
	var name string
	var ct commentType
	ast.Inspect(
		p.file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Comment:
				ct.value = x.Text
				ct.line = p.fset.Position(x.Pos()).Line
				return true
			case *ast.TypeSpec:
				if x.Type == nil {
					return true
				}
				name = x.Name.Name
				// return true
				// _, ok := x.Type.(*ast.StructType)
				// if !ok {
				// 	return true
				// }
				// fmt.Printf(
				// 	"(%d) Type.Comment: %v, Type.Name: %v, Type.IsStruct: %v, Type.Params: %v\n",
				// 	x.Pos(),
				// 	x.Comment.Text(),
				// 	x.Name,
				// 	ok,
				// 	x.TypeParams,
				// )
				// return true
			case *ast.StructType:
				st = &structType{
					name:    name,
					fields:  make([]fieldType, x.Fields.NumFields()),
					node:    x,
					begLine: p.fset.Position(x.Pos()).Line,
					endLine: p.fset.Position(x.End()).Line,
				}
				if ct.line > 0 && ct.value != "" {
					if st.begLine-1 == ct.line {
						st.comment = ct
					}
				}
				for i, field := range x.Fields.List {
					st.fields[i] = fieldType{
						name:  field.Names[0].Name,
						kind:  field.Names[0].Obj.Kind.String(),
						value: field.Names[0].Obj.Name,
						tag:   field.Tag.Value,
						line:  p.fset.Position(field.Pos()).Line,
					}
					// fmt.Printf("(%d) struct...\n", x.Pos())
					// fmt.Printf("\nField: %s\n", field.Names[0].Name)
					// fmt.Printf("\nTag:   %s\n", field.Tag.Value)
				}
				return false
			default:
				return true
			}
			return true
		},
	)
	fmt.Printf("%s", st)
	return nil
}

func (p *Parser) ParseIter(fn func(n ast.Node) bool) error {
	ast.Inspect(p.file, fn)
	return nil
}

func (p *Parser) Parse() error {
	ast.Inspect(
		p.file, func(node ast.Node) bool {
			fmt.Printf("(%T) %#v\n", node, node)
			return true
		},
	)
	return nil
}

func (p *Parser) Parse2() error {
	ast.Inspect(
		p.file, func(n ast.Node) bool {
			var s string
			switch x := n.(type) {
			case *ast.BasicLit:
				s = x.Value
			case *ast.Ident:
				s = x.Name
			}
			if s != "" {
				fmt.Printf("%s:\t%s\n", p.fset.Position(n.Pos()), s)
			}
			return true
		},
	)
	return nil
}
