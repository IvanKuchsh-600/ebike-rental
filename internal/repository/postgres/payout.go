package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/payout"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PayoutRepository struct {
	db     Querier
	logger *slog.Logger
}

func NewPayoutRepository(pool *pgxpool.Pool, logger *slog.Logger) *PayoutRepository {
	return &PayoutRepository{
		db:     pool,
		logger: logger,
	}
}

// ============ Create ============

func (r *PayoutRepository) Create(ctx context.Context, p *payout.Payout) (*payout.Payout, error) {
	query := `
		INSERT INTO payouts (period_start, period_end, total_revenue, partner_share, is_paid, paid_at, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, period_start, period_end, total_revenue, partner_share, is_paid, paid_at, comment, created_at
	`

	var created payout.Payout
	err := r.db.QueryRow(ctx, query,
		p.PeriodStart, p.PeriodEnd, p.TotalRevenue, p.PartnerShare, p.IsPaid, p.PaidAt, p.Comment,
	).Scan(
		&created.ID,
		&created.PeriodStart,
		&created.PeriodEnd,
		&created.TotalRevenue,
		&created.PartnerShare,
		&created.IsPaid,
		&created.PaidAt,
		&created.Comment,
		&created.CreatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create payout", "error", err)
		return nil, fmt.Errorf("create payout: %w", err)
	}

	return &created, nil
}

// ============ Read ============

func (r *PayoutRepository) GetByID(ctx context.Context, id int64) (*payout.Payout, error) {
	query := `
		SELECT id, period_start, period_end, total_revenue, partner_share, is_paid, paid_at, comment, created_at
		FROM payouts
		WHERE id = $1
	`

	var p payout.Payout
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.PeriodStart,
		&p.PeriodEnd,
		&p.TotalRevenue,
		&p.PartnerShare,
		&p.IsPaid,
		&p.PaidAt,
		&p.Comment,
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, payout.ErrNotFound
		}
		r.logger.Error("failed to get payout", "id", id, "error", err)
		return nil, fmt.Errorf("get payout: %w", err)
	}

	return &p, nil
}

func (r *PayoutRepository) List(ctx context.Context) ([]payout.Payout, error) {
	query := `
		SELECT id, period_start, period_end, total_revenue, partner_share, is_paid, paid_at, comment, created_at
		FROM payouts
		ORDER BY period_start DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to list payouts", "error", err)
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	defer rows.Close()

	result := make([]payout.Payout, 0)
	for rows.Next() {
		var p payout.Payout
		if err := rows.Scan(
			&p.ID,
			&p.PeriodStart,
			&p.PeriodEnd,
			&p.TotalRevenue,
			&p.PartnerShare,
			&p.IsPaid,
			&p.PaidAt,
			&p.Comment,
			&p.CreatedAt,
		); err != nil {
			r.logger.Error("failed to scan payout", "error", err)
			return nil, fmt.Errorf("scan payout: %w", err)
		}
		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return result, nil
}

// ExistsForPeriod — есть ли уже payout за этот период.
// Используется при создании, чтобы вернуть 409 Conflict.
func (r *PayoutRepository) ExistsForPeriod(ctx context.Context, periodStart, periodEnd time.Time) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM payouts
			WHERE period_start = $1 AND period_end = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, periodStart, periodEnd).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check payout exists: %w", err)
	}
	return exists, nil
}

// ============ Update ============

// MarkPaid — отметить payout как выплаченный.
// Идемпотентен: повторный вызов на уже оплаченном — no-op.
func (r *PayoutRepository) MarkPaid(ctx context.Context, id int64) error {
	query := `
		UPDATE payouts
		SET is_paid = TRUE, paid_at = NOW()
		WHERE id = $1 AND is_paid = FALSE
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to mark payout paid", "id", id, "error", err)
		return fmt.Errorf("mark payout paid: %w", err)
	}

	if result.RowsAffected() == 0 {
		exists, err := r.exists(ctx, id)
		if err != nil {
			return err
		}
		if !exists {
			return payout.ErrNotFound
		}
		// Уже выплачен — не ошибка, no-op.
		return nil
	}

	return nil
}

// ============ Delete ============

func (r *PayoutRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, "DELETE FROM payouts WHERE id = $1", id)
	if err != nil {
		r.logger.Error("failed to delete payout", "id", id, "error", err)
		return fmt.Errorf("delete payout: %w", err)
	}

	if result.RowsAffected() == 0 {
		return payout.ErrNotFound
	}

	return nil
}

// ============ Aggregate ============

// CalculateRevenue — сумма total_amount всех ОПЛАЧЕННЫХ аренд за период.
// Период полуоткрытый: [periodStart, periodEnd).
func (r *PayoutRepository) CalculateRevenue(
	ctx context.Context,
	periodStart, periodEnd time.Time,
) (int64, error) {
	query := `
		SELECT COALESCE(SUM(total_amount), 0)
		FROM rentals
		WHERE is_paid = TRUE
		  AND started_at >= $1
		  AND started_at <  $2
	`

	var revenue int64
	err := r.db.QueryRow(ctx, query, periodStart, periodEnd).Scan(&revenue)
	if err != nil {
		r.logger.Error("failed to calculate revenue",
			"period_start", periodStart,
			"period_end", periodEnd,
			"error", err,
		)
		return 0, fmt.Errorf("calculate revenue: %w", err)
	}

	return revenue, nil
}

// ============ helpers ============

func (r *PayoutRepository) exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM payouts WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check payout exists: %w", err)
	}
	return exists, nil
}
