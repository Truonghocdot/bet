package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"gin/internal/domain/auth"
	"gin/internal/service"
	"gin/internal/support/message"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

type authRequestError struct {
	Message string
	Code    string
}

type Authentication struct {
	authService *service.AuthService
}

func NewAuthentication(authService *service.AuthService) *Authentication {
	return &Authentication{authService: authService}
}

func (m *Authentication) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok, err := m.authenticateRequest(r)
		if err != nil {
			payload := map[string]string{"message": err.Message}
			if err.Code != "" {
				payload["code"] = err.Code
			}
			writeJSON(w, http.StatusUnauthorized, payload)
			return
		}
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.MissingBearerToken})
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Authentication) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok, err := m.authenticateRequest(r)
		if err != nil {
			payload := map[string]string{"message": err.Message}
			if err.Code != "" {
				payload["code"] = err.Code
			}
			writeJSON(w, http.StatusUnauthorized, payload)
			return
		}
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CurrentClaims(ctx context.Context) (auth.TokenClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(auth.TokenClaims)
	return claims, ok
}

func extractBearerToken(authorization string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(authorization), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	tokenValue := strings.TrimSpace(parts[1])
	if tokenValue == "" {
		return "", false
	}

	return tokenValue, true
}

func (m *Authentication) authenticateRequest(r *http.Request) (auth.TokenClaims, bool, *authRequestError) {
	tokenValue, ok := extractBearerToken(r.Header.Get("Authorization"))
	if !ok {
		queryToken := strings.TrimSpace(r.URL.Query().Get("access_token"))
		if queryToken != "" {
			tokenValue = queryToken
			ok = true
		}
	}
	if !ok {
		return auth.TokenClaims{}, false, nil
	}

	claims, err := m.authService.VerifyAccessToken(tokenValue)
	if err != nil {
		return auth.TokenClaims{}, false, &authRequestError{Message: message.InvalidAccessToken}
	}

	if !m.authService.VerifySession(r.Context(), claims.UserID, claims.SessionID) {
		log.Printf("[auth][session.invalidated] user_id=%d path=%s method=%s", claims.UserID, r.URL.Path, r.Method)
		return auth.TokenClaims{}, false, &authRequestError{
			Message: "Tài khoản của bạn đã được đăng nhập từ một thiết bị khác. Vui lòng đăng nhập lại.",
			Code:    "SESSION_INVALIDATED",
		}
	}

	return claims, true, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
