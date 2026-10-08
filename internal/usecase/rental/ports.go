package rental

import (
	"context"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/rental"
)

// ============ Repositories ============

type Repository interface {
	Create(ctx context.Context, r *rental.Rental) (*rental.Rental, error)
	GetByID(ctx context.Context, id int64) (*rental.Rental, error)
	List(ctx context.Context) ([]rental.Rental, error)
	ListActive(ctx context.Context) ([]rental.Rental, error)
	ListByBike(ctx context.Context, bikeID int64) ([]rental.Rental, error)
	Update(ctx context.Context, r *rental.Rental) (*rental.Rental, error)
	MarkReturned(ctx context.Context, id int64) error
	MarkPaid(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}

type BikeRepository interface {
	GetByIDForUpdate(ctx context.Context, id int64) (*bike.Bike, error)
	SetRented(ctx context.Context, id int64, rented bool) error
}

// ============ Transactor ============

// Repos — контейнер репозиториев, привязанных к одной транзакции.
// Usecase работает с ними как с обычными, но они уже внутри tx.
type Repos struct {
	Rentals Repository
	Bikes   BikeRepository
}

type Transactor interface {
	WithinTransaction(
		ctx context.Context,
		fn func(ctx context.Context, repos Repos) error,
	) error
}

// ============ Usecase ============

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*rental.Rental, error)
	GetByID(ctx context.Context, id int64) (*rental.Rental, error)
	List(ctx context.Context) ([]rental.Rental, error)
	ListActive(ctx context.Context) ([]rental.Rental, error)
	ListByBike(ctx context.Context, bikeID int64) ([]rental.Rental, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*rental.Rental, error)
	Return(ctx context.Context, id int64) (*rental.Rental, error)
	MarkPaid(ctx context.Context, id int64) (*rental.Rental, error)
	Delete(ctx context.Context, id int64) error
}

// ============ DTOs ============

type CreateInput struct {
	BikeID       int64
	RenterName   string
	RenterPhone  string
	StartedAt    *time.Time
	PlannedEndAt time.Time
	TotalAmount  int64
	Prepayment   *int64
	Deposit      *int64
	Comment      *string
}

type UpdateInput struct {
	RenterName   *string
	RenterPhone  *string
	PlannedEndAt *time.Time
	TotalAmount  *int64
	Comment      *string
}
