package partner

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type Service struct {
	bikeRepo        BikeRepository
	rentalRepo      RentalRepository
	revenueProvider RevenueProvider
	partnerShare    float64
	logger          *slog.Logger
}

func NewService(
	bikeRepo BikeRepository,
	rentalRepo RentalRepository,
	revenueProvider RevenueProvider,
	partnerShare float64,
	logger *slog.Logger,
) *Service {
	return &Service{
		bikeRepo:        bikeRepo,
		rentalRepo:      rentalRepo,
		revenueProvider: revenueProvider,
		partnerShare:    partnerShare,
		logger:          logger,
	}
}

func (s *Service) Dashboard(ctx context.Context, month string) (*DashboardResult, error) {
	// 1. Парсим месяц → periodStart, periodEnd
	periodStart, periodEnd, monthKey, err := parseMonth(month)
	if err != nil {
		return nil, err
	}

	// 2. Статистика по велосипедам
	bikes, err := s.bikeRepo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list bikes", "error", err)
		return nil, fmt.Errorf("list bikes: %w", err)
	}

	bikesStats := BikesStats{Total: len(bikes)}
	for i := range bikes {
		switch {
		case bikes[i].IsBroken:
			bikesStats.Broken++
		case bikes[i].IsRented:
			bikesStats.Rented++
		default:
			bikesStats.Available++
		}
	}

	// 3. Статистика по активным арендам
	activeRentals, err := s.rentalRepo.ListActive(ctx)
	if err != nil {
		s.logger.Error("failed to list active rentals", "error", err)
		return nil, fmt.Errorf("list active rentals: %w", err)
	}

	now := time.Now().UTC()
	rentalsStats := RentalsStats{Active: len(activeRentals)}
	for i := range activeRentals {
		if activeRentals[i].IsOverdue(now) {
			rentalsStats.Overdue++
		}
	}

	// 4. Выручка за месяц
	revenue, err := s.revenueProvider.CalculateRevenue(ctx, periodStart, periodEnd)
	if err != nil {
		s.logger.Error("failed to calculate revenue", "month", month, "error", err)
		return nil, fmt.Errorf("calculate revenue: %w", err)
	}

	partnerShare := int64(float64(revenue) * s.partnerShare)

	s.logger.Info("partner dashboard calculated",
		"month", monthKey,
		"bikes_total", bikesStats.Total,
		"rentals_active", rentalsStats.Active,
		"revenue", revenue,
	)

	return &DashboardResult{
		Bikes:   bikesStats,
		Rentals: rentalsStats,
		CurrentMonth: MonthStats{
			Month:        monthKey,
			PeriodStart:  periodStart,
			PeriodEnd:    periodEnd,
			Revenue:      revenue,
			PartnerShare: partnerShare,
		},
	}, nil
}

// parseMonth — превращает "YYYY-MM" в periodStart, periodEnd (полуоткрытый) и месяц в том же формате.
func parseMonth(month string) (periodStart, periodEnd time.Time, monthKey string, err error) {
	month = strings.TrimSpace(month)
	if month == "" {
		// Если месяц не задан — берём текущий
		now := time.Now().UTC()
		month = now.Format("2006-01")
	}

	t, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, "", fmt.Errorf("month must be in YYYY-MM format: %w", err)
	}

	periodStart = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd = periodStart.AddDate(0, 1, 0)
	monthKey = periodStart.Format("2006-01")

	return periodStart, periodEnd, monthKey, nil
}
