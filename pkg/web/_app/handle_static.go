package _app

import (
	"net/http"
)

type staticModel struct {
	staticPath string
	routes     *RouteMap
}

func NewStaticController() *staticModel {
	return &staticModel{
		staticPath: InternalResourcesPath("static/"),
		routes:     NewRouteMap(),
	}
}

func (s *staticModel) Routes() *RouteMap {
	s.routes.Handle("/favicon.ico", handleFavicon())
	s.routes.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticPath))))
	return s.routes
}

func handleFavicon() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/img/favicon.ico", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}
