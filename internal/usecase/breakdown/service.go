package breakdown

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/breakdown"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*breakdown.Breakdown, error) {
	if err := validateCreateInput(input); err != nil {
		s.logger.Warn("validation failed", "error", err)
		return nil, err
	}

	// Если BrokenAt не задан — оставляем как zero time,
	// репозиторий/БД подставит NOW().
	b := &breakdown.Breakdown{
		BikeID:  input.BikeID,
		Reason:  strings.TrimSpace(input.Reason),
		Cost:    input.Cost,
		Comment: input.Comment,
	}
	if input.BrokenAt != nil {
		b.BrokenAt = *input.BrokenAt
	} else {
		b.BrokenAt = time.Now().UTC() // ← вот это
	}

	created, err := s.repo.Create(ctx, b)
	if err != nil {
		s.logger.Error("failed to create breakdown", "bike_id", input.BikeID, "error", err)
		return nil, fmt.Errorf("create breakdown: %w", err)
	}

	s.logger.Info("breakdown created", "id", created.ID, "bike_id", created.BikeID)
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*breakdown.Breakdown, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", breakdown.ErrInvalidInput)
	}

	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get breakdown", "id", id, "error", err)
		return nil, err
	}
	return b, nil
}

func (s *Service) List(ctx context.Context) ([]breakdown.Breakdown, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list breakdowns", "error", err)
		return nil, fmt.Errorf("list breakdowns: %w", err)
	}
	return items, nil
}

func (s *Service) ListByBike(ctx context.Context, bikeID int64) ([]breakdown.Breakdown, error) {
	if bikeID <= 0 {
		return nil, fmt.Errorf("%w: bike_id must be positive", breakdown.ErrInvalidInput)
	}

	items, err := s.repo.ListByBike(ctx, bikeID)
	if err != nil {
		s.logger.Error("failed to list breakdowns by bike", "bike_id", bikeID, "error", err)
		return nil, fmt.Errorf("list breakdowns by bike: %w", err)
	}
	return items, nil
}

func (s *Service) ListUnfixed(ctx context.Context) ([]breakdown.Breakdown, error) {
	items, err := s.repo.ListUnfixed(ctx)
	if err != nil {
		s.logger.Error("failed to list unfixed breakdowns", "error", err)
		return nil, fmt.Errorf("list unfixed breakdowns: %w", err)
	}
	return items, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*breakdown.Breakdown, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", breakdown.ErrInvalidInput)
	}

	// Загружаем текущее состояние для мержа.
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Применяем только пришедшие поля.
	if input.Reason != nil {
		reason := strings.TrimSpace(*input.Reason)
		if reason == "" {
			return nil, fmt.Errorf("%w: reason cannot be empty", breakdown.ErrInvalidInput)
		}
		existing.Reason = reason
	}
	if input.Cost != nil {
		if *input.Cost < 0 {
			return nil, fmt.Errorf("%w: cost cannot be negative", breakdown.ErrInvalidInput)
		}
		existing.Cost = *input.Cost
	}
	if input.Comment != nil {
		existing.Comment = input.Comment
	}

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		s.logger.Error("failed to update breakdown", "id", id, "error", err)
		return nil, fmt.Errorf("update breakdown: %w", err)
	}

	s.logger.Info("breakdown updated", "id", updated.ID)
	return updated, nil
}

func (s *Service) MarkFixed(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", breakdown.ErrInvalidInput)
	}

	err := s.repo.MarkFixed(ctx, id)
	if err != nil {
		s.logger.Error("failed to mark breakdown as fixed", "id", id, "error", err)
		return fmt.Errorf("mark fixed: %w", err)
	}

	s.logger.Info("breakdown marked as fixed", "id", id)
	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", breakdown.ErrInvalidInput)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete breakdown", "id", id, "error", err)
		return fmt.Errorf("delete breakdown: %w", err)
	}

	s.logger.Info("breakdown deleted", "id", id)
	return nil
}

// ============ validation ============

func validateCreateInput(input CreateInput) error {
	if input.BikeID <= 0 {
		return fmt.Errorf("%w: bike_id is required", breakdown.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", breakdown.ErrInvalidInput)
	}
	if input.Cost < 0 {
		return fmt.Errorf("%w: cost cannot be negative", breakdown.ErrInvalidInput)
	}
	return nil
}
