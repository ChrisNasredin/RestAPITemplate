package postgres

import (
	"RestAPI/internal/domain"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testStorage *Storage
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	container, err := pgContainer.Run(
		ctx,
		"postgres:16-alpine",
		pgContainer.WithDatabase("restapi"),
		pgContainer.WithUsername("username"),
		pgContainer.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			// Ждем, пока Postgres явно напишет в логи, что готов принимать коннекты
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2). // На alpine он пишет это дважды при старте
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}

	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://username:password@127.0.0.1:%s/restapi?sslmode=disable", port.Port())

	if err = RunMigrations(dsn); err != nil {
		panic(err)
	}

	// 2. Initialize the ONLY storage pool
	testStorage, err = New(ctx, &StorageConfig{
		Host:            "127.0.0.1:" + port.Port(),
		User:            "username",
		Password:        "password",
		DBName:          "restapi",
		SSLMode:         "disable",
		MaxConns:        20,
		MinConns:        5,
		ConnectTimeout:  20 * time.Second,
		MaxConnLifetime: 5 * time.Second,
		MaxConnIdleTime: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	testStorage.pool.Close()
	container.Terminate(ctx)
	os.Exit(code)
}

func TestStorage_ItemLifecycle(t *testing.T) {
	ctx := context.Background()

	// Use the same pool for cleanup to avoid desync
	_, err := testStorage.pool.Exec(ctx, "DELETE FROM items")
	require.NoError(t, err)

	items := []*domain.Item{
		&domain.Item{
			ItemOpt1: "Value 1.1",
			ItemOpt2: "Value 1.2",
		},
		&domain.Item{
			ItemOpt1: "Value 2.1",
			ItemOpt2: "Value 2.2",
		},
	}

	for _, item := range items {
		// Create
		created, err := testStorage.CreateItem(ctx, item)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.NotZero(t, created.ID)
		// Get
		fetched, err := testStorage.GetItemByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ItemOpt1, fetched.ItemOpt1)
		assert.Equal(t, item.ItemOpt2, fetched.ItemOpt2)
	}

	// All & Count
	count, err := testStorage.GetAllItemsCount(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, len(items), count)

	allItems, err := testStorage.GetAllItems(ctx, 2, 0)
	require.NoError(t, err)
	assert.Len(t, allItems, len(items))
}

// cleanItems удаляет все записи между тестами, чтобы они не зависели друг от друга.
func cleanItems(t *testing.T) {
	t.Helper()
	_, err := testStorage.pool.Exec(context.Background(), "DELETE FROM items")
	require.NoError(t, err)
}

// seedItems создаёт n записей и возвращает их в порядке вставки.
func seedItems(t *testing.T, n int) []*domain.Item {
	t.Helper()
	ctx := context.Background()
	var created []*domain.Item
	for i := range n {
		item, err := testStorage.CreateItem(ctx, &domain.Item{
			ItemOpt1: fmt.Sprintf("opt1-%d", i+1),
			ItemOpt2: fmt.Sprintf("opt2-%d", i+1),
		})
		require.NoError(t, err)
		created = append(created, item)
	}
	return created
}

// TestGetAllItems_ReturnsAllItems проверяет, что при limit >= кол-ва записей
// возвращаются все вставленные элементы с правильными полями.
func TestGetAllItems_ReturnsAllItems(t *testing.T) {
	cleanItems(t)
	seeded := seedItems(t, 3)

	got, err := testStorage.GetAllItems(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, got, len(seeded))

	// Проверяем, что каждый вставленный item присутствует в ответе
	gotIDs := make(map[int64]bool, len(got))
	for _, item := range got {
		gotIDs[item.ID] = true
	}
	for _, s := range seeded {
		assert.True(t, gotIDs[s.ID], "item with ID %d not found in result", s.ID)
	}
}

// TestGetAllItems_LimitRestrictsCount проверяет, что параметр limit
// ограничивает количество возвращаемых элементов.
func TestGetAllItems_LimitRestrictsCount(t *testing.T) {
	cleanItems(t)
	seedItems(t, 5)

	got, err := testStorage.GetAllItems(context.Background(), 3, 0)
	require.NoError(t, err)
	assert.Len(t, got, 3)
}

// TestGetAllItems_OffsetSkipsItems проверяет, что параметр offset
// пропускает нужное количество записей.
func TestGetAllItems_OffsetSkipsItems(t *testing.T) {
	cleanItems(t)
	seeded := seedItems(t, 4)

	// Берём все, чтобы знать порядок сортировки по умолчанию (по id)
	firstPage, err := testStorage.GetAllItems(context.Background(), 4, 0)
	require.NoError(t, err)
	require.Len(t, firstPage, 4)

	// Со смещением 2 должны вернуться последние два элемента
	secondPage, err := testStorage.GetAllItems(context.Background(), 4, 2)
	require.NoError(t, err)
	require.Len(t, secondPage, 2)

	_ = seeded // seeded используется только для наглядности
	assert.Equal(t, firstPage[2].ID, secondPage[0].ID)
	assert.Equal(t, firstPage[3].ID, secondPage[1].ID)
}

// TestGetAllItems_EmptyTable проверяет, что при пустой таблице
// возвращается domain.ErrNotFound.
func TestGetAllItems_EmptyTable(t *testing.T) {
	cleanItems(t)

	got, err := testStorage.GetAllItems(context.Background(), 10, 0)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// TestGetAllItems_ExcludesSoftDeleted проверяет, что записи с
// ненулевым deleted_at не попадают в результат.
func TestGetAllItems_ExcludesSoftDeleted(t *testing.T) {
	cleanItems(t)
	seeded := seedItems(t, 3)

	// Мягко удаляем первую запись напрямую через пул
	_, err := testStorage.pool.Exec(
		context.Background(),
		"UPDATE items SET deleted_at = NOW() WHERE id = $1",
		seeded[0].ID,
	)
	require.NoError(t, err)

	got, err := testStorage.GetAllItems(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, got, 2, "soft-deleted item should not be returned")

	for _, item := range got {
		assert.NotEqual(t, seeded[0].ID, item.ID, "soft-deleted item must not appear in results")
	}
}
