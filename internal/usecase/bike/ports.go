package bike

import (
	"context"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
)

// Repository — контракт, который должен реализовать слой хранения данных.
type Repository interface {
	Create(ctx context.Context, b *bike.Bike) (*bike.Bike, error)
	GetByID(ctx context.Context, id int64) (*bike.Bike, error)
	List(ctx context.Context) ([]bike.Bike, error)
	Update(ctx context.Context, b *bike.Bike) (*bike.Bike, error)
	Delete(ctx context.Context, id int64) error
	SetRented(ctx context.Context, id int64, rented bool) error
	SetBroken(ctx context.Context, id int64, broken bool) error
}

// Usecase — публичный контракт сервиса.
type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*bike.Bike, error)
	GetByID(ctx context.Context, id int64) (*bike.Bike, error)
	List(ctx context.Context) ([]bike.Bike, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*bike.Bike, error)
	Delete(ctx context.Context, id int64) error
	SetRented(ctx context.Context, id int64, rented bool) error
	SetBroken(ctx context.Context, id int64, broken bool) error
}

type CreateInput struct {
	SerialNumber string
	Model        *string
	Comment      *string
}

type UpdateInput struct {
	SerialNumber *string
	Model        *string
	Comment      *string
}
