package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/rental"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RentalRepository struct {
	db     Querier
	logger *slog.Logger
}

func NewRentalRepository(pool *pgxpool.Pool, logger *slog.Logger) *RentalRepository {
	return &RentalRepository{
		db:     pool,
		logger: logger,
	}
}

// ============ Create ============

func (r *RentalRepository) Create(ctx context.Context, rr *rental.Rental) (*rental.Rental, error) {
	query := `
		INSERT INTO rentals (
			bike_id, renter_name, renter_phone,
			started_at, planned_end_at,
			total_amount, is_paid,
			prepayment, deposit,
			comment
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING
			id, bike_id, renter_name, renter_phone,
			started_at, planned_end_at, returned_at,
			total_amount, is_paid,
			prepayment, deposit, deposit_returned, is_deposit_returned,
			comment, created_at
	`

	var created rental.Rental
	err := r.db.QueryRow(ctx, query,
		rr.BikeID, rr.RenterName, rr.RenterPhone,
		rr.StartedAt, rr.PlannedEndAt,
		rr.TotalAmount, rr.IsPaid,
		rr.Prepayment, rr.Deposit,
		rr.Comment,
	).Scan(
		&created.ID, &created.BikeID, &created.RenterName, &created.RenterPhone,
		&created.StartedAt, &created.PlannedEndAt, &created.ReturnedAt,
		&created.TotalAmount, &created.IsPaid,
		&created.Prepayment, &created.Deposit, &created.DepositReturned, &created.IsDepositReturned,
		&created.Comment, &created.CreatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create rental", "error", err)
		return nil, fmt.Errorf("create rental: %w", err)
	}

	return &created, nil
}

// ============ Read ============

func (r *RentalRepository) GetByID(ctx context.Context, id int64) (*rental.Rental, error) {
	query := `
		SELECT
			id, bike_id, renter_name, renter_phone,
			started_at, planned_end_at, returned_at,
			total_amount, is_paid,
			prepayment, deposit, deposit_returned, is_deposit_returned,
			comment, created_at
		FROM rentals
		WHERE id = $1
	`

	var rr rental.Rental
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rr.ID, &rr.BikeID, &rr.RenterName, &rr.RenterPhone,
		&rr.StartedAt, &rr.PlannedEndAt, &rr.ReturnedAt,
		&rr.TotalAmount, &rr.IsPaid,
		&rr.Prepayment, &rr.Deposit, &rr.DepositReturned, &rr.IsDepositReturned,
		&rr.Comment, &rr.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, rental.ErrNotFound
		}
		r.logger.Error("failed to get rental", "id", id, "error", err)
		return nil, fmt.Errorf("get rental: %w", err)
	}

	return &rr, nil
}

func (r *RentalRepository) List(ctx context.Context) ([]rental.Rental, error) {
	query := `
		SELECT
			id, bike_id, renter_name, renter_phone,
			started_at, planned_end_at, returned_at,
			total_amount, is_paid,
			prepayment, deposit, deposit_returned, is_deposit_returned,
			comment, created_at
		FROM rentals
		ORDER BY started_at DESC
	`

	return r.queryList(ctx, query)
}

func (r *RentalRepository) ListActive(ctx context.Context) ([]rental.Rental, error) {
	query := `
		SELECT
			id, bike_id, renter_name, renter_phone,
			started_at, planned_end_at, returned_at,
			total_amount, is_paid,
			prepayment, deposit, deposit_returned, is_deposit_returned,
			comment, created_at
		FROM rentals
		WHERE returned_at IS NULL
		ORDER BY planned_end_at ASC
	`

	return r.queryList(ctx, query)
}

func (r *RentalRepository) ListByBike(ctx context.Context, bikeID int64) ([]rental.Rental, error) {
	query := `
		SELECT
			id, bike_id, renter_name, renter_phone,
			started_at, planned_end_at, returned_at,
			total_amount, is_paid,
			prepayment, deposit, deposit_returned, is_deposit_returned,
			comment, created_at
		FROM rentals
		WHERE bike_id = $1
		ORDER BY started_at DESC
	`

	return r.queryList(ctx, query, bikeID)
}

