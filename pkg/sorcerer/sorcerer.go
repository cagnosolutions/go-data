package sorcerer

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	srcDir = 1 << iota
	srcFile
	srcExpr

	srcModeFlags = parser.AllErrors | parser.ParseComments
)

var (
	ErrPkgNotSpecified = errors.New("package name not specified")
	ErrPkgNotFound     = errors.New("package not found")
	ErrFileNotFound    = errors.New("file not found")
)

// Sorcerer is a source code wizard. It parses go source code and
// gives you access to things like package name, comments, structs,
// methods, functions and clear methods to access, transform and
// generate new output from any of this data.
type Sorcerer struct {
	fs        *token.FileSet          // A FileSet represents a set of source files.
	Files     map[string]*File        // Contains Go source files and raw source.
	Structs   map[string]StructType   // Contains the structs in the source file.
	Functions map[string]FunctionType // Contains the functions in the source file.
}

// NewSorcerer instantiates a new source code wizard that can do magical things with
// go source code. It uses the provided source and tries to locate the file or files,
// or directly parse it as a raw statement. If there is an error it will be returned.
func NewSorcerer() *Sorcerer {
	// Create and return new *Sorcerer, so we can start our wizardry.
	return &Sorcerer{
		fs:        token.NewFileSet(),
		Files:     make(map[string]*File),
		Structs:   make(map[string]StructType),
		Functions: make(map[string]FunctionType),
	}
}

// GetFile loads and returns the source file along with the raw source code. If
// the file cannot be found, or if there is an error parsing the raw source an
// error is returned.
func (s *Sorcerer) GetFile(filename string) (*File, error) {
	// Check for file
	f, found := s.Files[filename]
	if !found {
		return nil, ErrFileNotFound
	}
	// Found the file, return it.
	return f, nil
}

// ParseExpression parses a source string specified by source, as an
// expression. If successful, the parsed expression is added to the files
// map with "expr" set as the key for that file entry.
func (s *Sorcerer) ParseExpression(source string) error {
	// Attempt to parse the source provided as an expression.
	astf, err := parser.ParseFile(s.fs, "", source, srcModeFlags)
	if err != nil {
		return err
	}
	// Create a new file instance.
	file := &File{
		Name:   "expr",
		File:   astf,
		Source: []byte(source),
	}
	// Run our struct collector and add collected
	// structs to our structs map.
	for _, st := range file.collectStructs() {
		s.Structs[st.Name] = st
	}
	// Add the parsed file to the files map.
	s.Files["expr"] = file
	// return
	return nil
}

// ParseFile parses a single source code file specified by filename. A
// successfully parsed file is added to the files map.
func (s *Sorcerer) ParseFile(filename string) error {
	// Attempt to parse the source provided as a single file.
	astf, err := parser.ParseFile(s.fs, filename, nil, srcModeFlags)
	if err != nil {
		return err
	}
	// Read in the source code of the file. The source gets added to the
	// *File in case we need to parse anything using the positional markers
	// found in the syntax tree.
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	// Create a new file instance.
	file := &File{
		Name:   filename,
		File:   astf,
		Source: src,
	}
	// Run our struct collector and add collected
	// structs to our structs map.
	for _, st := range file.collectStructs() {
		s.Structs[st.Name] = st
	}
	// Add the parsed file to the files map.
	s.Files[filename] = file
	// return
	return nil
}

// ParseDir parses a set of files in the directory specified by dir. Only
// entries passing through the filter (and ending in ".go") are considered.
// If the filter is nil, it is skipped, and it attempts to parse all files
// ending in ".go" in the directory specified by dir. All files that are
// successfully parsed are added to the files map.
func (s *Sorcerer) ParseDir(dir string, filter func(fs.FileInfo) bool) error {
	// Attempt to parse the source files provided by dir and filtered
	// through the filter, if one exists.
	pkgs, err := parser.ParseDir(s.fs, dir, filter, srcModeFlags)
	if err != nil {
		return err
	}
	// Attempt to locate package using dir specified.
	astp, found := pkgs[dir]
	if !found {
		return ErrPkgNotFound
	}
	// Parsing was successful, and we now have an *ast.Package type. We must
	// now range the files within this package and add them to our files map.
	for filename, astf := range astp.Files {
		// Read in the source code of the file. The source gets added to the
		// *File in case we need to parse anything using the positional markers
		// found in the syntax tree.
		src, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		// Create a new file instance.
		file := &File{
			Name:   filepath.ToSlash(filename),
			File:   astf,
			Source: src,
		}
		// Run our struct collector and add collected
		// structs to our structs map.
		for _, st := range file.collectStructs() {
			s.Structs[st.Name] = st
		}
		// Add the parsed file to the files map.
		s.Files[filepath.Base(filename)] = file
	}
	// And finally, we return
	return nil
}

// collectStructs attempts to collect all the struct types located the supplied
// *File. It parses them into a map of struct types.
func collectStructs(file *File) []StructType {
	// instantiate return type
	var ret []StructType
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
			for _, field := range e.Fields.List {
				// First, instantiate new field type for this field. We
				// can add the field name, and derive the type using the
				// positional markers and reading from the original source.
				fieldType := FieldType{
					Name: field.Names[0].Name,
					Type: string(file.Source[field.Type.Pos()-1 : field.Type.End()-1]),
				}
				// Check to see if there is a tag, and if so, add it.
				if field.Tag != nil {
					fieldType.Tag = field.Tag.Value
				}
				// Finally, append field type to struct type, and then
				// we continue on to the next field in the struct.
				structType.Fields = append(structType.Fields, fieldType)
			}
			// We are finished inspecting this struct. Add it to our struct map.
			ret = append(ret, structType)
			return false
		}
		return true
	}
	// Inspect the AST using our parsing function.
	ast.Inspect(file.File, parseFunc)
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
