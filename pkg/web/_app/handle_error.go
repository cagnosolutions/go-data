package _app

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/cagnosolutions/go-data/pkg/web"
)

// errorModel is a default error handling model
type errorModel struct {
	errTmpl *template.Template
	routes  *RouteMap
}

// NewErrorController is a default generic error handling controller
func NewErrorController() *errorModel {
	return &errorModel{
		errTmpl: LoadTemplates().Lookup("error.html"),
		routes:  NewRouteMap(),
	}
}

func (e *errorModel) Routes() *RouteMap {
	e.routes.Handle("/error", http.RedirectHandler("/error/206", http.StatusPartialContent))
	e.routes.Handle("/error/", e.handleErrorWithCode())
	return e.routes
}

func (e *errorModel) handleErrorWithCode() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// reject post request
		if r.Method == http.MethodPost {
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
			return
		}
		// handle get request
		if r.Method == http.MethodGet {
			p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(p) > 1 {
				code, err := strconv.Atoi(p[1])
				if err != nil {
					code := http.StatusExpectationFailed
					http.Error(w, http.StatusText(code), code)
					return
				}
				err = e.errTmpl.Execute(
					w, struct {
						ErrorCode     int
						ErrorText     string
						ErrorTextLong string
					}{
						ErrorCode:     code,
						ErrorText:     http.StatusText(code),
						ErrorTextLong: web.HTTPCodesLongFormat[code],
					},
				)
				if err != nil {
					code := http.StatusExpectationFailed
					http.Error(w, http.StatusText(code), code)
					return
				}
			}
		}
	}
	return http.HandlerFunc(fn)
}
