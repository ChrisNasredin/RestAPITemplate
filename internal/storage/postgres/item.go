package postgres

import (
	"RestAPI/internal/domain"
	"context"
)

func (s *Storage) GetItemByID(ctx context.Context, id int64) (*domain.Item, error) {
	const (
		op    = "storage.postgres.GetItemByID"
		query = `
			SELECT id, item_opt1, item_opt2
			FROM items 
			WHERE id = $1 
			AND deleted_at IS NULL`
	)
	var item ItemStorage
	err := s.pool.QueryRow(ctx, query,
		id).Scan(&item.ID, &item.ItemOpt1, &item.ItemOpt2)
	if err != nil {
		return nil, mapErr(op, err)
	}
	return &domain.Item{
		ID:       item.ID,
		ItemOpt1: item.ItemOpt1,
		ItemOpt2: item.ItemOpt2,
	}, nil
}

func (s *Storage) CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	const (
		op    = "storage.postgres.CreateItem"
		query = `
		INSERT INTO items (item_opt1, item_opt2) 
		VALUES ($1, $2)
		RETURNING id, item_opt1, item_opt2
		`
	)
	var itemStorage ItemStorage
	err := s.pool.QueryRow(ctx, query, item.ItemOpt1, item.ItemOpt2).Scan(
		&itemStorage.ID,
		&itemStorage.ItemOpt1,
		&itemStorage.ItemOpt2,
	)
	if err != nil {
		return nil, mapErr(op, err)
	}

	item.ID = itemStorage.ID
	item.ItemOpt1 = itemStorage.ItemOpt1
	item.ItemOpt2 = itemStorage.ItemOpt2

	return item, nil
}
