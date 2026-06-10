package domain

import (
	"errors"
	"fmt"
)

type Repository struct {
	storage map[uint]*Item
}

func NewRepository() *Repository {
	return &Repository{
		storage: map[uint]*Item{
			1: &Item{ID: 1, ItemOpt1: "opt1", ItemOpt2: "opt2"},
			2: &Item{ID: 2, ItemOpt1: "opt1", ItemOpt2: "opt2"},
			3: &Item{ID: 3, ItemOpt1: "opt1", ItemOpt2: "opt2"},
		},
	}
}

func (r *Repository) GetItem(id uint) (*Item, error) {
	const op = "domain.Repository.GetItem"
	if _, exists := r.storage[id]; !exists {
		return nil, fmt.Errorf(op+" -> %w", errors.New("все жоска пизданулось, нам всем пизда"))
	}
	return r.storage[id], nil
}
