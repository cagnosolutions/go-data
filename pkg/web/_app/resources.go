package _app

import (
	"embed"
	"html/template"
	"path/filepath"
)

//go:embed resources/* resources/static/* resources/templates/*
var InternalResources embed.FS

func InternalResourcesPath(join string) string {
	return filepath.Join("app/resources", join)
}

func LoadTemplates() *template.Template {
	tmpls := template.Must(template.ParseFS(InternalResources, "resources/templates/*.html"))
	tmpls = template.Must(tmpls.ParseFS(InternalResources, "resources/templates/stubs/*.html"))
	return tmpls
}
