package api

import (
	"log/slog"
	"net/http"

	"nxCacheService/internal/cache"
)

func NewServer(store cache.Store, authenticator Authenticator, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, store, authenticator, log)

	return recoverPanic(log)(requestLog(log)(mux))
}
