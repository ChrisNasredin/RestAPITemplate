package main

import (
	"RestAPI/internal/config"
	"RestAPI/internal/domain"
	"RestAPI/internal/storage/postgres"
	"RestAPI/internal/transport/http-server"
	"RestAPI/internal/transport/http-server/middleware"
	"context"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	slog.SetDefault(log)
	log.Info("starting server")
	log.Debug("debug logging enabled")
	ctx := context.Background()
	storage, err := postgres.New(
		ctx,
		&postgres.StorageConfig{
			User:            cfg.Storage.User,
			Password:        cfg.Storage.Password,
			Host:            cfg.Storage.Host,
			DBName:          cfg.Storage.DBName,
			SSLMode:         cfg.Storage.SSLMode,
			MaxConns:        cfg.Storage.StoragePool.MaxConns,
			MinConns:        cfg.Storage.StoragePool.MinConns,
			MaxConnLifetime: cfg.Storage.StoragePool.MaxConnLifetime,
			ConnectTimeout:  cfg.Storage.StoragePool.ConnectTimeout,
			MaxConnIdleTime: cfg.Storage.StoragePool.MaxConnIdleTime,
		},
	)
	if err != nil {
		log.Error("failed to init storage", slog.Any("error", err))
		os.Exit(1)
	}

	domainService := domain.NewService(storage)
	domainHandler := http_server.NewHandler(domainService)

	router := http.NewServeMux()

	// Middleware
	mwErrHandler := middleware.ErrorHandler(http_server.ErrorToHTTPStatus)
	mwChain := middleware.Chain(
		middleware.PanicRecovery,
		middleware.RequestID(log),
		middleware.Logging,
	)

	router.Handle("GET /domain/{id}", mwChain(mwErrHandler(domainHandler.GetItem())))
	router.Handle("POST /domain", mwChain(mwErrHandler(domainHandler.AddItem())))

	log.Info("starting server", slog.Any("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:    cfg.HTTPServer.Address,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("failed to start server", slog.Any("error", err))
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}

	return log
}
