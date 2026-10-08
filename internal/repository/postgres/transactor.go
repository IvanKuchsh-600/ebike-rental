package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	rentalusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/rental"
)

// RentalTransactor — реализация rentalusecase.Transactor.
// Создаёт репозитории, привязанные к транзакции.
type RentalTransactor struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewRentalTransactor(pool *pgxpool.Pool, logger *slog.Logger) *RentalTransactor {
	return &RentalTransactor{
		pool:   pool,
		logger: logger,
	}
}

func (t *RentalTransactor) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context, repos rentalusecase.Repos) error,
) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	repos := rentalusecase.Repos{
		Rentals: NewRentalRepositoryWithTx(tx, t.logger),
		Bikes:   NewBikeRepositoryWithTx(tx, t.logger),
	}

	if err := fn(ctx, repos); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// ============ Фабрики репозиториев ============

// NewRentalRepositoryWithTx — репозиторий аренд, привязанный к транзакции.
// Используется внутри RentalTransactor.
func NewRentalRepositoryWithTx(tx pgx.Tx, logger *slog.Logger) *RentalRepository {
	return &RentalRepository{db: tx, logger: logger}
}

// NewBikeRepositoryWithTx — репозиторий велосипедов, привязанный к транзакции.
func NewBikeRepositoryWithTx(tx pgx.Tx, logger *slog.Logger) *BikeRepository {
	return &BikeRepository{db: tx, logger: logger}
}
