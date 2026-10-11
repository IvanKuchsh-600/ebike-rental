package auth

import (
	"context"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/user"
)

// UserRepository — контракт репозитория пользователей.
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
}

// TokenGenerator — контракт генератора токенов.
// Абстракция нужна, чтобы usecase не зависел от конкретной реализации (JWT).
type TokenGenerator interface {
	Generate(userID int64, email, role string) (string, error)
}

// Usecase — публичный контракт сервиса.
type Usecase interface {
	Login(ctx context.Context, input LoginInput) (*LoginResult, error)
	Me(ctx context.Context, userID int64) (*user.User, error)
	ChangePassword(ctx context.Context, userID int64, input ChangePasswordInput) error
}

// LoginInput — данные для входа.
type LoginInput struct {
	Email    string
	Password string
}

// LoginResult — результат логина.
type LoginResult struct {
	Token string
	User  *user.User
}

// ChangePasswordInput — данные для смены пароля.
type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
}
