package api

import (
	"log/slog"
	"net/http"

	"nxCacheService/internal/cache"
)

func handleCacheGet(store cache.Store, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = writeErrorResponse(w, http.StatusNotImplemented, "Not implemented", nil)
	})
}

func handleCachePut(store cache.Store, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = writeErrorResponse(w, http.StatusNotImplemented, "Not implemented", nil)
	})
}
