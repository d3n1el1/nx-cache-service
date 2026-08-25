package api

import (
	"log/slog"
	"net/http"
	"nxCacheService/internal/auth"
)

func requireTokenAuth(authenticator auth.StaticTokenAuthenticator, next http.Handler, mode auth.AllowedAction) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		allowed := authenticator.IsAllowed(token, mode)

		if !allowed {
			_ = writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
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
