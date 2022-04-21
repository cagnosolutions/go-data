package sorcerer

import (
	"embed"
	_ "embed"
)

//go:embed templates/*.tmpl
var tmplFiles embed.FS
