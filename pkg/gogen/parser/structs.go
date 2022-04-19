package parser

import (
	"fmt"
	"go/ast"
	"go/token"
)

type commentType struct {
	value string
	line  int
}

func (ct commentType) String() string {
	return fmt.Sprintf("value=%q, line=%d", ct.value, ct.line)
}

type fieldType struct {
	name  string
	kind  string
	value string
	tag   string
	line  int
}

func (ft fieldType) String() string {
	return fmt.Sprintf(
		"name=%s, kind=%s, value=%s, tag=%s, line=%d",
		ft.name, ft.kind, ft.value, ft.tag, ft.line,
	)
}

type structType struct {
	comment commentType
	name    string
	fields  []fieldType
	node    *ast.StructType
	begLine int
	endLine int
}

func (st *structType) String() string {
	ss := fmt.Sprintf("structType{\n")
	ss += fmt.Sprintf("\tcomment: %s\n", st.comment)
	ss += fmt.Sprintf("\tname: %s\n", st.name)
	if st.fields != nil {
		ss += fmt.Sprintf("\tfields: []fieldType{\n")
		for i := range st.fields {
			ss += fmt.Sprintf("\t%s\n", st.fields[i])
		}
		ss += fmt.Sprintf("\t}\n")
	} else {
		ss += fmt.Sprintf("\tfields: <nil>\n")
	}
	ss += fmt.Sprintf("\tbegLine: %d\n", st.begLine)
	ss += fmt.Sprintf("\tendLine: %d\n", st.endLine)
	ss += fmt.Sprintf("}")
	return ss
}

// collectStructs collects and maps structType nodes to their positions
func collectStructs(node ast.Node, fset *token.FileSet) map[token.Pos]*structType {
	structs := make(map[token.Pos]*structType, 0)
	collectStructs := func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if t.Type == nil {
			return true
		}

		// structName := t.Name.Name

		x, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}
		structs[x.Pos()] = &structType{
			comment: commentType{},
			name:    t.Name.Name,
			fields:  nil,
			node:    x,
			begLine: fset.Position(x.Pos()).Line,
			endLine: fset.Position(x.End()).Line,
		}
		return true
	}
	ast.Inspect(node, collectStructs)
	return structs
}
