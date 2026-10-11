package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db     Querier
	logger *slog.Logger
}

func NewUserRepository(pool *pgxpool.Pool, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     pool,
		logger: logger,
	}
}

// ============ Create ============

func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	query := `
		INSERT INTO users (email, password_hash, role, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, role, name, created_at, updated_at
	`

	var created user.User
	err := r.db.QueryRow(ctx, query,
		u.Email, u.PasswordHash, u.Role, u.Name,
	).Scan(
		&created.ID,
		&created.Email,
		&created.PasswordHash,
		&created.Role,
		&created.Name,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		// нарушение UNIQUE по email?
		if isUniqueViolation(err) {
			return nil, user.ErrEmailAlreadyExists
		}
		r.logger.Error("failed to create user", "email", u.Email, "error", err)
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, role, name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var u user.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.Name,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		r.logger.Error("failed to get user", "id", id, "error", err)
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, role, name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var u user.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.Name,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		r.logger.Error("failed to get user by email", "email", email, "error", err)
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) List(ctx context.Context) ([]user.User, error) {
	query := `
		SELECT id, email, password_hash, role, name, created_at, updated_at
		FROM users
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to list users", "error", err)
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	result := make([]user.User, 0)
	for rows.Next() {
		var u user.User
		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.PasswordHash,
			&u.Role,
			&u.Name,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan user", "error", err)
			return nil, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return result, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, passwordHash, id)
	if err != nil {
		r.logger.Error("failed to update password", "id", id, "error", err)
		return fmt.Errorf("update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		r.logger.Error("failed to delete user", "id", id, "error", err)
		return fmt.Errorf("delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return user.ErrNotFound
	}

	return nil
}

// isUniqueViolation — проверяет, что ошибка — нарушение UNIQUE constraint.
// pgx возвращает ошибку типа *pgconn.PgError с Code = "23505".
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
