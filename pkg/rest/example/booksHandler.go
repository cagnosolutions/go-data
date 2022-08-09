package example

import (
	"fmt"
	"net/http"
	"path"
)

type BooksHandler struct {
	books Books
}

func (b *BooksHandler) ReturnAll(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%q, return all (%q)\n", b, r.RequestURI)
}

func (b *BooksHandler) ReturnOne(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.RequestURI)
	fmt.Fprintf(w, "%q, return one (%q), id=%v\n", b, r.RequestURI, id)
}

func (b *BooksHandler) InsertOne(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%q, insert one\n", b)
}

func (b *BooksHandler) UpdateOne(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%q, update one\n", b)
}

func (b *BooksHandler) DeleteOne(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%q, delete one\n", b)
}

func (b *BooksHandler) String() string {
	return "BooksHandler"
}
