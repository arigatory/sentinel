package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func Logger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				size:           0,
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			logger.Info("HTTP request",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Int("status", wrapped.statusCode),
				slog.Int("size", wrapped.size),
				slog.Duration("duration", duration),
			)
		})
	}
}
