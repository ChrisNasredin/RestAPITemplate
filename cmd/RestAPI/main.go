package main

import (
	"RestAPI/internal/config"
	"RestAPI/internal/domain"
	"RestAPI/internal/storage/postgres"
	"RestAPI/internal/transport/httpserver"
	"RestAPI/internal/transport/httpserver/middleware"
	"RestAPI/internal/transport/httpserver/observability"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	migrateFlag := flag.Bool("migrate", false, "run migrations and exit")
	configFlagPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg := config.MustLoad(*configFlagPath)

	log := setupLogger(cfg.Env)
	slog.SetDefault(log)

	// Migrations
	if *migrateFlag {
		fmt.Println("Starting migrations...")
		dsn := cfg.Storage.DSN()
		if err := postgres.RunMigrations(dsn); err != nil {
			log.Error("Migration failed: %w\n", slog.Any("error", err))
			os.Exit(1)
		}
		log.Info("Migrations finished successfully!")
		os.Exit(0) // Выход, чтобы не запускать сервер
	}

	log.Info("starting server")
	log.Debug("debug logging enabled")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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

	// Observability server
	obsRouter := http.NewServeMux()
	observability.NewHandler(obsRouter, storage)
	obsSrv := &http.Server{
		Addr:    cfg.Observability.Address,
		Handler: obsRouter,
	}
	go func() {
		if err := obsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start observability server", slog.Any("error", err))
		}
	}()

	// Service
	domainService := domain.NewService(storage)
	// Handlers
	domainHandler := httpserver.NewHandler(domainService)

	router := http.NewServeMux()

	// Middleware
	mwErrHandler := middleware.ErrorHandler(httpserver.ErrorToHTTPStatus)
	mwChain := middleware.Chain(
		middleware.PanicRecovery,
		middleware.Metrics,
		middleware.RequestID(log),
		middleware.Logging,
	)

	router.Handle("GET /domain/{id}", mwChain(mwErrHandler(domainHandler.GetItem())))
	router.Handle("POST /domain", mwChain(mwErrHandler(domainHandler.AddItem())))

	log.Info("starting server", slog.Any("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", slog.Any("error", err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("stopping server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", slog.Any("error", err))
	}
	if err = obsSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("observability server forced to shutdown", slog.Any("error", err))
	}
	storage.Close()
	log.Info("server stopped gracefully")
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
