package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"sort"
)

// resource is an internal representation wrapping a user supplied ResourceHandler
type resource struct {
	name string
	path string
	col  ResourceCollection
}

// checkID returns a boolean reporting true if a resource id can be identified
func (re *resource) checkID(uri string) bool {
	if len(uri) > 0 && uri[len(uri)-1] == '/' {
		uri = uri[:len(uri)-1]
	}
	i := len(uri) - 1
	for i >= 0 && uri[i] != '/' {
		i--
	}
	return uri[i+1:] != "" && uri[i+1:] != re.name
}

func logRequest(r *http.Request, msg string) {
	log.Printf("method=%q, path=%q, msg=%q\n", r.Method, r.RequestURI, msg)
}

func (re *resource) getAll() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// return all items
		re.col.Filter()
		WriteAsJSON(w, b.Books)
	}
	return http.HandlerFunc(fn)
}

func (re *resource) getOne(id string) http.Handler {
	// search book by id
	i, found := b.Books.searchByID(id)
	if !found {
		return http.NotFoundHandler()
	}
	// isolate book
	book := b.Books[i]
	fn := func(w http.ResponseWriter, r *http.Request) {
		// return book
		WriteAsJSON(w, book)
	}
	return http.HandlerFunc(fn)
}

func (re *resource) addOne(req *http.Request) http.Handler {
	// decode body into new book
	var book Book
	err := json.NewDecoder(req.Body).Decode(&book)
	// add book to set if no error
	if err == nil {
		b.Books = append(b.Books, book)
		// sort
		sort.Stable(b.Books)
	}
	// now we can start handing...
	fn := func(w http.ResponseWriter, r *http.Request) {
		// if err exists
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		// return books
		WriteAsJSON(w, b.Books)
	}
	return http.HandlerFunc(fn)
}

func (re *resource) setOne(req *http.Request, id string) http.Handler {
	// decode body into new book
	var book Book
	err := json.NewDecoder(req.Body).Decode(&book)
	// if no error
	if err == nil {
		// delete "old" book (for "update")
		delBookByID(&b.Books, id)
		// and add new book (to complete "update")
		b.Books = append(b.Books, book)
		// sort
		sort.Stable(b.Books)
	}
	// now we can start handling...
	fn := func(w http.ResponseWriter, r *http.Request) {
		// if err exists
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		// return books
		WriteAsJSON(w, b.Books)
	}
	return http.HandlerFunc(fn)
}

func (re *resource) delOne(id string) http.Handler {
	// delete book
	delBookByID(&b.Books, id)
	// now we can start handling...
	fn := func(w http.ResponseWriter, r *http.Request) {

		WriteAsJSON(w, b.Books)
	}
	return http.HandlerFunc(fn)
}

func (re *resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var h http.Handler
	hasID := re.checkID(r.RequestURI)
	var id string
	if hasID {
		id = path.Base(r.RequestURI)
	}
	if hasID {
		switch r.Method {
		case http.MethodGet:
			logRequest(r, "get one")
			h = re.getOne(id)
			goto serve
		case http.MethodPut:
			logRequest(r, "update one")
			h = re.setOne(r, id)
			goto serve
		case http.MethodDelete:
			logRequest(r, "delete one")
			h = re.delOne(id)
			goto serve
		default:
			logRequest(r, "bad request with id")
			h = http.NotFoundHandler()
			goto serve
		}
	}
	switch r.Method {
	case http.MethodGet:
		logRequest(r, "get all")
		h = re.getAll()
		goto serve
	case http.MethodPost:
		logRequest(r, "add one")
		h = re.addOne(r)
		goto serve
	default:
		// LogRequest(r, "bad request")
		h = http.NotFoundHandler()
		goto serve
	}
serve:
	h.ServeHTTP(w, r)
}
