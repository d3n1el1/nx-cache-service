package api

import (
	"log/slog"
	"net/http"

	"nxCacheService/internal/cache"
)

func NewServer(store cache.Store, auth Authenticator, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, store, auth, log)

	return recoverPanic(log)(requestLog(log)(mux))
}
