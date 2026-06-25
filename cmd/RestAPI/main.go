package main

import (
	"RestAPI/internal/app"
	"RestAPI/internal/config"
	"RestAPI/internal/storage/postgres"
	"context"
	"flag"
	"fmt"
	"log/slog"
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

	// Migrations (WARNING! User in dsn must have DDL premissions for --migrate options)
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
		panic("Failed to initialize storage: " + err.Error())
	}

	application, err := app.New(cfg, log, storage)
	if err != nil {
		log.Error("app initialization failed: %w", slog.Any("error", err))
	}
	application.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	application.Stop(shutdownCtx)
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
