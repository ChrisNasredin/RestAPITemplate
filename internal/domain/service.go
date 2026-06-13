package domain

import (
	"context"
	"fmt"
)

type RepositoryInterface interface {
	GetItemByID(ctx context.Context, id int64) (*Item, error)
	CreateItem(ctx context.Context, item *Item) (*Item, error)
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

func (s *Service) CreateItem(ctx context.Context, item *Item) (*Item, error) {
	const op = "domain.Service.CreateItem"

	item, err := s.repository.CreateItem(ctx, item)
	if err != nil {
		return nil, fmt.Errorf(op+"-> %w", err)
	}
	return item, err
}
