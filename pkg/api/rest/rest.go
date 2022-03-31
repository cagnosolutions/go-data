package rest

import (
	"net/http"
)

// RESTApiServer can be used to create API's
type RESTApiServer struct {
	base string
	h    *http.ServeMux
}

// NewAPIServer creates and returns a new server instance
func NewAPIServer(base string) *RESTApiServer {
	srv := &RESTApiServer{
		base: clean(base),
		h:    http.NewServeMux(),
	}
	srv.h.Handle("/", http.RedirectHandler(srv.base, http.StatusSeeOther))
	return srv
}

func (srv *RESTApiServer) RegisterResourceHandler(name string, re ResourceHandler) {
	r := &resourceHandler{
		name: name,
		path: clean(srv.base + name),
		re:   re,
	}
	srv.h.Handle(r.path, r)
}

func (srv *RESTApiServer) RegisterResource(name string, re Resource) {
	r := &resource{
		name:     name,
		path:     clean(srv.base + name),
		Resource: re,
	}
	srv.h.Handle(r.path, r)
}

// ServeHTTP is the APIServer's default handler
func (srv *RESTApiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// lookup resource handler
	h, pat := srv.h.Handler(r)
	// do something with pattern if we need to
	_ = pat
	// call handler
	h.ServeHTTP(w, r)
	return
}
