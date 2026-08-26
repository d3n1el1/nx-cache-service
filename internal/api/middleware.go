package api

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"nxCacheService/internal/auth"
)

const bearerPrefix = "bearer "

type Authenticator interface {
	IsKnown(token string) bool
	IsAllowed(token string, action auth.AllowedAction) bool
}

func requireStaticTokenAuth(authenticator Authenticator, next http.Handler, mode auth.AllowedAction) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))

		if !authenticator.IsKnown(token) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			_ = writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}

		if !authenticator.IsAllowed(token, mode) {
			_ = writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func bearerToken(header string) string {
	raw := strings.TrimSpace(header)

	if len(raw) < len(bearerPrefix) || !strings.EqualFold(raw[:len(bearerPrefix)], bearerPrefix) {
		return ""
	}

	return strings.TrimSpace(raw[len(bearerPrefix):])
}

func requestLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &responseRecorder{ResponseWriter: w}

			next.ServeHTTP(recorder, r)

			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.statusCode(),
				"bytes", recorder.bytes,
				"duration", time.Since(started))
		})
	}
}

func recoverPanic(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &responseRecorder{ResponseWriter: w}

			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}

				log.Error("panic recovered",
					"err", recovered,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()))

				if !recorder.wroteHeader() {
					_ = writeErrorResponse(recorder, http.StatusInternalServerError, "Internal Server Error", nil)
				}
			}()

			next.ServeHTTP(recorder, r)
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rec *responseRecorder) WriteHeader(status int) {
	if rec.status != 0 {
		return
	}

	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.WriteHeader(http.StatusOK)
	}

	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n

	return n, err
}

func (rec *responseRecorder) Flush() {
	if flusher, ok := rec.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rec *responseRecorder) wroteHeader() bool {
	return rec.status != 0
}

func (rec *responseRecorder) statusCode() int {
	if rec.status == 0 {
		return http.StatusOK
	}

	return rec.status
}
