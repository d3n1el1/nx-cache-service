package api

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"nxCacheService/internal/cache"
)

func handleCacheGet(store cache.Store, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")
		hash := r.PathValue("hash")

		if err := validatePathParam("project", project); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if err := validatePathParam("hash", hash); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		body, size, err := store.Get(r.Context(), project, hash)

		if err != nil {
			if !errors.Is(err, cache.ErrNotFound) {
				log.Error("cache get failed", "project", project, "hash", hash, "err", err)
			}

			_ = writeErrorResponse(w, http.StatusNotFound, "Not Found", nil)

			return
		}

		defer func() { _ = body.Close() }()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		w.WriteHeader(http.StatusOK)

		if _, err := io.Copy(w, body); err != nil {
			log.Error("cache get stream failed", "project", project, "hash", hash, "err", err)
		}
	})
}

func handleCachePut(store cache.Store, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")
		hash := r.PathValue("hash")

		if err := validatePathParam("project", project); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if err := validatePathParam("hash", hash); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		err := store.Put(r.Context(), project, hash, r.Body)

		if err != nil {
			if errors.Is(err, cache.ErrExists) {
				_ = writeErrorResponse(w, http.StatusConflict, "Conflict", nil)
				return
			}

			log.Error("cache put failed", "project", project, "hash", hash, "err", err)
			_ = writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", nil)

			return
		}

		_ = writeJSON(w, http.StatusOK, struct{}{})
	})
}

func handleCacheFlush(store cache.Store, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")

		if err := validatePathParam("project", project); err != nil {
			_ = writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if err := store.Flush(r.Context(), project); err != nil {
			log.Error("cache flush failed", "project", project, "err", err)
			_ = writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", nil)

			return
		}

		_ = writeJSON(w, http.StatusOK, struct{}{})
	})
}
