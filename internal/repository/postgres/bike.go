package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BikeRepository struct {
	db     Querier
	logger *slog.Logger
}

func NewBikeRepository(pool *pgxpool.Pool, logger *slog.Logger) *BikeRepository {
	return &BikeRepository{
		db:     pool,
		logger: logger,
	}
}

func (r *BikeRepository) Create(ctx context.Context, b *bike.Bike) (*bike.Bike, error) {
	query := `
		INSERT INTO bikes (serial_number, model, comment)
		VALUES ($1, $2, $3)
		RETURNING id, serial_number, model, is_rented, is_broken, comment, created_at
	`

	var created bike.Bike
	err := r.db.QueryRow(ctx, query,
		b.SerialNumber, b.Model, b.Comment,
	).Scan(
		&created.ID,
		&created.SerialNumber,
		&created.Model,
		&created.IsRented,
		&created.IsBroken,
		&created.Comment,
		&created.CreatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create bike", "error", err)
		return nil, fmt.Errorf("create bike: %w", err)
	}

	return &created, nil
}

func (r *BikeRepository) GetByID(ctx context.Context, id int64) (*bike.Bike, error) {
	query := `
		SELECT id, serial_number, model, is_rented, is_broken, comment, created_at
		FROM bikes
		WHERE id = $1
	`

	var b bike.Bike
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.SerialNumber,
		&b.Model,
		&b.IsRented,
		&b.IsBroken,
		&b.Comment,
		&b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, bike.ErrNotFound
		}
		r.logger.Error("failed to get bike", "id", id, "error", err)
		return nil, fmt.Errorf("get bike: %w", err)
	}

	return &b, nil
}

func (r *BikeRepository) List(ctx context.Context) ([]bike.Bike, error) {
	query := `
		SELECT id, serial_number, model, is_rented, is_broken, comment, created_at
		FROM bikes
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to list bikes", "error", err)
		return nil, fmt.Errorf("list bikes: %w", err)
	}
	defer rows.Close()

	bikes := make([]bike.Bike, 0)
	for rows.Next() {
		var b bike.Bike

		err := rows.Scan(
			&b.ID,
			&b.SerialNumber,
			&b.Model,
			&b.IsRented,
			&b.IsBroken,
			&b.Comment,
			&b.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan bike", "error", err)
			return nil, fmt.Errorf("scan bike: %w", err)
		}
		bikes = append(bikes, b)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return bikes, nil
}

func (r *BikeRepository) Update(ctx context.Context, b *bike.Bike) (*bike.Bike, error) {
	query := `
		UPDATE bikes
		SET serial_number = $1, model = $2, comment = $3
		WHERE id = $4
		RETURNING id, serial_number, model, is_rented, is_broken, comment, created_at
	`

	var updated bike.Bike
	err := r.db.QueryRow(ctx, query,
		b.SerialNumber, b.Model, b.Comment, b.ID,
	).Scan(
		&updated.ID,
		&updated.SerialNumber,
		&updated.Model,
		&updated.IsRented,
		&updated.IsBroken,
		&updated.Comment,
		&updated.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, bike.ErrNotFound
		}
		r.logger.Error("failed to update bike", "id", b.ID, "error", err)
		return nil, fmt.Errorf("update bike: %w", err)
	}

	return &updated, nil
}

func (r *BikeRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, "DELETE FROM bikes WHERE id = $1", id)
	if err != nil {
		r.logger.Error("failed to delete bike", "id", id, "error", err)
		return fmt.Errorf("delete bike: %w", err)
	}

	if result.RowsAffected() == 0 {
		return bike.ErrNotFound
	}

	return nil
}

// SetRented — отметить велосипед как сданный/возвращённый.
func (r *BikeRepository) SetRented(ctx context.Context, id int64, rented bool) error {
	result, err := r.db.Exec(ctx,
		"UPDATE bikes SET is_rented = $1 WHERE id = $2",
		rented, id,
	)
	if err != nil {
		r.logger.Error("failed to set rented", "id", id, "error", err)
		return fmt.Errorf("set rented: %w", err)
	}

	if result.RowsAffected() == 0 {
		return bike.ErrNotFound
	}

	return nil
}

// SetBroken — отметить велосипед как сломанный/починенный.
func (r *BikeRepository) SetBroken(ctx context.Context, id int64, broken bool) error {
	result, err := r.db.Exec(ctx,
		"UPDATE bikes SET is_broken = $1 WHERE id = $2",
		broken, id,
	)
	if err != nil {
		r.logger.Error("failed to set broken", "id", id, "error", err)
		return fmt.Errorf("set broken: %w", err)
	}

	if result.RowsAffected() == 0 {
		return bike.ErrNotFound
	}

	return nil
}

// GetByIDForUpdate — SELECT ... FOR UPDATE, блокирует строку на время транзакции.
// Вне транзакции использовать НЕЛЬЗЯ (pgx вернёт ошибку).
func (r *BikeRepository) GetByIDForUpdate(ctx context.Context, id int64) (*bike.Bike, error) {
	query := `
		SELECT id, serial_number, model, is_rented, is_broken, comment, created_at
		FROM bikes
		WHERE id = $1
		FOR UPDATE
	`

	var b bike.Bike
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.SerialNumber, &b.Model,
		&b.IsRented, &b.IsBroken, &b.Comment, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, bike.ErrNotFound
		}
		r.logger.Error("failed to get bike for update", "id", id, "error", err)
		return nil, fmt.Errorf("get bike for update: %w", err)
	}

	return &b, nil
}
