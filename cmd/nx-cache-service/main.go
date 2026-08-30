package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"nxCacheService/internal/api"
	"nxCacheService/internal/auth"
	"nxCacheService/internal/cache"
	"nxCacheService/internal/env"
)

const (
	addr     = "localhost:8080"
	cacheDir = ".nx-cache-service"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	authenticator, err := auth.NewStaticTokenAuthenticator()
	if err != nil {
		log.Error("configure auth", "err", err)
		os.Exit(1)
	}

	store, location, err := openStore(context.Background())
	if err != nil {
		log.Error("open cache", "err", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: api.NewServer(store, authenticator, log),
	}

	log.Info("nx cache service listening",
		"addr", addr,
		"cache", location,
		"readOnlyToken", authenticator.HasReadOnlyToken())

	if err := srv.ListenAndServe(); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func openStore(ctx context.Context) (cache.Store, string, error) {
	switch selected := env.CacheStore.GetValue(); selected {
	case "", "disk":
		store, err := cache.NewDisk(cacheDir)
		if err != nil {
			return nil, "", err
		}

		return store, store.Root(), nil
	case "s3":
		store, err := cache.NewS3(ctx)
		if err != nil {
			return nil, "", err
		}

		return store, "s3://" + store.Bucket(), nil
	default:
		return nil, "", fmt.Errorf("unknown cache store %q, want 'disk' or 's3'", selected)
	}
}
