package domain

import (
	"HiveAPI/internal/transport/http/resp"
	"log/slog"
	"net/http"
	"strconv"
)

type ServiceInterface interface {
	GetItem(id uint) (*Item, error)
}

type Handler struct {
	service ServiceInterface
	log     *slog.Logger
}

func NewHandler(service ServiceInterface, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     logger,
	}
}

func (h *Handler) GetItem() func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		idItem := r.PathValue("id")
		id, err := strconv.ParseUint(idItem, 10, 32)

		item, err := h.service.GetItem(uint(id))
		if err != nil {
			return err
		}
		h.log.Debug("Код после error")
		resp.ResponseJson(w, item, http.StatusOK)
		return nil
	}
}
