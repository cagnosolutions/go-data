package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	flagDir    string
	flagFilter string
	flagCreate bool
	flagInsert bool
	flagUpdate bool
	flagDelete bool
	flagSelect bool
)

const (
	flagDirDefault = "."
	flagDirUsage   = "specify the directory to use"

	flagFilterDefault = "(.+)(_test.go)"
	flagFilterUsage   = "ignore anything that matches supplied regex pattern"

	flagCreateDefault = false
	flagCreateUsage   = "generate create statement"

	flagInsertDefault = false
	flagInsertUsage   = "generate insert statement"

	flagUpdateDefault = false
	flagUpdateUsage   = "generate update statement"

	flagDeleteDefault = false
	flagDeleteUsage   = "generate delete statement"

	flagSelectDefault = false
	flagSelectUsage   = "generate select statement"

	srcModeFlags = parser.AllErrors | parser.ParseComments
)

func init() {
	flag.StringVar(&flagDir, "dir", flagDirDefault, flagDirUsage)
	flag.StringVar(&flagFilter, "filter", flagFilterDefault, flagFilterUsage)
	flag.BoolVar(&flagCreate, "create", flagCreateDefault, flagCreateUsage)
	flag.BoolVar(&flagInsert, "insert", flagInsertDefault, flagInsertUsage)
	flag.BoolVar(&flagUpdate, "update", flagUpdateDefault, flagUpdateUsage)
	flag.BoolVar(&flagDelete, "delete", flagDeleteDefault, flagDeleteUsage)
	flag.BoolVar(&flagSelect, "select", flagSelectDefault, flagSelectUsage)
}

func reportErr(err error) {
	fmt.Fprintf(os.Stderr, "sqlwizard: %s\n", err)
}

func main() {
	// parse flags
	flag.Parse()
	// get current working directory
	dir, err := os.Getwd()
	if err != nil {
		reportErr(err)
	}
	dir = filepath.Join(dir, flagDir)
	// create new wizard
	wizard := &sqlWizard{
		fset:   new(token.FileSet),
		files:  make(map[string]*ast.File),
		filter: regexp.MustCompile(flagFilter),
	}
	// sanitize filepath
	fpath := filepath.ToSlash(dir)
	// parse files in path
	err = wizard.parseFiles(fpath)
	if err != nil {
		reportErr(err)
	}
	// generate create
	if flagCreate {
		err = wizard.generate(fpath, tmplCreate)
		if err != nil {
			reportErr(err)
		}
	}
	// generate insert
	if flagInsert {
		err = wizard.generate(fpath, tmplInsert)
		if err != nil {
			reportErr(err)
		}
	}
	// generate update
	if flagUpdate {
		err = wizard.generate(fpath, tmplUpdate)
		if err != nil {
			reportErr(err)
		}
	}
	// generate delete
	if flagDelete {
		err = wizard.generate(fpath, tmplDelete)
		if err != nil {
			reportErr(err)
		}
	}
	// generate select
	if flagSelect {
		err = wizard.generate(fpath, tmplSelect)
		if err != nil {
			reportErr(err)
		}
	}
	fmt.Fprintf(os.Stdout, "%s\n", wizard)
}

type sqlWizard struct {
	fset    *token.FileSet
	files   map[string]*ast.File
	filter  *regexp.Regexp
	structs []structType
}

func (s *sqlWizard) String() string {
	ss := fmt.Sprintf("wizard.files=%d\n", len(s.files))
	for name, _ := range s.files {
		ss += fmt.Sprintf("\tfile=%q\n", name)
	}
	ss += fmt.Sprintf("wizard.structs=%d\n", len(s.structs))
	for _, st := range s.structs {
		ss += fmt.Sprintf("\tstruct=%q\n", st.Name)
	}
	return ss
}

