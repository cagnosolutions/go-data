package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"
)

type CommentType struct {
	Line int
	Text string
}

type FieldType struct {
	Name  string
	Type  string
	Value interface{}
	Tag   string
}

type StructType struct {
	Name   string
	Fields []FieldType
}

func (st *StructType) String() string {
	ss := fmt.Sprintf("type %s struct {\n", st.Name)
	for _, f := range st.Fields {
		ss += fmt.Sprintf("\t%s\t%s\n", f.Name, f.Type)
	}
	ss += fmt.Sprintf("}")
	return ss
}

type FunctionType struct {
	Name string
}

type MethodType struct {
	Name string
}

type Sorcerer struct {
	fs        *token.FileSet
	f         *ast.File
	PkgName   string
	Comments  []CommentType
	Structs   map[string]StructType
	Functions []FunctionType
	Methods   []MethodType
}

func NewSorcerer() *Sorcerer {
	return &Sorcerer{
		fs:        token.NewFileSet(),
		PkgName:   "",
		Comments:  make([]CommentType, 0),
		Structs:   make(map[string]StructType, 0),
		Functions: make([]FunctionType, 0),
		Methods:   make([]MethodType, 0),
	}
}

func (s *Sorcerer) Position(p token.Pos) token.Position {
	return s.fs.Position(p)
}

func (s *Sorcerer) astComment(x *ast.Comment) ast.Visitor {
	fmt.Printf(">> Handling *ast.Comment:\n\t")
	fmt.Printf("Line=%d, Text=%q\n", s.Position(x.Slash).Line, x.Text)

	return s
}

func (s *Sorcerer) astFile(x *ast.File) ast.Visitor {
	fmt.Printf(">> Handling *ast.File:\n\t")
	fmt.Printf("Line=%d, Text=%q\n", s.Position(x.Name.NamePos).Line, x.Name)

	return s
}

func (s *Sorcerer) astTypeSpec(x *ast.TypeSpec) ast.Visitor {
	fmt.Printf(">> Handling *ast.TypeSpec:\n\t")
	fmt.Printf(
		"Line=%d, Name=%q, Kind=%s, Decl=%v\n", s.Position(x.Name.NamePos).Line,
		x.Name, x.Name.Obj.Kind, x.Name.Obj.Decl,
	)

	return s
}

func (s *Sorcerer) astStructType(x *ast.StructType) ast.Visitor {
	fmt.Printf(">> Handling *ast.StructType:\n\t")
	fmt.Printf("Line=%d, Fields=%d\n", s.Position(x.Struct).Line, x.Fields.NumFields())
	return s
}

func (s *Sorcerer) astFuncDecl(x *ast.FuncDecl) ast.Visitor {
	fmt.Printf(">> Handling *ast.FuncDecl:\n\t")
	fmt.Printf("Line=%d, Name=%q\n", s.Position(x.Pos()).Line, x.Name)
	if fn := assembleMethod(x, s.fs); fn != "" {
		fmt.Printf(">>> METHOD RECEIVER: %s\n", fn)
	}
	return s
}

func (s *Sorcerer) AllMethodsOfStruct(x *ast.Package) {
	// Find all methods that are bound to the specific struct

	// Check that the receiver is actually the struct we want

}

func (s *Sorcerer) astFieldList(x *ast.FieldList) ast.Visitor {
	if x.NumFields() == 0 || s.Position(x.Opening).Line == 0 {
		return s
	}
	fmt.Printf(">> Handling *ast.FieldList:\n\t")
	fmt.Printf("Line=%d, Fields=%d\n", s.Position(x.Opening).Line, x.NumFields())
	return s
}

func (s *Sorcerer) astField(x *ast.Field) ast.Visitor {
	if len(x.Names) == 0 {
		return s
	}
	fmt.Printf(">> Handling *ast.Field:\n\t")
	fmt.Printf("Line=%d, Name=%s, Type=%v\n", s.Position(x.Names[0].Pos()).Line, x.Names[0].Name, x.Type)

	return s
}

type ParseFunc func(src []byte)

