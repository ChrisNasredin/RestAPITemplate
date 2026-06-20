package domain

import (
	"context"
	"fmt"
)

type RepositoryInterface interface {
	GetItemByID(ctx context.Context, id int64) (*Item, error)
	CreateItem(ctx context.Context, item *Item) (*Item, error)
	GetAllItems(ctx context.Context, limit, offset int) ([]*Item, error)
	GetAllItemsCount(ctx context.Context) (int, error)
	DeleteItemByID(ctx context.Context, id int64) error
}
type Service struct {
	repository RepositoryInterface
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetItem(ctx context.Context, id int64) (*Item, error) {
	const op = "domain.Service.GetItem"

	item, err := s.repository.GetItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(op+"-> %w", err)
	}
	return item, err
}

func (s *Service) DeleteItem(ctx context.Context, id int64) error {
	const op = "domain.Service.DeleteItem"
	err := s.repository.DeleteItemByID(ctx, id)
	if err != nil {
		return fmt.Errorf(op+"-> %w", err)
	}
	return nil
}

func (s *Service) GetAllItems(ctx context.Context, limit, offset int) ([]*Item, int, error) {
	const op = "domain.Service.GetItem"

	items, err := s.repository.GetAllItems(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf(op+"-> %w", err)
	}
	count, err := s.repository.GetAllItemsCount(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf(op+"-> %w", err)
	}
	return items, count, err
}

func (s *Service) CreateItem(ctx context.Context, item *Item) (*Item, error) {
	const op = "domain.Service.CreateItem"

	item, err := s.repository.CreateItem(ctx, item)
	if err != nil {
		return nil, fmt.Errorf(op+"-> %w", err)
	}
	return item, err
}
