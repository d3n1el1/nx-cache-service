package api

import (
	"log/slog"
	"net/http"

	"nxCacheService/internal/cache"
)

func addRoutes(mux *http.ServeMux, store cache.Store, auth Authenticator, log *slog.Logger) {
	mux.Handle("GET /v1/cache/{hash}", requireAuth(auth, handleCacheGet(store, log)))
	mux.Handle("PUT /v1/cache/{hash}", requireAuth(auth, handleCachePut(store, log)))

	mux.Handle("GET /health", handleHealth())

	mux.Handle("/", http.NotFoundHandler())
}
