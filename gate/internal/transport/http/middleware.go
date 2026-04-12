package http

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// RecoverMiddleware catches any panics, logs the stack trace to our registered output (which includes file), and responds with 500.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.Printf("[PANIC RECOVERED] url=%s method=%s err=%v\n%s", r.URL.Path, r.Method, rvr, debug.Stack())
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"message": "Internal Server Error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
