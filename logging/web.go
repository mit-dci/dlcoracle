package logging

import "net/http"

func WebLoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Info.Printf("[%s] - %s", r.Method, r.URL.String())
		h.ServeHTTP(w, r)
	})
}
