package app

import (
	"RestAPI/internal/config"
	"RestAPI/internal/domain"
	"RestAPI/internal/metrics"
	"RestAPI/internal/storage/postgres"
	"RestAPI/internal/transport/httpserver"
	"RestAPI/internal/transport/httpserver/middleware"
	"RestAPI/internal/transport/httpserver/observability"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Database interface {
	ItemsCount(ctx context.Context) (int64, error)
	Close()
}
type App struct {
	cfg           *config.Config
	log           *slog.Logger
	storage       Database
	APISrv        *http.Server
	obsSrv        *http.Server
	stopCtxCancel context.CancelFunc
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
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
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	// Observability http-server
	obsRouter := http.NewServeMux()
	observability.NewHandler(obsRouter, storage)
	obsSrv := &http.Server{
		Addr:    cfg.Observability.Address,
		Handler: obsRouter,
	}
	// Инициализация бизнес-логики и хэндлеров
	domainService := domain.NewService(storage)
	domainHandler := httpserver.NewHandler(domainService)

	// Настройка роутинга и Middleware
	router := http.NewServeMux()
	mwErrHandler := middleware.ErrorHandler(httpserver.ErrorToHTTPStatus)
	mwChain := middleware.Chain(
		middleware.PanicRecovery,
		middleware.Metrics,
		middleware.RequestID(log),
		middleware.Logging,
	)

	// Добавляем ручки
	router.Handle("GET /domain/{id}", mwChain(mwErrHandler(domainHandler.GetItem())))
	router.Handle("POST /domain", mwChain(mwErrHandler(domainHandler.AddItem())))
	router.Handle("GET /domain", mwChain(mwErrHandler(domainHandler.GetAll())))
	router.Handle("PATCH /domain/{id}", mwChain(mwErrHandler(domainHandler.UpdateItem())))
	router.Handle("DELETE /domain/{id}", mwChain(mwErrHandler(domainHandler.DeleteItem())))

	// Инициализация основного HTTP-сервера
	mainSrv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	return &App{
		cfg:     cfg,
		log:     log,
		storage: storage,
		APISrv:  mainSrv,
		obsSrv:  obsSrv,
	}, nil
}

func (a *App) Run() {
	go metrics.TrackBusinessMetrics(a.storage, 15*time.Second)

	a.log.Info("starting Observability server", slog.String("address", a.cfg.Observability.Address))
	go func() {
		if err := a.obsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("failed to start observability server", slog.Any("error", err))
		}
	}()

	a.log.Info("starting API server", slog.String("address", a.cfg.HTTPServer.Address))
	go func() {
		if err := a.APISrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("failed to start server", slog.Any("error", err))
		}
	}()
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info("stopping HTTP server")

	if err := a.APISrv.Shutdown(ctx); err != nil {
		a.log.Error("server forced to shutdown", slog.Any("error", err))
	}

	// 2. Остановка observability сервера
	if err := a.obsSrv.Shutdown(ctx); err != nil {
		a.log.Error("observability server forced to shutdown", slog.Any("error", err))
	}

	// 3. Закрытие соединения с БД
	if a.storage != nil {
		a.storage.Close()
	}

	a.log.Info("server stopped gracefully")
}
