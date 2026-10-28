package partner

import (
	"context"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/rental"
)

// BikeRepository — минимальный контракт для dashboard.
// Нам нужно только посчитать велосипеды по статусам.
type BikeRepository interface {
	List(ctx context.Context) ([]bike.Bike, error)
}

// RentalRepository — минимальный контракт для dashboard.
// Нужно посчитать активные и просроченные аренды.
type RentalRepository interface {
	ListActive(ctx context.Context) ([]rental.Rental, error)
}

// RevenueProvider — контракт для получения выручки за месяц.
// Реализуется payout-репозиторием.
type RevenueProvider interface {
	CalculateRevenue(ctx context.Context, periodStart, periodEnd time.Time) (int64, error)
}

// Usecase — публичный контракт.
type Usecase interface {
	Dashboard(ctx context.Context, month string) (*DashboardResult, error)
}

// DashboardResult — сводка для партнёра.
type DashboardResult struct {
	Bikes        BikesStats
	Rentals      RentalsStats
	CurrentMonth MonthStats
}

type BikesStats struct {
	Total     int
	Rented    int
	Broken    int
	Available int
}

type RentalsStats struct {
	Active  int
	Overdue int
}

type MonthStats struct {
	Month             string    // YYYY-MM
	PeriodStart       time.Time // начало периода
	PeriodEnd         time.Time // конец периода (полуоткрытый: не включается)
	Revenue           int64
	PartnerShare      int64
	AmortizationShare int64
}
