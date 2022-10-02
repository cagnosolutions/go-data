package middleware

import "net/http"

func HandleStatic(prefix, path string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(path)))
}
