package api

import (
	"log/slog"
	"net/http"
	"time"
)

// responseRecorder wraps http.ResponseWriter to capture the status code.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware wraps a handler to log incoming requests with method, path, status, and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", recorder.status, "duration_ms", time.Since(start).Milliseconds())
	})
}
