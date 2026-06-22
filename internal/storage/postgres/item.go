package postgres

import (
	"RestAPI/internal/domain"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"
)

type ItemStorage struct {
	ID        int64      `db:"id"`
	ItemOpt1  *string    `db:"item_opt1"`
	ItemOpt2  *string    `db:"item_opt2"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (s *Storage) GetItemByID(ctx context.Context, id int64) (*domain.Item, error) {
	const (
		op    = "storage.postgres.GetItemByID"
		query = `
			SELECT id, item_opt1, item_opt2, created_at, updated_at, deleted_at
			FROM items 
			WHERE id = $1 
			AND deleted_at IS NULL`
	)
	var item ItemStorage
	// Без использования pgxutils:
	//rows, err := s.pool.Query(ctx, query, id)
	//if err != nil {
	//	return nil, err
	//}
	//item, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[ItemStorage])
	//if err != nil {
	//	return nil, err
	//}
	// Однострочник:
	item, err := pgxutil.SelectRow[ItemStorage](ctx, s.pool, query, []any{id}, pgx.RowToStructByName[ItemStorage])
	if err != nil {
		return nil, mapErr(op, err)
	}
	return &domain.Item{
		ID:       item.ID,
		ItemOpt1: *item.ItemOpt1,
		ItemOpt2: *item.ItemOpt2,
	}, nil
}

func (s *Storage) DeleteItem(ctx context.Context, id int64) error {
	const (
		op    = "storage.postgres.DeleteItem"
		query = `
			UPDATE items 
			SET deleted_at = NOW()
			WHERE id = $1 
			AND deleted_at IS NULL`
	)
	err := s.pool.QueryRow(ctx, query, id).Scan()
	if err != nil {
		return mapErr(op, err)
	}
	return nil
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
	item.ItemOpt1 = *itemStorage.ItemOpt1
	item.ItemOpt2 = *itemStorage.ItemOpt2

	return item, nil
}

func (s *Storage) GetAllItems(ctx context.Context, limit, offset int) ([]*domain.Item, error) {
	const (
		op    = "storage.postgres.GetAllItems"
		query = `
		SELECT id, item_opt1, item_opt2, created_at, updated_at, deleted_at
			FROM items 
			WHERE deleted_at IS NULL
			LIMIT $1 OFFSET $2
			`
	)
	itemsStorage, err := pgxutil.Select[ItemStorage](ctx, s.pool, query, []any{limit, offset}, pgx.RowToStructByName[ItemStorage])
	if err != nil {
		return nil, mapErr(op, err)
	}
	if len(itemsStorage) == 0 {
		return nil, domain.ErrNotFound
	}
	var result []*domain.Item
	for _, item := range itemsStorage {
		result = append(result, &domain.Item{
			ID:       item.ID,
			ItemOpt1: *item.ItemOpt1,
			ItemOpt2: *item.ItemOpt2,
		})
	}
	return result, nil
}

func (s *Storage) GetAllItemsCount(ctx context.Context) (int, error) {
	const (
		op    = "storage.postgres.GetAllItemsCount"
		query = `
		SELECT count(*)
			FROM items 
			WHERE deleted_at IS NULL
			`
	)
	var count int
	err := s.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, mapErr(op, err)
	}
	return count, nil
}

// TODO: сделать возможность поставить NULL. Сейчас пользователь не может затереть поле в NULL
// Да и в целом подумать над методом, чето он не особо клевый. Если не передано ничо, то он тоже подорвется
// все менять, да и еще и поставит таймстемп.
// Ну и потом покрыть тестами надо, когда передается пустая структурка - нужно ничего не делать
func (s *Storage) UpdateItem(ctx context.Context, item *domain.UpdateItemInput, id int64) (*domain.Item, error) {
	const (
		op    = "storage.postgres.UpdateItem"
		query = `
		UPDATE items
			SET item_opt1 = COALESCE($1, item_opt1), item_opt2 = COALESCE($2, item_opt2), updated_at = NOW()
			WHERE ID = $3 AND deleted_at IS NULL
		RETURNING id, item_opt1, item_opt2
			`
	)
	var updatedItem ItemStorage
	err := s.pool.QueryRow(ctx, query, item.ItemOpt1, item.ItemOpt2, id).Scan(
		&updatedItem.ID,
		&updatedItem.ItemOpt1,
		&updatedItem.ItemOpt2,
	)
	if err != nil {
		return nil, mapErr(op, err)
	}
	return &domain.Item{
		ID:       updatedItem.ID,
		ItemOpt1: *updatedItem.ItemOpt1,
		ItemOpt2: *updatedItem.ItemOpt2,
	}, nil
}

func (s *Storage) ItemsCount(ctx context.Context) (int64, error) {
	const (
		op    = "storage.postgres.ItemCount"
		query = `
		SELECT count(*)
			FROM items 
			WHERE deleted_at IS NULL
			`
	)
	var count int64
	err := s.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, mapErr(op, err)
	}
	return count, nil
}
