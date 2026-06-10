package middleware

import (
	"RestAPI/internal/lib/sl"
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

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := sl.FromContext(r.Context())
		start := time.Now()
		wrapper := &WrapperWriter{ResponseWriter: w, StatusCode: http.StatusOK}
		next.ServeHTTP(wrapper, r)
		duration := time.Since(start)
		attrs := []any{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("uri", r.RequestURI),
			slog.String("ip", r.RemoteAddr),
			slog.String("user-agent", r.UserAgent()),
			slog.Duration("duration", duration),
			slog.Int("status", wrapper.StatusCode),
		}
		switch {
		case wrapper.StatusCode >= 500:
			log.Error("server error", attrs...)
		case wrapper.StatusCode >= 400:
			log.Warn("client error", attrs...)
		default:
			log.Debug("request processed", attrs...)
		}

	})
}
