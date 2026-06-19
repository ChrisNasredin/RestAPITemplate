package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total Number HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
		},
		[]string{"method", "path"},
	)
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &WrapperWriter{ResponseWriter: w, StatusCode: http.StatusOK}
		next.ServeHTTP(wrapper, r)
		duration := time.Since(start)
		status := strconv.Itoa(wrapper.StatusCode)
		httpRequestTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
	})
}
