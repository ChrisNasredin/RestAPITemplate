package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type WrapperWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WrapperWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.StatusCode = code
}

func Logging(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapper := &WrapperWriter{ResponseWriter: w, StatusCode: http.StatusOK}
			next.ServeHTTP(wrapper, r)
			duration := time.Since(start)
			log.Debug("",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("uri", r.RequestURI),
				slog.String("ip", r.RemoteAddr),
				slog.String("user-agent", r.UserAgent()),
				slog.String("duration", duration.String()),
				slog.Int("status", wrapper.StatusCode),
			)

		})
	}
}
