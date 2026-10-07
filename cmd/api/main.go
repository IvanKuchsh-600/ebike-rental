package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/config"
	"github.com/IvanKuchsh-600/ebike-rental/internal/http/handlers"
	"github.com/IvanKuchsh-600/ebike-rental/internal/http/router"
	"github.com/IvanKuchsh-600/ebike-rental/internal/repository/postgres"
	"github.com/IvanKuchsh-600/ebike-rental/internal/storage"
	bikeusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/bike"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := setupLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Пул соединений — один на всё приложение
	db, err := storage.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	logger.Info("connected to database")

	// 2. Репозитории
	bikeRepo := postgres.NewBikeRepository(db, logger)

	// 3. Usecases
	bikeSvc := bikeusecase.NewService(bikeRepo, logger)

	// 4. Handlers
	bikeHandler := handlers.NewBikeHandler(bikeSvc)

	handlersContainer := &router.Handlers{
		Bike: bikeHandler,
	}

	// 5. Router
	r := router.New(db, handlersContainer)

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		logger.Info("starting server", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("shutting down...")
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("server stopped")
	return nil
}

func setupLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
