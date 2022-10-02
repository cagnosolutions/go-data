package admin

import (
	"embed"
	"html/template"
	"path/filepath"
)

const adminRoot = "resources/templates"

//go:embed resources/templates/*
var adminTmplFiles embed.FS

var AdminTemplates *template.Template

func init() {

	baseTmpls := filepath.ToSlash(filepath.Join(adminRoot, "*.gohtml"))
	stubTmpls := filepath.ToSlash(filepath.Join(adminRoot, "stubs/*.gohtml"))

	AdminTemplates = template.Must(template.ParseFS(adminTmplFiles, baseTmpls))
	AdminTemplates = template.Must(AdminTemplates.ParseFS(adminTmplFiles, stubTmpls))
}
