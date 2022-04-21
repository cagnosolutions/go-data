package sorcerer

import (
	"fmt"
	"go/ast"
)

type File struct {
	Name   string
	File   *ast.File
	Source []byte
}

func (f *File) String() string {
	return fmt.Sprintf(
		"name=%q, pkg=%q, bytes=%d, decls=%d, scope=%v",
		f.Name,
		f.File.Name.Name,
		len(f.Source),
		len(f.File.Decls),
		f.File.Scope.Objects,
	)
}

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
