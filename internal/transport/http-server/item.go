package http_server

import (
	"RestAPI/internal/domain"
	"context"
	"fmt"
	"net/http"
	"strconv"
)

type ServiceInterface interface {
	GetItem(ctx context.Context, id int64) (*domain.Item, error)
	CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error)
}

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetItem() APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "transport.http-server.GetItem"
		itemID := r.PathValue("id")
		id, err := strconv.ParseUint(itemID, 10, 32)
		if err != nil {
			return ErrBadRequest
		}
		item, err := h.service.GetItem(r.Context(), int64(id))
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		ResponseJson(w, &GetItemResponse{
			ID:       item.ID,
			ItemOpt1: item.ItemOpt1,
			ItemOpt2: item.ItemOpt2,
		}, http.StatusOK)
		return nil
	}
}

func (h *Handler) AddItem() APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "transport.http-server.AddItem"
		body, err := HandleBody[CreateItemRequest](&w, r)
		if err != nil {
			return err
		}
		newItem := &domain.Item{
			ItemOpt1: body.ItemOpt1,
			ItemOpt2: body.ItemOpt2,
		}
		createdItem, err := h.service.CreateItem(r.Context(), newItem)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		ResponseJson(w, &CreateItemResponse{
			ID:       createdItem.ID,
			ItemOpt1: createdItem.ItemOpt1,
			ItemOpt2: createdItem.ItemOpt2,
		}, http.StatusCreated)
		return nil
	}
}
