package api

import (
	"log/slog"
	"net/http"

	"nxCacheService/internal/cache"
)

func handleCacheGet(store cache.Store, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := validatePathParam("project", r.PathValue("project")); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if err := validatePathParam("hash", r.PathValue("hash")); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		_ = writeErrorResponse(w, http.StatusNotImplemented, "Not implemented", nil)
	})
}

func handleCachePut(store cache.Store, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := validatePathParam("project", r.PathValue("project")); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if err := validatePathParam("hash", r.PathValue("hash")); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		_ = writeErrorResponse(w, http.StatusNotImplemented, "Not implemented", nil)
	})
}

func handleCacheFlush(store cache.Store, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := validatePathParam("project", r.PathValue("project")); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		_ = writeErrorResponse(w, http.StatusNotImplemented, "Not implemented", nil)
	})
}
