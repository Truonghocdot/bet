package http

import (
	"net"
	"net/http"
	"strings"

	"gin/internal/domain/auth"
)

func extractRequestMeta(r *http.Request) auth.RequestMeta {
	return auth.RequestMeta{
		IP:        clientIP(r),
		UserAgent: strings.TrimSpace(r.UserAgent()),
	}
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}
