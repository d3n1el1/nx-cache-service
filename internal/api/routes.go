package api

import (
	"log/slog"
	"net/http"
	"nxCacheService/internal/auth"

	"nxCacheService/internal/cache"
)

func addRoutes(mux *http.ServeMux, store cache.Store, authorization Authenticator, log *slog.Logger) {
	mux.Handle("GET /project/{project}/v1/cache/{hash}", requireStaticTokenAuth(authorization, handleCacheGet(store, log), auth.Read))
	mux.Handle("PUT /project/{project}/v1/cache/{hash}", requireStaticTokenAuth(authorization, handleCachePut(store, log), auth.Write))

	mux.Handle("GET /health", handleHealth())

	mux.Handle("/", http.NotFoundHandler())
}
