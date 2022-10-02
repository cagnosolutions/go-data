package middleware

import (
	"fmt"
	"mime"
	"net/http"
	"sort"
	"strings"
)

func HandleMetrics(title string, ss []string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var data []string
		data = append(data, fmt.Sprintf("<h3>%s</h3>", title))
		sort.Strings(ss)
		w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
		_, err := fmt.Fprintf(w, strings.Join(data, "<br>"))
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}
		return
	}
	return http.HandlerFunc(fn)
}