// ============ Update ============

func (r *RentalRepository) Update(ctx context.Context, rr *rental.Rental) (*rental.Rental, error) {
	query := `
		UPDATE rentals
		SET renter_name = $1, renter_phone = $2,
		    planned_end_at = $3, total_amount = $4,
		    comment = $5
		WHERE id = $6
		RETURNING
			id, bike_id, renter_name, renter_phone,
			started_at, planned_end_at, returned_at,
			total_amount, is_paid,
			prepayment, deposit, deposit_returned, is_deposit_returned,
			comment, created_at
	`

	var updated rental.Rental
	err := r.db.QueryRow(ctx, query,
		rr.RenterName, rr.RenterPhone,
		rr.PlannedEndAt, rr.TotalAmount,
		rr.Comment, rr.ID,
	).Scan(
		&updated.ID, &updated.BikeID, &updated.RenterName, &updated.RenterPhone,
		&updated.StartedAt, &updated.PlannedEndAt, &updated.ReturnedAt,
		&updated.TotalAmount, &updated.IsPaid,
		&updated.Prepayment, &updated.Deposit, &updated.DepositReturned, &updated.IsDepositReturned,
		&updated.Comment, &updated.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, rental.ErrNotFound
		}
		r.logger.Error("failed to update rental", "id", rr.ID, "error", err)
		return nil, fmt.Errorf("update rental: %w", err)
	}

	return &updated, nil
}

// MarkReturned — отметить аренду как возвращённую.
// Устанавливает returned_at = NOW(). Идемпотентен: повторный вызов — no-op.
func (r *RentalRepository) MarkReturned(ctx context.Context, id int64) error {
	query := `
		UPDATE rentals
		SET returned_at = NOW()
		WHERE id = $1 AND returned_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to mark returned", "id", id, "error", err)
		return fmt.Errorf("mark returned: %w", err)
	}

	if result.RowsAffected() == 0 {
		exists, err := r.exists(ctx, id)
		if err != nil {
			return err
		}
		if !exists {
			return rental.ErrNotFound
		}
		// Уже возвращена — не ошибка.
		return nil
	}

	return nil
}

// MarkPaid — отметить аренду как оплаченную.
func (r *RentalRepository) MarkPaid(ctx context.Context, id int64) error {
	query := `
		UPDATE rentals
		SET is_paid = TRUE
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to mark paid", "id", id, "error", err)
		return fmt.Errorf("mark paid: %w", err)
	}

	if result.RowsAffected() == 0 {
		return rental.ErrNotFound
	}

	return nil
}

// ============ Delete ============

func (r *RentalRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, "DELETE FROM rentals WHERE id = $1", id)
	if err != nil {
		r.logger.Error("failed to delete rental", "id", id, "error", err)
		return fmt.Errorf("delete rental: %w", err)
	}

	if result.RowsAffected() == 0 {
		return rental.ErrNotFound
	}

	return nil
}

// ============ helpers ============

func (r *RentalRepository) queryList(ctx context.Context, query string, args ...any) ([]rental.Rental, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list rentals", "error", err)
		return nil, fmt.Errorf("list rentals: %w", err)
	}
	defer rows.Close()

	result := make([]rental.Rental, 0)
	for rows.Next() {
		var rr rental.Rental
		if err := rows.Scan(
			&rr.ID, &rr.BikeID, &rr.RenterName, &rr.RenterPhone,
			&rr.StartedAt, &rr.PlannedEndAt, &rr.ReturnedAt,
			&rr.TotalAmount, &rr.IsPaid,
			&rr.Prepayment, &rr.Deposit, &rr.DepositReturned, &rr.IsDepositReturned,
			&rr.Comment, &rr.CreatedAt,
		); err != nil {
			r.logger.Error("failed to scan rental", "error", err)
			return nil, fmt.Errorf("scan rental: %w", err)
		}
		result = append(result, rr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return result, nil
}

func (r *RentalRepository) exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM rentals WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check rental exists: %w", err)
	}
	return exists, nil
}
