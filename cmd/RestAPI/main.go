package main

import (
	"RestAPI/internal/config"
	"RestAPI/internal/domain"
	"RestAPI/internal/transport/http-server"
	"RestAPI/internal/transport/http-server/middleware"
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

	domainRepository := domain.NewRepository()
	domainService := domain.NewService(domainRepository)
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
