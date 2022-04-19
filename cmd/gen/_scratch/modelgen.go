package _scratch

import (
	"os"
	"text/template"
)

func foo() {
	modelTemplate, err := os.ReadFile("templates/models.gotmpl")
	if err != nil {
		panic(err)
	}
	tmpl := template.Must(template.New("model").Parse(string(modelTemplate)))
	// _generateCode(tmpl, "models.go")
	_ = tmpl
}
