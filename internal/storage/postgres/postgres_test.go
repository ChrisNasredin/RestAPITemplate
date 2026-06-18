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

	item := &domain.Item{
		ItemOpt1: "Value 1",
		ItemOpt2: "Value 2",
	}

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
