package payout

import (
	"context"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/payout"
)

// Repository — контракт слоя хранения данных.
type Repository interface {
	Create(ctx context.Context, p *payout.Payout) (*payout.Payout, error)
	GetByID(ctx context.Context, id int64) (*payout.Payout, error)
	List(ctx context.Context) ([]payout.Payout, error)
	ExistsForPeriod(ctx context.Context, periodStart, periodEnd time.Time) (bool, error)
	MarkPaid(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	CalculateRevenue(ctx context.Context, periodStart, periodEnd time.Time) (int64, error)
}

// Usecase — публичный контракт сервиса.
type Usecase interface {
	Preview(ctx context.Context, month string) (*PreviewResult, error)
	Create(ctx context.Context, input CreateInput) (*payout.Payout, error)
	GetByID(ctx context.Context, id int64) (*payout.Payout, error)
	List(ctx context.Context) ([]payout.Payout, error)
	MarkPaid(ctx context.Context, id int64) (*payout.Payout, error)
	Delete(ctx context.Context, id int64) error
}

// PreviewResult — результат preview (не сохраняется).
type PreviewResult struct {
	PeriodStart  time.Time
	PeriodEnd    time.Time
	TotalRevenue int64
	PartnerShare int64
}

// CreateInput — данные для фиксации выплаты.
type CreateInput struct {
	Month   string // YYYY-MM
	Comment *string
}
