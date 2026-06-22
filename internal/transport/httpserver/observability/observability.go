package observability

import (
	"RestAPI/internal/transport/httpserver"
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Pinger interface {
	Ping(ctx context.Context) error
}
type Handler struct {
	db Pinger
}

func NewHandler(obsRouter *http.ServeMux, db Pinger) *Handler {
	handler := &Handler{
		db: db,
	}
	obsRouter.Handle("GET /healthz", handler.Liveness())
	obsRouter.Handle("GET /readyz", handler.Readiness())
	obsRouter.Handle("GET /metrics", promhttp.Handler())

	return handler
}

func (h *Handler) Liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpserver.ResponseJson(w, map[string]string{"status": "ok"}, http.StatusOK)
	}
}

func (h *Handler) Readiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.db.Ping(ctx); err != nil {
			httpserver.ResponseJson(w, map[string]string{
				"status": "unavailable",
				"error":  err.Error(),
			}, http.StatusServiceUnavailable)
			return
		}
		httpserver.ResponseJson(w, map[string]string{"status": "ok"}, http.StatusOK)
	}
}
