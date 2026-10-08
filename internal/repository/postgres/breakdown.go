package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/breakdown"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BreakdownRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewBreakdownRepository(pool *pgxpool.Pool, logger *slog.Logger) *BreakdownRepository {
	return &BreakdownRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *BreakdownRepository) Create(ctx context.Context, b *breakdown.Breakdown) (*breakdown.Breakdown, error) {
	query := `
		INSERT INTO breakdowns (bike_id, reason, cost, broken_at, is_fixed, fixed_at, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, bike_id, reason, cost, broken_at, is_fixed, fixed_at, comment, created_at
	`

	var created breakdown.Breakdown
	err := r.pool.QueryRow(ctx, query,
		b.BikeID, b.Reason, b.Cost, b.BrokenAt, b.IsFixed, b.FixedAt, b.Comment,
	).Scan(
		&created.ID,
		&created.BikeID,
		&created.Reason,
		&created.Cost,
		&created.BrokenAt,
		&created.IsFixed,
		&created.FixedAt,
		&created.Comment,
		&created.CreatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create breakdown", "error", err)
		return nil, fmt.Errorf("create breakdown: %w", err)
	}

	return &created, nil
}

func (r *BreakdownRepository) GetByID(ctx context.Context, id int64) (*breakdown.Breakdown, error) {
	query := `
		SELECT id, bike_id, reason, cost, broken_at, is_fixed, fixed_at, comment, created_at
		FROM breakdowns
		WHERE id = $1
	`

	var b breakdown.Breakdown
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.BikeID,
		&b.Reason,
		&b.Cost,
		&b.BrokenAt,
		&b.IsFixed,
		&b.FixedAt,
		&b.Comment,
		&b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, breakdown.ErrNotFound
		}
		r.logger.Error("failed to get breakdown", "id", id, "error", err)
		return nil, fmt.Errorf("get breakdown: %w", err)
	}

	return &b, nil
}

func (r *BreakdownRepository) List(ctx context.Context) ([]breakdown.Breakdown, error) {
	query := `
		SELECT id, bike_id, reason, cost, broken_at, is_fixed, fixed_at, comment, created_at
		FROM breakdowns
		ORDER BY broken_at DESC
	`

	return r.queryList(ctx, query)
}

func (r *BreakdownRepository) ListByBike(ctx context.Context, bikeID int64) ([]breakdown.Breakdown, error) {
	query := `
		SELECT id, bike_id, reason, cost, broken_at, is_fixed, fixed_at, comment, created_at
		FROM breakdowns
		WHERE bike_id = $1
		ORDER BY broken_at DESC
	`

	return r.queryList(ctx, query, bikeID)
}

func (r *BreakdownRepository) ListUnfixed(ctx context.Context) ([]breakdown.Breakdown, error) {
	query := `
		SELECT id, bike_id, reason, cost, broken_at, is_fixed, fixed_at, comment, created_at
		FROM breakdowns
		WHERE is_fixed = FALSE
		ORDER BY broken_at DESC
	`

	return r.queryList(ctx, query)
}

func (r *BreakdownRepository) Update(ctx context.Context, b *breakdown.Breakdown) (*breakdown.Breakdown, error) {
	query := `
		UPDATE breakdowns
		SET reason = $1, cost = $2, comment = $3
		WHERE id = $4
		RETURNING id, bike_id, reason, cost, broken_at, is_fixed, fixed_at, comment, created_at
	`

	var updated breakdown.Breakdown
	err := r.pool.QueryRow(ctx, query,
		b.Reason, b.Cost, b.Comment, b.ID,
	).Scan(
		&updated.ID,
		&updated.BikeID,
		&updated.Reason,
		&updated.Cost,
		&updated.BrokenAt,
		&updated.IsFixed,
		&updated.FixedAt,
		&updated.Comment,
		&updated.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, breakdown.ErrNotFound
		}
		r.logger.Error("failed to update breakdown", "id", b.ID, "error", err)
		return nil, fmt.Errorf("update breakdown: %w", err)
	}

	return &updated, nil
}

func (r *BreakdownRepository) MarkFixed(ctx context.Context, id int64) error {
	query := `
		UPDATE breakdowns
		SET is_fixed = TRUE, fixed_at = NOW()
		WHERE id = $1 AND is_fixed = FALSE
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to mark fixed", "id", id, "error", err)
		return fmt.Errorf("mark fixed: %w", err)
	}

	if result.RowsAffected() == 0 {
		exists, err := r.exists(ctx, id)
		if err != nil {
			return err
		}
		if !exists {
			return breakdown.ErrNotFound
		}
		return nil
	}

	return nil
}

func (r *BreakdownRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM breakdowns WHERE id = $1", id)
	if err != nil {
		r.logger.Error("failed to delete breakdown", "id", id, "error", err)
		return fmt.Errorf("delete breakdown: %w", err)
	}

	if result.RowsAffected() == 0 {
		return breakdown.ErrNotFound
	}

	return nil
}

func (r *BreakdownRepository) queryList(ctx context.Context, query string, args ...any) ([]breakdown.Breakdown, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list breakdowns", "error", err)
		return nil, fmt.Errorf("list breakdowns: %w", err)
	}
	defer rows.Close()

	result := make([]breakdown.Breakdown, 0)
	for rows.Next() {
		var b breakdown.Breakdown

		err := rows.Scan(
			&b.ID,
			&b.BikeID,
			&b.Reason,
			&b.Cost,
			&b.BrokenAt,
			&b.IsFixed,
			&b.FixedAt,
			&b.Comment,
			&b.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan breakdown", "error", err)
			return nil, fmt.Errorf("scan breakdown: %w", err)
		}

		result = append(result, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return result, nil
}

func (r *BreakdownRepository) exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM breakdowns WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check breakdown exists: %w", err)
	}
	return exists, nil
}
