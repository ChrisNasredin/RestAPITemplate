package http_server

import (
	"RestAPI/internal/domain"
	"log/slog"
	"net/http"
	"strconv"
)

type ServiceInterface interface {
	GetItem(id uint) (*domain.Item, error)
}

type Handler struct {
	service ServiceInterface
	log     *slog.Logger
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetItem() APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		idItem := r.PathValue("id")
		id, err := strconv.ParseUint(idItem, 10, 32)
		if err != nil {
			return ErrBadRequest
		}
		item, err := h.service.GetItem(uint(id))
		if err != nil {
			return err
		}
		ResponseJson(w, item, http.StatusOK)
		return nil
	}
}