func (s *Sorcerer) Transfigure(filename string) {
	// Check empty file set, initialize if necessary.
	if s.fs == nil {
		s.fs = token.NewFileSet()
	}
	// Reset the *ast.File if it is already set.
	if s.f != nil {
		s.f = nil
	}
	// Then, we can call the parser.
	var err error
	s.f, err = parser.ParseFile(s.fs, filename, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	// Read in the source code of the file. This will be useful later in the
	// event in which we need to parse anything using the positional markers
	// found in the syntax tree.
	src, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	// Inspect the AST using the provided parsing function.
	// ast.Inspect(s.f, )
	_ = src
}

// ParseStructs is a function for isolating and inspecting
// the struct type AST nodes. It takes an ast.File, and
func ParseStructs(f *ast.File, src []byte) map[string]StructType {
	// instantiate return type
	ret := make(map[string]StructType)
	// parsing function
	parseFunc := func(n ast.Node) bool {
		switch t := n.(type) {
		// Keep an eye out for a type specification in the AST. This
		// is where we will get the name of our type from.
		case *ast.TypeSpec:
			// We have encountered a type spec node. We can now attempt
			// to assert if it is a struct type or not.
			e, correct := t.Type.(*ast.StructType)
			if !correct {
				// Not a struct, we will return true, so we can parse the
				// next node in the AST.
				return true
			}
			// Otherwise, our struct type assertion is good, so we can
			// instantiate a new StructType and start filling it out.
			structType := StructType{
				Name: t.Name.Name,
			}
			// Now we must range over our struct field set and read each
			// field name, get the type and look for any tags or comments.
			for _, f := range e.Fields.List {
				// First, instantiate new field type for this field. We
				// can add the field name, and derive the type using the
				// positional markers and reading from the original source.
				fieldType := FieldType{
					Name: f.Names[0].Name,
					Type: string(src[f.Type.Pos()-1 : f.Type.End()-1]),
				}
				// Check to see if there is a tag, and if so, add it.
				if f.Tag != nil {
					fieldType.Tag = f.Tag.Value
				}
				// Finally, append field type to struct type, and then
				// we continue on to the next field in the struct.
				structType.Fields = append(structType.Fields, fieldType)
			}
			// We are finished inspecting this struct. Add it to our struct map.
			ret[structType.Name] = structType
		}
		return true
	}
	// Inspect the AST using our parsing function.
	ast.Inspect(f, parseFunc)
	// Finally, return
	return ret
}

func (s *Sorcerer) Summon(f *ast.File) {
	ast.Walk(s, f)
}

func (s *Sorcerer) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	switch x := n.(type) {
	case *ast.Comment:
		return s.astComment(x)
	case *ast.File:
		return s.astFile(x)
	case *ast.TypeSpec:
		return s.astTypeSpec(x)
	case *ast.Field:
		return s.astField(x)
	case *ast.FieldList:
		return s.astFieldList(x)
	case *ast.StructType:
		return s.astStructType(x)
	case *ast.FuncDecl:
		fmt.Println(functionDef(x, s.fs))
		return s.astFuncDecl(x)
	}
	return s
}

func assembleMethod(x *ast.FuncDecl, fs *token.FileSet) string {
	// Validate that method is exported and has a receiver
	if x.Name.IsExported() && x.Recv != nil && len(x.Recv.List) == 1 {
		// Check that the receiver is actually the struct we want
		r, rok := x.Recv.List[0].Type.(*ast.StarExpr)
		if rok && r.X.(*ast.Ident).Name == "MyType" {
			return functionDef(x, fs)
		}
	}
	return ""
}

func functionDef(fun *ast.FuncDecl, fset *token.FileSet) string {
	name := fun.Name.Name
	params := make([]string, 0)
	for _, p := range fun.Type.Params.List {
		var typeNameBuf bytes.Buffer
		err := printer.Fprint(&typeNameBuf, fset, p.Type)
		if err != nil {
			log.Fatalf("failed printing %s", err)
		}
		names := make([]string, 0)
		for _, name := range p.Names {
			names = append(names, name.Name)
		}
		params = append(params, fmt.Sprintf("%s %s", strings.Join(names, ","), typeNameBuf.String()))
	}
	returns := make([]string, 0)
	if fun.Type.Results != nil {
		for _, r := range fun.Type.Results.List {
			var typeNameBuf bytes.Buffer
			err := printer.Fprint(&typeNameBuf, fset, r.Type)
			if err != nil {
				log.Fatalf("failed printing %s", err)
			}

			returns = append(returns, typeNameBuf.String())
		}
	}
	returnString := ""
	if len(returns) == 1 {
		returnString = returns[0]
	} else if len(returns) > 1 {
		returnString = fmt.Sprintf("(%s)", strings.Join(returns, ", "))
	}
	return fmt.Sprintf("%s(%s) %v", name, strings.Join(params, ", "), returnString)
}
