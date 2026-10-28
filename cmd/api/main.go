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

	"github.com/IvanKuchsh-600/ebike-rental/internal/auth"

	"github.com/IvanKuchsh-600/ebike-rental/internal/config"
	"github.com/IvanKuchsh-600/ebike-rental/internal/http/handlers"
	"github.com/IvanKuchsh-600/ebike-rental/internal/http/router"
	"github.com/IvanKuchsh-600/ebike-rental/internal/repository/postgres"
	"github.com/IvanKuchsh-600/ebike-rental/internal/storage"
	authusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/auth"
	bikeusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/bike"
	breakdownusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/breakdown"
	partnerusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/partner"
	payoutusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/payout"
	rentalusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/rental"
	webpkg "github.com/IvanKuchsh-600/ebike-rental/internal/web"
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

	// 1. Пул соединений
	db, err := storage.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	logger.Info("connected to database")

	// 2. Репозитории
	bikeRepo := postgres.NewBikeRepository(db, logger)
	breakdownRepo := postgres.NewBreakdownRepository(db, logger)
	rentalRepo := postgres.NewRentalRepository(db, logger)
	payoutRepo := postgres.NewPayoutRepository(db, logger)
	userRepo := postgres.NewUserRepository(db, logger)

	// 3. Транзактор для rentals
	rentalTransactor := postgres.NewRentalTransactor(db, logger)

	// 4. Генератор токенов
	tokenGen := auth.NewJWTGenerator(cfg.JWTSecret, cfg.JWTExpiryHours)

	// 5. Usecases
	bikeSvc := bikeusecase.NewService(bikeRepo, logger)
	breakdownSvc := breakdownusecase.NewService(breakdownRepo, bikeRepo, logger)
	rentalSvc := rentalusecase.NewService(rentalRepo, bikeRepo, rentalTransactor, logger)
	payoutSvc := payoutusecase.NewService(
		payoutRepo,
		cfg.OwnerShare,
		cfg.PartnerShare,
		cfg.AmortizationShare,
		logger,
	)
	partnerSvc := partnerusecase.NewService(
		bikeRepo,
		rentalRepo,
		payoutRepo,
		cfg.PartnerShare,
		cfg.AmortizationShare,
		logger,
	)
	authSvc := authusecase.NewService(userRepo, tokenGen, logger)

	// 6. Handlers
	bikeHandler := handlers.NewBikeHandler(bikeSvc)
	breakdownHandler := handlers.NewBreakdownHandler(breakdownSvc)
	rentalHandler := handlers.NewRentalHandler(rentalSvc)
	payoutHandler := handlers.NewPayoutHandler(payoutSvc)
	authHandler := handlers.NewAuthHandler(authSvc)
	partnerHandler := handlers.NewPartnerHandler(
		partnerSvc,
		bikeSvc,
		rentalSvc,
		breakdownSvc,
		payoutSvc,
	)
	webRenderer := webpkg.NewRenderer(logger)
	webHandler := webpkg.NewHandler(webRenderer, logger)

	handlersContainer := &router.Handlers{
		Bike:      bikeHandler,
		Breakdown: breakdownHandler,
		Rental:    rentalHandler,
		Payout:    payoutHandler,
		Auth:      authHandler,
		Partner:   partnerHandler,
		Web:       webHandler,
	}

	// 7. Router
	routerCfg := router.Config{
		JWTSecret: cfg.JWTSecret,
	}
	r := router.New(db, handlersContainer, routerCfg)

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
