package logger

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// New creates a JSON zerolog logger configured at the given level. Unknown or
// empty levels fall back to info.
func New(level string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil || level == "" {
		lvl = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano
	return zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Str("service", "shiftmanager").
		Logger()
}

// RequestLogger returns a chi-compatible middleware that emits a structured
// JSON log line for every HTTP request. It pairs with chi's RequestID
// middleware to include the request id when present.
func RequestLogger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				ev := log.Info()
				if rec := recover(); rec != nil {
					ev = log.Error().Interface("panic", rec)
					ww.WriteHeader(http.StatusInternalServerError)
				}
				ev.
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", ww.Status()).
					Int("bytes", ww.BytesWritten()).
					Dur("duration", time.Since(start)).
					Str("remote", r.RemoteAddr).
					Str("request_id", middleware.GetReqID(r.Context())).
					Msg("http_request")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
