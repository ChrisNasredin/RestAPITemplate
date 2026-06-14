package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StorageConfig struct {
	User            string
	Password        string
	Host            string
	DBName          string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	ConnectTimeout  time.Duration
	MaxConnIdleTime time.Duration
}
type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *StorageConfig) (*Storage, error) {
	const op = "storage.postgres.NewRepository"
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.DBName, cfg.SSLMode)
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: parse config: %w", op, err)
	}
	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	config.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%s: create pool: %w", op, err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err = pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("%s: ping db: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}
