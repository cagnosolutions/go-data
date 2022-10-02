package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var ErrBadCode = errors.New("bad code")

// ErrorResponse represents an HTTP error response for a REST service
type ErrorResponse struct {
	Code    int            `json:"code"`    // http status code
	Status  string         `json:"status"`  // http status text
	Message string         `json:"message"` // human-readable representation of the error
	Details map[string]any `json:"details"` // any other details
}

func NewErrorResponse(code int, details JSON) (*ErrorResponse, error) {
	if !Is4xxCode(code) && !Is5xxCode(code) {
		return nil, ErrBadCode
	}
	return &ErrorResponse{
		Code:    code,
		Status:  http.StatusText(code),
		Message: httpCodes[code][1],
		Details: details,
	}, nil
}

func (er *ErrorResponse) WriteTo(w io.Writer) (int64, error) {
	b, err := json.Marshal(er)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

func (er *ErrorResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(er)
}

func (er *ErrorResponse) UnmarshalJSON(p []byte) error {
	var r ErrorResponse
	err := json.Unmarshal(p, &r)
	if err != nil {
		return err
	}
	*er = r
	return nil
}
