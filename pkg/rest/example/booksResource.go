package example

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/cagnosolutions/go-data/pkg/rest"
)

type BooksResource struct {
	Books
}

var DefaultBookResource = &BooksResource{
	Books: make([]Book, 0),
}

func (b *BooksResource) AppendBook(book Book) {
	b.Books = append(b.Books, book)
}

func (b *BooksResource) GetAll() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// return all books
		rest.WriteAsJSON(w, b.Books)
	}
	return http.HandlerFunc(fn)
}

func (b *BooksResource) GetOne(id string) http.Handler {
	// search book by id
	i, found := b.Books.searchByID(id)
	if !found {
		return http.NotFoundHandler()
	}
	// isolate book
	book := b.Books[i]
	fn := func(w http.ResponseWriter, r *http.Request) {
		// return book
		rest.WriteAsJSON(w, book)
	}
	return http.HandlerFunc(fn)
}

func (b *BooksResource) Add(req *http.Request) http.Handler {
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
		rest.WriteAsJSON(w, b.Books)
	}
	return http.HandlerFunc(fn)
}

func (b *BooksResource) Set(req *http.Request, id string) http.Handler {
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
		rest.WriteAsJSON(w, b.Books)
	}
	return http.HandlerFunc(fn)
}

func (b *BooksResource) Del(id string) http.Handler {
	// delete book
	delBookByID(&b.Books, id)
	// now we can start handling...
	fn := func(w http.ResponseWriter, r *http.Request) {
		// return books
		rest.WriteAsJSON(w, b.Books)
	}
	return http.HandlerFunc(fn)
}
