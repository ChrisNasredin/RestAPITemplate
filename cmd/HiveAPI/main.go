package main

import (
	"HiveAPI/internal/config"
	"HiveAPI/internal/domain"
	"HiveAPI/internal/transport/rest"
	"HiveAPI/internal/transport/rest/middleware"
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
	log.Info("starting server")
	log.Debug("debug logging enabled")

	domainRepository := domain.NewRepository()
	domainService := domain.NewService(domainRepository)
	domainHandler := rest.NewHandler(domainService, log)

	router := http.NewServeMux()

	// Middleware
	mwErrHandler := middleware.ErrorHandler(domain.ErrorToHTTPStatus, log)
	mwChain := middleware.Chain(
		middleware.Logging(log),
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
