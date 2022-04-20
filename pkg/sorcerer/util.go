package sorcerer

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
)

// isDirectory determines if a file represented
// by `path` is a directory or not
func isDirectory(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
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
