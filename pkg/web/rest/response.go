package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

var ErrEmptyRequest = errors.New("got empty request")

// Response represents an HTTP response object
type Response struct {
	Path          string    `json:"path"`           // path in request
	Version       string    `json:"version"`        // http or api version
	Code          int       `json:"code"`           // http status code
	Status        string    `json:"status"`         // http status text
	ContentType   string    `json:"content_type"`   // content type
	LastModified  time.Time `json:"last_modified"`  // last modified
	ContentLength int64     `json:"content_length"` // content length in bytes
	Allow         string    `json:"allow"`          // allow header
}

// NewResponse creates and returns a new Response
func NewResponse(r *http.Request, code int) (*Response, error) {
	if r == nil {
		return nil, ErrEmptyRequest
	}
	if !IsValidCode(code) {
		return nil, ErrBadCode
	}
	return &Response{
		Path:          r.RequestURI,
		Version:       r.Response.Proto,
		Code:          code,
		Status:        http.StatusText(code),
		ContentType:   r.Header.Get("Content-Type"),
		LastModified:  time.Now(),
		ContentLength: r.ContentLength,
		Allow:         r.Header.Get("Allow"),
	}, nil
}

func (r *Response) WriteTo(w io.Writer) (int64, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}
