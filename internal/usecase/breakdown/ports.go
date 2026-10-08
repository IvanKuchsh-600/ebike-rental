package breakdown

import (
	"context"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/breakdown"
)

// Repository — контракт слоя хранения данных.
type Repository interface {
	Create(ctx context.Context, b *breakdown.Breakdown) (*breakdown.Breakdown, error)
	GetByID(ctx context.Context, id int64) (*breakdown.Breakdown, error)
	List(ctx context.Context) ([]breakdown.Breakdown, error)
	ListByBike(ctx context.Context, bikeID int64) ([]breakdown.Breakdown, error)
	ListUnfixed(ctx context.Context) ([]breakdown.Breakdown, error)
	Update(ctx context.Context, b *breakdown.Breakdown) (*breakdown.Breakdown, error)
	MarkFixed(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}

// Usecase — публичный контракт сервиса.
type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*breakdown.Breakdown, error)
	GetByID(ctx context.Context, id int64) (*breakdown.Breakdown, error)
	List(ctx context.Context) ([]breakdown.Breakdown, error)
	ListByBike(ctx context.Context, bikeID int64) ([]breakdown.Breakdown, error)
	ListUnfixed(ctx context.Context) ([]breakdown.Breakdown, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*breakdown.Breakdown, error)
	MarkFixed(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}

// CreateInput — данные для создания поломки.
type CreateInput struct {
	BikeID   int64
	Reason   string
	Cost     int64
	BrokenAt *time.Time // если не задано — текущее время на стороне БД
	Comment  *string
}

// UpdateInput — частичное обновление. nil = «не трогать».
type UpdateInput struct {
	Reason  *string
	Cost    *int64
	Comment *string
}
