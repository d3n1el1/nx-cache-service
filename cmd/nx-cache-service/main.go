package main

import (
	"log/slog"
	"net/http"
	"os"

	"nxCacheService/internal/api"
	"nxCacheService/internal/auth"
	"nxCacheService/internal/cache"
)

const (
	addr     = "localhost:8080"
	cacheDir = ".nx-cache-service"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store, err := cache.NewDisk(cacheDir)
	if err != nil {
		log.Error("open cache", "err", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: api.NewServer(store, auth.StaticTokenAuthenticator{}, log),
	}

	log.Info("nx cache service listening", "addr", addr, "dir", store.Root())
	if err := srv.ListenAndServe(); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