func (s *sqlWizard) parseFiles(dir string) error {
	// Get a list of files in dir
	list, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	// Range our list of files, excluding all directories
	// as well as anything that is not a .go file or does
	// not match our filter.
	for _, d := range list {
		// Skip over dirs, or any non go files.
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") {
			continue
		}
		// Apply filter (if we passed one in)
		if s.filter != nil {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if s.filter.MatchString(info.Name()) {
				continue
			}
		}
		filename := filepath.Join(dir, d.Name())
		f, err := parser.ParseFile(s.fset, filename, nil, srcModeFlags)
		if err != nil {
			return err
		}
		// Read in the source of the file
		src, err := os.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		log.Printf("adding file: %s\n", f.Name.Name)
		s.files[path.Base(filename)] = f
		s.collectStructs(src, f)
	}
	return nil
}

type fieldType struct {
	Name string
	Kind string
	Tag  string
}

func (ft fieldType) NameLower() string {
	return strings.ToLower(ft.Name)
}

func (ft fieldType) NameShort() string {
	return string(strings.ToLower(ft.Name)[0])
}

func (ft fieldType) SQLType() string {
	switch {
	case ft.Kind == "string":
		return "TEXT"
	case strings.HasPrefix(ft.Kind, "int"):
		return "INTEGER"
	case strings.HasPrefix(ft.Kind, "uint"):
		return "INTEGER"
	case strings.HasPrefix(ft.Kind, "float"):
		return "REAL"
	case strings.HasPrefix(ft.Kind, "time"):
		return "TIMESTAMP"
	case ft.Kind == "bool":
		return "BOOLEAN"
	default:
		return "JSON"
	}
}

type structType struct {
	Package string
	Name    string
	Fields  []fieldType
}

func (st structType) String() string {
	var sb strings.Builder
	sb.Grow(len(st.Name) + (len(st.Fields) * 32))
	sb.WriteString(st.Name)
	sb.WriteString("\n")
	for _, f := range st.Fields {
		sb.WriteString("\t-")
		sb.WriteString(f.Name)
		sb.WriteString(", ")
		sb.WriteString(f.Kind)
		sb.WriteString(" (")
		sb.WriteString(f.Tag)
		sb.WriteString(")\n")
	}
	return sb.String()
}

func (st structType) NameLower() string {
	return strings.ToLower(st.Name)
}

func (st structType) NameShort() string {
	return string(strings.ToLower(st.Name)[0])
}

func (st structType) GetFields() []fieldType {
	return st.Fields
}

// collectStructs attempts to collect all the struct types located the supplied
// *File. It parses them into a map of struct types.
func (s *sqlWizard) collectStructs(src []byte, f *ast.File) {
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
			// Get the package name offset
			// beg := s.fset.Position(f.Name.Pos())
			// end := s.fset.Position(f.Name.End())
			// Otherwise, our struct type assertion is good, so we can
			// instantiate a new StructType and start filling it out.
			st := structType{
				Package: f.Name.String(),
				Name:    t.Name.Name,
			}
			// Now we must range over our struct field set and read each
			// field name, get the type and look for any tags or comments.
			for _, field := range e.Fields.List {
				// First, instantiate new field type for this field. We
				// can add the field name, and derive the type using the
				// positional markers and reading from the original source.
				beg := s.fset.Position(field.Type.Pos())
				end := s.fset.Position(field.Type.End())
				ft := fieldType{
					Name: field.Names[0].Name,
					Kind: string(src[beg.Offset:end.Offset]),
				}
				// Check to see if there is a tag, and if so, add it.
				if field.Tag != nil {
					ft.Tag = field.Tag.Value
				}
				// Finally, append field type to struct type, and then
				// we continue on to the next field in the struct.
				st.Fields = append(st.Fields, ft)
			}
			// We are finished inspecting this struct. Add it to our struct map.
			s.structs = append(s.structs, st)
			return false
		}
		return true
	}
	// Inspect the AST using our parsing function.
	ast.Inspect(f, parseFunc)
	// Finally, return
	return
}

