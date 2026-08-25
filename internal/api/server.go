package api

import (
	"log/slog"
	"net/http"
	"nxCacheService/internal/auth"

	"nxCacheService/internal/cache"
)

func NewServer(store cache.Store, authenticator auth.StaticTokenAuthenticator, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, store, authenticator, log)

	return recoverPanic(log)(requestLog(log)(mux))
}
