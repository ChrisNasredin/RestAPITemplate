package domain

import (
	"HiveAPI/internal/lib"
	"log/slog"
	"net/http"
	"strconv"
)

type ServiceInterface interface {
	GetItem(id uint) (*Item, error)
}

type Handler struct {
	service ServiceInterface
	logger  *slog.Logger
}

func NewHandler(service ServiceInterface, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) GetItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idItem := r.PathValue("id")
		id, err := strconv.ParseUint(idItem, 10, 32)

		item, err := h.service.GetItem(uint(id))
		if err != nil {
			statusCode, message := lib.ResolveHTTPStatus(err, errorToHTTPStatus, h.logger)
			lib.ResponseJson(w, map[string]string{"message": message}, statusCode)
			return
		}
		lib.ResponseJson(w, item, http.StatusOK)
	}
}
