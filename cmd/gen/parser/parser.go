package parser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

var re = regexp.MustCompile(`data:"(\d|\w+)(,omitempty)?"`)

const Object = "Object"

type ModelField struct {
	Field       string
	Type        string
	Description string
	Tag         string
}

type DataModel struct {
	Name         string
	Fields       []ModelField
	FieldsByName map[string]ModelField
}

func NewDataModel(name string) *DataModel {
	return &DataModel{
		Name:         name,
		Fields:       make([]ModelField, 0),
		FieldsByName: make(map[string]ModelField),
	}
}

type Parser struct {
	DataModels        []*DataModel
	DataModelsByName  map[string]*DataModel
	file              *ast.File
	dataModelTemplate *template.Template
}

func NewParser(filename string) (*Parser, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	dataModelTemplate, err := ioutil.ReadFile("tmpls/dataModelTemplate.go.tmpl")
	if err != nil {
		return nil, err
	}

	return &Parser{
		file:              f,
		DataModels:        make([]*DataModel, 0),
		DataModelsByName:  make(map[string]*DataModel),
		dataModelTemplate: template.Must(template.New("dataModel").Parse(string(dataModelTemplate))),
	}, nil
}

func (p *Parser) Parse() error {
	ast.Inspect(
		p.file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if err := p.parseType(x); err != nil {
					return false
				}
			case *ast.FuncDecl:
				p.parseFunction(x)
			}
			return true
		},
	)
	return p.generate()
}

func (p *Parser) generate() error {
	err := p.generateCode(p.dataModelTemplate, "model_ds.go")
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) generateCode(tmpl *template.Template, fn string) error {
	fmt.Printf("Generating %s\n", fn)
	buf := bytes.NewBuffer([]byte{})
	err := tmpl.Execute(buf, p)
	if err != nil {
		return err
	}
	res, err := imports.Process(fn, buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fn, res, 0666)
}

func (p *Parser) parseType(st *ast.TypeSpec) error {
	if strings.HasSuffix(st.Name.Name, "Model") {
		model := strings.Replace(st.Name.Name, "Model", "", -1)
		p.addParameter(model, st.Type.(*ast.StructType))
	}
	return nil
}

func (p *Parser) addParameter(name string, st *ast.StructType) {
	for _, field := range st.Fields.List {
		field := ModelField{
			Field:       field.Names[0].Name,
			Description: field.Doc.Text(),
			Tag:         parseTag(field.Tag.Value),
			Type:        mapFieldType(field.Type),
		}
		if p.DataModelsByName[name] == nil {
			p.DataModelsByName[name] = NewDataModel(name)
		}
		p.DataModelsByName[name].Fields = append(p.DataModelsByName[name].Fields, field)
		p.DataModelsByName[name].FieldsByName[field.Field] = field
	}
}

func parseTag(tag string) string {
	match := re.FindStringSubmatch(tag)
	return match[1]
}

func mapFieldType(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		ident, ok := x.X.(*ast.Ident)
		if ok {
			return ident.Name
		}
		return Object
	case *ast.SelectorExpr:
		name := fmt.Sprintf("%v.%s", x.X, x.Sel.Name)
		switch name {
		case "globalid.ID":
			return "UUID"
		case "model.ReactionType":
			return "string"
		case "model.CardsResponse", "model.CardResponse", "model.Draft":
			return Object
		}
		return name
	case *ast.ArrayType:
		return "Array"
	default:
		panic(fmt.Sprintf("Unmapped type %T %v", x, x))
	}
}

func (p *Parser) parseFunction(fd *ast.FuncDecl) {
	if fd.Recv == nil {
		return
	}
	if recv, ok := fd.Recv.List[0].Type.(*ast.StarExpr); ok {
		if ident, ok := recv.X.(*ast.Ident); ok {
			name := fd.Name.Name
			description := fd.Doc.Text()
			firstChar := string(name[0])
			if ident.Name == "model" && firstChar == strings.ToUpper(firstChar) {
				p.AddDataModel(name, description)
			}
		}
	}
}

func (p *Parser) AddDataModel(name, description string) {
	model := p.DataModelsByName[name]
	if model == nil {
		model = NewDataModel(name)
	}
	// model.Description = enhanceDescription(description, name)

	p.DataModelsByName[name] = model
	p.DataModels = append(p.DataModels, model)
}
