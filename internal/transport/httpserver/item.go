package httpserver

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
	GetAllItems(ctx context.Context, limit, offset int) ([]*domain.Item, int, error)
	DeleteItem(ctx context.Context, id int64) error
	UpdateItem(ctx context.Context, item *domain.UpdateItemInput, id int64) (*domain.Item, error)
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
		const op = "transport.httpserver.GetItem"
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

func (h *Handler) GetAll() APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "transport.httpserver.GetAll"
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			// TODO: Проверить, что ошибка корректно отдается юзеру
			return fmt.Errorf("%w: %s", ErrBadRequest, "wrong limit/offset")
		}
		items, count, err := h.service.GetAllItems(r.Context(), limit, offset)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		var responseItems []*GetItemResponse
		for _, item := range items {
			newItem := &GetItemResponse{
				ID:       item.ID,
				ItemOpt1: item.ItemOpt1,
				ItemOpt2: item.ItemOpt2,
			}
			responseItems = append(responseItems, newItem)
		}
		resp := &GetAllItemsResponse{
			Items: responseItems,
			Count: count,
		}
		ResponseJson(w, &resp, http.StatusOK)
		return nil
	}
}

func (h *Handler) AddItem() APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "transport.httpserver.AddItem"
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

func (h *Handler) DeleteItem() APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "transport.httpserver.DeleteItem"
		itemID := r.PathValue("id")
		id, err := strconv.ParseUint(itemID, 10, 32)
		if err != nil {
			return ErrBadRequest
		}
		err = h.service.DeleteItem(r.Context(), int64(id))
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		ResponseJson(w, &MessageResponseJSON{
			Message: "Success",
		}, http.StatusAccepted)
		return nil
	}
}

func (h *Handler) UpdateItem() APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "transport.httpserver.UpdateItem"
		itemID := r.PathValue("id")
		id, err := strconv.ParseUint(itemID, 10, 32)
		if err != nil {
			return ErrBadRequest
		}
		body, err := HandleBody[UpdateItemRequest](&w, r)
		if err != nil {
			return err
		}
		item := &domain.UpdateItemInput{
			ItemOpt1: body.ItemOpt1,
			ItemOpt2: body.ItemOpt2,
		}
		updatedItem, err := h.service.UpdateItem(r.Context(), item, int64(id))
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		ResponseJson(w, &GetItemResponse{
			ID:       updatedItem.ID,
			ItemOpt1: updatedItem.ItemOpt1,
			ItemOpt2: updatedItem.ItemOpt2,
		}, http.StatusOK)
		return nil
	}
}
