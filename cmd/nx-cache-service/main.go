package main

import (
	"context"
	"crypto/tls"
	"errors"
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
	defaultAddr = "localhost:8080"
	cacheDir    = ".nx-cache-service"
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

	addr := env.Addr.GetValue()
	if len(addr) == 0 {
		addr = defaultAddr
	}

	certFile, keyFile, err := tlsFiles()
	if err != nil {
		log.Error("configure tls", "err", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:      addr,
		Handler:   api.NewServer(store, authenticator, log),
		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	log.Info("nx cache service listening",
		"addr", addr,
		"cache", location,
		"tls", len(certFile) > 0,
		"readOnlyToken", authenticator.HasReadOnlyToken())

	if len(certFile) > 0 {
		err = srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func tlsFiles() (string, string, error) {
	certFile := env.TlsCertFile.GetValue()
	keyFile := env.TlsKeyFile.GetValue()

	if len(certFile) == 0 && len(keyFile) == 0 {
		return "", "", nil
	}

	if len(certFile) == 0 || len(keyFile) == 0 {
		return "", "", errors.New("TLS_CERT_FILE and TLS_KEY_FILE must be set together")
	}

	return certFile, keyFile, nil
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
