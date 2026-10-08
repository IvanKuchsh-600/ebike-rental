package payout

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/payout"
)

type Service struct {
	repo         Repository
	partnerShare float64 // например, 0.10
	logger       *slog.Logger
}

func NewService(repo Repository, partnerShare float64, logger *slog.Logger) *Service {
	return &Service{
		repo:         repo,
		partnerShare: partnerShare,
		logger:       logger,
	}
}

// ============ Preview ============

// Preview — посчитать долг партнёру за месяц БЕЗ сохранения.
func (s *Service) Preview(ctx context.Context, month string) (*PreviewResult, error) {
	periodStart, periodEnd, err := parseMonth(month)
	if err != nil {
		return nil, err
	}

	revenue, err := s.repo.CalculateRevenue(ctx, periodStart, periodEnd)
	if err != nil {
		s.logger.Error("failed to calculate revenue", "month", month, "error", err)
		return nil, fmt.Errorf("calculate revenue: %w", err)
	}

	partnerShare := calculatePartnerShare(revenue, s.partnerShare)

	s.logger.Info("payout preview calculated",
		"month", month,
		"revenue", revenue,
		"partner_share", partnerShare,
	)

	return &PreviewResult{
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		TotalRevenue: revenue,
		PartnerShare: partnerShare,
	}, nil
}

// ============ Create ============

// Create — зафиксировать выплату за месяц.
func (s *Service) Create(ctx context.Context, input CreateInput) (*payout.Payout, error) {
	periodStart, periodEnd, err := parseMonth(input.Month)
	if err != nil {
		return nil, err
	}

	// Проверить, что payout за этот период ещё нет
	exists, err := s.repo.ExistsForPeriod(ctx, periodStart, periodEnd)
	if err != nil {
		s.logger.Error("failed to check payout existence", "month", input.Month, "error", err)
		return nil, fmt.Errorf("check payout existence: %w", err)
	}
	if exists {
		return nil, payout.ErrAlreadyExists
	}

	// Посчитать выручку
	revenue, err := s.repo.CalculateRevenue(ctx, periodStart, periodEnd)
	if err != nil {
		s.logger.Error("failed to calculate revenue", "month", input.Month, "error", err)
		return nil, fmt.Errorf("calculate revenue: %w", err)
	}

	partnerShare := calculatePartnerShare(revenue, s.partnerShare)

	p := &payout.Payout{
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		TotalRevenue: revenue,
		PartnerShare: partnerShare,
		IsPaid:       false,
		Comment:      input.Comment,
	}

	created, err := s.repo.Create(ctx, p)
	if err != nil {
		s.logger.Error("failed to create payout", "month", input.Month, "error", err)
		return nil, fmt.Errorf("create payout: %w", err)
	}

	s.logger.Info("payout created",
		"id", created.ID,
		"month", input.Month,
		"revenue", created.TotalRevenue,
		"partner_share", created.PartnerShare,
	)
	return created, nil
}

// ============ Read ============

func (s *Service) GetByID(ctx context.Context, id int64) (*payout.Payout, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", payout.ErrInvalidInput)
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get payout", "id", id, "error", err)
		return nil, err
	}
	return p, nil
}

func (s *Service) List(ctx context.Context) ([]payout.Payout, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list payouts", "error", err)
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	return items, nil
}

// ============ MarkPaid ============

func (s *Service) MarkPaid(ctx context.Context, id int64) (*payout.Payout, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", payout.ErrInvalidInput)
	}

	if err := s.repo.MarkPaid(ctx, id); err != nil {
		s.logger.Error("failed to mark payout paid", "id", id, "error", err)
		return nil, fmt.Errorf("mark payout paid: %w", err)
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get updated payout: %w", err)
	}

	s.logger.Info("payout marked as paid", "id", id)
	return updated, nil
}

// ============ Delete ============

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", payout.ErrInvalidInput)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, payout.ErrNotFound) {
			return err
		}
		s.logger.Error("failed to delete payout", "id", id, "error", err)
		return fmt.Errorf("delete payout: %w", err)
	}

	s.logger.Info("payout deleted", "id", id)
	return nil
}

// ============ helpers ============

// parseMonth — превращает "YYYY-MM" в начало и конец месяца.
// periodEnd — это НАЧАЛО СЛЕДУЮЩЕГО месяца (полуоткрытый интервал).
func parseMonth(month string) (periodStart, periodEnd time.Time, err error) {
	month = strings.TrimSpace(month)
	if month == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: month is required", payout.ErrInvalidInput)
	}

	// Парсим "2006-01" — это YYYY-MM в Go-формате
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: month must be in YYYY-MM format", payout.ErrInvalidInput)
	}

	periodStart = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd = periodStart.AddDate(0, 1, 0)

	return periodStart, periodEnd, nil
}

// calculatePartnerShare — 10% от выручки, округление до целого рубля.
func calculatePartnerShare(revenue int64, share float64) int64 {
	return int64(float64(revenue) * share)
}
