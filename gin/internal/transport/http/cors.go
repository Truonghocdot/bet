package http

import (
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	corsOnce           sync.Once
	corsAllowedOrigins []string
	corsAllowAll       bool
)

func loadCORSConfig() {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		corsAllowAll = true
		corsAllowedOrigins = nil
		return
	}

	corsAllowAll = false
	items := strings.Split(raw, ",")
	origins := make([]string, 0, len(items))
	for _, item := range items {
		origin := strings.TrimSpace(item)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}
	corsAllowedOrigins = origins
}

func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	if corsAllowAll {
		return true
	}

	for _, allowed := range corsAllowedOrigins {
		if allowed == origin {
			return true
		}
		if strings.HasPrefix(allowed, "*.") {
			suffix := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(origin, suffix) {
				return true
			}
		}
	}

	return false
}

func withCORS(next http.Handler) http.Handler {
	corsOnce.Do(loadCORSConfig)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		originAllowed := isAllowedOrigin(origin)

		if corsAllowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, Authorization, Content-Type, X-Requested-With, X-Connection-ID, X-Internal-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "600")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			if !corsAllowAll && origin != "" && !originAllowed {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if !corsAllowAll && origin != "" && !originAllowed {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
