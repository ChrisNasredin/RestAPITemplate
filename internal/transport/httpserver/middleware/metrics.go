package middleware

import (
	"RestAPI/internal/metrics"
	"net/http"
	"strconv"
	"time"
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &WrapperWriter{ResponseWriter: w, StatusCode: http.StatusOK}
		next.ServeHTTP(wrapper, r)
		duration := time.Since(start)
		status := strconv.Itoa(wrapper.StatusCode)
		metrics.HttpRequestTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		metrics.HttpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
	})
}
