package middleware

import (
	"HiveAPI/internal/lib/sl"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func RequestID(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Вытащить хедер X-Request-ID, если нет, то сгенерить
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = uuid.NewString()
			}
			w.Header().Set("X-Request-ID", id)
			childLog := log.With("request_id", id)
			ctx := sl.ContextWithLogger(r.Context(), childLog)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
