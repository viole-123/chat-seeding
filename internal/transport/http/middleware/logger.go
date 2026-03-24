package middleware

import "net/http"

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: add structured logging
		next.ServeHTTP(w, r)
	})
}
