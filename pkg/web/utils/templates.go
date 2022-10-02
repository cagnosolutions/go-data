package utils

import (
	"html/template"
	"net/http"
)

type TemplateCacheConfig struct {
	BasePattern string
	StubPattern string
	FuncMap     template.FuncMap
}

type TemplateCache struct {
	*TemplateCacheConfig
	t *template.Template
}

func initTemplates(pattern string, funcMap template.FuncMap) *template.Template {
	return template.Must(template.New("*").Funcs(funcMap).ParseGlob(pattern))
}

func NewTemplateCache(conf *TemplateCacheConfig) *TemplateCache {
	if conf.FuncMap == nil {
		conf.FuncMap = template.FuncMap{}
	}
	tc := &TemplateCache{
		TemplateCacheConfig: conf,
	}
	tc.t = initTemplates(tc.BasePattern, tc.FuncMap)
	if conf.StubPattern != "" {
		tc.ParseGlob(conf.StubPattern)
	}
	return tc
}

func (tc *TemplateCache) ParseGlob(pattern string) {
	t, err := tc.t.Funcs(tc.FuncMap).ParseGlob(pattern)
	if err != nil {
		panic(err)
	}
	tc.t = t
}

func (tc *TemplateCache) ExecuteTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	err := tc.t.ExecuteTemplate(w, name, data)
	if err != nil {
		// code := http.StatusExpectationFailed
		// http.Error(w, http.StatusText(code), code)
		http.RedirectHandler("/error/417", http.StatusFound)
		return
	}
}

func (tc *TemplateCache) DefinedTemplates() string {
	return tc.t.DefinedTemplates()
}

func (tc *TemplateCache) Lookup(name string) *template.Template {
	return tc.t.Lookup(name)
}
