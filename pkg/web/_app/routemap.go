package _app

import (
	"net/http"

	"github.com/cagnosolutions/webapp/pkg/webapp/servemux"
)

type RouteMap struct {
	routes []servemux.Route
}

func NewRouteMap() *RouteMap {
	return &RouteMap{
		routes: make([]servemux.Route, 0),
	}
}

func (rm *RouteMap) Handle(pattern string, handler http.Handler) {
	rm.routes = append(
		rm.routes, servemux.Route{
			Method:  "*",
			Pattern: pattern,
			Handler: handler,
		},
	)
}

func (rm *RouteMap) Get(pattern string, handler http.Handler) {
	rm.routes = append(
		rm.routes, servemux.Route{
			Method:  http.MethodGet,
			Pattern: pattern,
			Handler: handler,
		},
	)
}

func (rm *RouteMap) Post(pattern string, handler http.Handler) {
	rm.routes = append(
		rm.routes, servemux.Route{
			Method:  http.MethodPost,
			Pattern: pattern,
			Handler: handler,
		},
	)
}

func (rm *RouteMap) AllRoutes() []servemux.Route {
	return rm.routes
}

func (rm *RouteMap) Range(fn func(route servemux.Route) bool) {
	for i := range rm.routes {
		if !fn(rm.routes[i]) {
			break
		}
	}
}

func (rm *RouteMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range rm.routes {
		switch route.Pattern {
		case r.URL.Path:
			if route.Method == "*" {
				route.Handler.ServeHTTP(w, r)
				return
			}
			if route.Method == http.MethodGet {
				route.Handler.ServeHTTP(w, r)
				return
			}
			if route.Method == http.MethodPost {
				route.Handler.ServeHTTP(w, r)
				return
			}
		}
	}
}