func (s *sqlWizard) generate(path string, tmpl *template.Template) error {
	for _, st := range s.structs {
		name := fmt.Sprintf("%s_gen.go", strings.ToLower(st.Name))
		err := os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
		fp, err := os.Create(filepath.Join(path, name))
		if err != nil {
			return err
		}
		err = tmpl.Execute(fp, st)
		if err != nil {
			fp.Close()
			return err
		}
		err = fp.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

var createStmt = `// AUTO GENERATED; DO NOT EDIT! (CREATE)

package {{ .Package }}

import (
	"database/sql"
	"log"
)

const create{{ .Name }}Stmt = {{ tick }}CREATE TABLE IF NOT EXISTS '{{ .NameLower }}' (
	{{- $fields := .GetFields }}
	{{- range $i, $field := $fields }}
	'{{ $field.NameLower }}' {{ $field.SQLType }}
		{{- if isPK $field }} PRIMARY KEY{{ end }}
		{{- if notEnd $i $fields }},{{ end }}
	{{- end }}
);{{ tick }}

func Create{{ .Name }}Table(db *sql.DB) error {
	// prepare to protect against sql-injection
	stmt, err := db.Prepare(create{{ .Name }}Stmt)
	if err != nil {
		log.Fatalf("preparing(%q): %s", create{{ .Name }}Stmt, err)
	}
	// execute
	_, err = stmt.Exec()
	if err != nil {
		log.Fatalf("executing prepared statement: %s", err)
	}
	return nil
}
`

var insertStmt = `// AUTO GENERATED; DO NOT EDIT! (INSERT)

package {{ .Package }}

import (
	"database/sql"
	"log"
)
{{ $fields := .GetFields }}
const insert{{ .Name }}Stmt = {{ tick }}INSERT INTO '{{ .NameLower }}' (
	{{- range $i, $field := $fields }}
	'{{- $field.NameLower }}'{{- if notEnd $i $fields }},{{- end }}
	{{- end }}
) VALUES ({{- range $i, $field := $fields }}?{{- if notEnd $i $fields }},{{ end }}{{- end }});{{ tick }}

func Insert{{ .Name }}Table(db *sql.DB, data ...interface{}) error {
	// prepare to protect against sql-injection
	stmt, err := db.Prepare(insert{{ .Name }}Stmt)
	if err != nil {
		log.Fatalf("preparing(%q): %s", insert{{ .Name }}Stmt, err)
	}
	// execute
	_, err = stmt.Exec(data...)
	if err != nil {
		log.Fatalf("executing prepared statement: %s", err)
	}
	return nil
}
`

var updateStmt = `// AUTO GENERATED; DO NOT EDIT! (UPDATE)

package {{ .Package }}

import (
	"database/sql"
	"log"
)

const update{{ .Name }}Stmt = {{ tick }}
{{ tick }}
`

var deleteStmt = `// AUTO GENERATED; DO NOT EDIT! (DELETE)

package {{ .Package }}

import (
	"database/sql"
	"log"
)

const delete{{ .Name }}Stmt = {{ tick }}
{{ tick }}
`

var selectStmt = `// AUTO GENERATED; DO NOT EDIT! (SELECT)

package {{ .Package }}

import (
	"database/sql"
	"log"
)

const select{{ .Name }}Stmt = {{ tick }}
{{ tick }}
`

var fm = template.FuncMap{
	"notEnd": func(index int, set []fieldType) bool { return index < len(set)-1 },
	"isPK":   func(f fieldType) bool { return strings.Contains(strings.ToLower(f.Tag), "pk") },
	"tick":   func() string { return "`" },
}

var (
	tmplCreate = template.Must(template.New("create").Funcs(fm).Parse(createStmt))
	tmplInsert = template.Must(template.New("insert").Funcs(fm).Parse(insertStmt))
	tmplUpdate = template.Must(template.New("update").Funcs(fm).Parse(updateStmt))
	tmplDelete = template.Must(template.New("delete").Funcs(fm).Parse(deleteStmt))
	tmplSelect = template.Must(template.New("select").Funcs(fm).Parse(selectStmt))
)
