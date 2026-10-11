package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/IvanKuchsh-600/ebike-rental/internal/auth"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/user"
)

type Service struct {
	repo   UserRepository
	tokens TokenGenerator
	logger *slog.Logger
}

func NewService(repo UserRepository, tokens TokenGenerator, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		tokens: tokens,
		logger: logger,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", user.ErrInvalidInput)
	}
	if input.Password == "" {
		return nil, fmt.Errorf("%w: password is required", user.ErrInvalidInput)
	}

	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			s.logger.Warn("login attempt with unknown email", "email", email)
			return nil, user.ErrInvalidCredentials
		}
		s.logger.Error("failed to get user by email", "email", email, "error", err)
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Проверяем пароль
	if !auth.CheckPassword(u.PasswordHash, input.Password) {
		s.logger.Warn("login attempt with wrong password", "email", email)
		return nil, user.ErrInvalidCredentials
	}

	// Генерируем токен
	token, err := s.tokens.Generate(u.ID, u.Email, u.Role)
	if err != nil {
		s.logger.Error("failed to generate token", "user_id", u.ID, "error", err)
		return nil, fmt.Errorf("generate token: %w", err)
	}

	s.logger.Info("user logged in", "user_id", u.ID, "email", u.Email, "role", u.Role)

	return &LoginResult{
		Token: token,
		User:  u,
	}, nil
}

func (s *Service) Me(ctx context.Context, userID int64) (*user.User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user_id", user.ErrInvalidInput)
	}

	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user", "user_id", userID, "error", err)
		return nil, err
	}
	return u, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, input ChangePasswordInput) error {
	if userID <= 0 {
		return fmt.Errorf("%w: invalid user_id", user.ErrInvalidInput)
	}
	if len(input.NewPassword) < 8 {
		return fmt.Errorf("%w: new password must be at least 8 characters", user.ErrInvalidInput)
	}
	if input.OldPassword == input.NewPassword {
		return fmt.Errorf("%w: new password must be different", user.ErrInvalidInput)
	}

	// Загружаем пользователя
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Проверяем старый пароль
	if !auth.CheckPassword(u.PasswordHash, input.OldPassword) {
		s.logger.Warn("change password attempt with wrong old password", "user_id", userID)
		return user.ErrInvalidCredentials
	}

	// Хешируем новый
	newHash, err := auth.HashPassword(input.NewPassword)
	if err != nil {
		s.logger.Error("failed to hash new password", "user_id", userID, "error", err)
		return fmt.Errorf("hash password: %w", err)
	}

	// Сохраняем
	err = s.repo.UpdatePassword(ctx, userID, newHash)
	if err != nil {
		s.logger.Error("failed to update password", "user_id", userID, "error", err)
		return fmt.Errorf("update password: %w", err)
	}

	s.logger.Info("password changed", "user_id", userID)
	return nil
}
