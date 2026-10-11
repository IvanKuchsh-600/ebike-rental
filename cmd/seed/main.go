package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/IvanKuchsh-600/ebike-rental/internal/auth"
	"github.com/IvanKuchsh-600/ebike-rental/internal/config"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/user"
	"github.com/IvanKuchsh-600/ebike-rental/internal/repository/postgres"
	"github.com/IvanKuchsh-600/ebike-rental/internal/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ============ Флаги ============
	ownerEmail := flag.String("owner-email", "", "Owner email (required)")
	ownerPassword := flag.String("owner-password", "", "Owner password (required, min 8 chars)")
	ownerName := flag.String("owner-name", "Owner", "Owner display name")

	partnerEmail := flag.String("partner-email", "", "Partner email (optional)")
	partnerPassword := flag.String("partner-password", "", "Partner password (optional, min 8 chars)")
	partnerName := flag.String("partner-name", "Partner", "Partner display name")

	flag.Parse()

	if *ownerEmail == "" || *ownerPassword == "" {
		flag.Usage()
		return errors.New("owner-email and owner-password are required")
	}
	if len(*ownerPassword) < 8 {
		return errors.New("owner-password must be at least 8 characters")
	}

	// ============ Конфиг и БД ============
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx := context.Background()

	db, err := storage.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer db.Close()

	repo := postgres.NewUserRepository(db, logger)

	// ============ Создаём owner ============
	err = upsertUser(ctx, repo, logger, user.RoleOwner, *ownerEmail, *ownerPassword, *ownerName)
	if err != nil {
		return fmt.Errorf("create owner: %w", err)
	}

	// ============ Создаём partner (если указан) ============
	if *partnerEmail != "" || *partnerPassword != "" {
		if *partnerEmail == "" || *partnerPassword == "" {
			return errors.New("both partner-email and partner-password must be set, or neither")
		}
		if len(*partnerPassword) < 8 {
			return errors.New("partner-password must be at least 8 characters")
		}

		if err := upsertUser(ctx, repo, logger, user.RolePartner, *partnerEmail, *partnerPassword, *partnerName); err != nil {
			return fmt.Errorf("create partner: %w", err)
		}
	}

	logger.Info("seed completed successfully")
	return nil
}

// upsertUser — создаёт или обновляет пользователя по email.
func upsertUser(
	ctx context.Context,
	repo *postgres.UserRepository,
	logger *slog.Logger,
	role, email, password, name string,
) error {
	// Нормализуем email (как в usecase)
	email = normalizeEmail(email)

	// Проверяем, есть ли уже
	existing, err := repo.GetByEmail(ctx, email)
	if err == nil {
		// Пользователь существует — обновляем пароль и имя
		logger.Info("user already exists, updating password", "email", email, "role", existing.Role)

		newHash, err := auth.HashPassword(password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		if err := repo.UpdatePassword(ctx, existing.ID, newHash); err != nil {
			return fmt.Errorf("update password: %w", err)
		}

		logger.Info("user updated", "id", existing.ID, "email", email)
		return nil
	}

	if !errors.Is(err, user.ErrNotFound) {
		return fmt.Errorf("check existing user: %w", err)
	}

	// Пользователя нет — создаём
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	namePtr := &name
	u := &user.User{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		Name:         namePtr,
	}

	created, err := repo.Create(ctx, u)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	logger.Info("user created", "id", created.ID, "email", email, "role", role)
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
