package domain

import (
	"fmt"
)

type RepositoryInterface interface {
	GetItem(id uint) (*Item, error)
}
type Service struct {
	repository RepositoryInterface
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetItem(id uint) (*Item, error) {
	const op = "domain.Service.GetItem"

	item, err := s.repository.GetItem(id)
	if err != nil {
		return nil, fmt.Errorf(op+"-> %w", err)
	}
	return item, err
}
