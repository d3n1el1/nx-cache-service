package api

import (
	"log/slog"
	"net/http"
	"strings"

	"nxCacheService/internal/auth"
)

const bearerPrefix = "bearer "

type Authenticator interface {
	IsKnown(token string) bool
	IsAllowed(token string, action auth.AllowedAction) bool
}

func requireStaticTokenAuth(authenticator Authenticator, next http.Handler, mode auth.AllowedAction) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))

		if !authenticator.IsKnown(token) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			_ = writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}

		if !authenticator.IsAllowed(token, mode) {
			_ = writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func bearerToken(header string) string {
	raw := strings.TrimSpace(header)

	if len(raw) < len(bearerPrefix) || !strings.EqualFold(raw[:len(bearerPrefix)], bearerPrefix) {
		return ""
	}

	return strings.TrimSpace(raw[len(bearerPrefix):])
}

func requestLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func recoverPanic(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
