package bike

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
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

func (s *Service) Create(ctx context.Context, input CreateInput) (*bike.Bike, error) {
	if err := validateCreateInput(input); err != nil {
		s.logger.Warn("validation failed", "error", err)
		return nil, err
	}

	b := &bike.Bike{
		SerialNumber: strings.TrimSpace(input.SerialNumber),
		Model:        input.Model,
		Comment:      input.Comment,
	}

	created, err := s.repo.Create(ctx, b)
	if err != nil {
		s.logger.Error("failed to create bike", "error", err)
		return nil, fmt.Errorf("create bike: %w", err)
	}

	s.logger.Info("bike created", "id", created.ID, "serial", created.SerialNumber)
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*bike.Bike, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", bike.ErrInvalidInput)
	}

	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get bike", "id", id, "error", err)
		return nil, err
	}

	return b, nil
}

func (s *Service) List(ctx context.Context) ([]bike.Bike, error) {
	bikes, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list bikes", "error", err)
		return nil, fmt.Errorf("list bikes: %w", err)
	}

	return bikes, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*bike.Bike, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", bike.ErrInvalidInput)
	}

	// Загружаем существующий велосипед — нужно для частичного обновления.
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Применяем только те поля, что пришли.
	if input.SerialNumber != nil {
		if *input.SerialNumber == "" {
			return nil, fmt.Errorf("%w: serial_number cannot be empty", bike.ErrInvalidInput)
		}
		existing.SerialNumber = strings.TrimSpace(*input.SerialNumber)
	}
	if input.Model != nil {
		existing.Model = input.Model
	}
	if input.Comment != nil {
		existing.Comment = input.Comment
	}

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		s.logger.Error("failed to update bike", "id", id, "error", err)
		return nil, fmt.Errorf("update bike: %w", err)
	}

	s.logger.Info("bike updated", "id", updated.ID)
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", bike.ErrInvalidInput)
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, bike.ErrNotFound) {
			s.logger.Warn("bike not found for deletion", "id", id)
			return err
		}
		s.logger.Error("failed to delete bike", "id", id, "error", err)
		return fmt.Errorf("delete bike: %w", err)
	}

	s.logger.Info("bike deleted", "id", id)
	return nil
}

func (s *Service) SetRented(ctx context.Context, id int64, rented bool) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", bike.ErrInvalidInput)
	}

	err := s.repo.SetRented(ctx, id, rented)
	if err != nil {
		s.logger.Error("failed to set rented", "id", id, "value", rented, "error", err)
		return fmt.Errorf("set rented: %w", err)
	}

	s.logger.Info("bike rented status changed", "id", id, "is_rented", rented)
	return nil
}

func (s *Service) SetBroken(ctx context.Context, id int64, broken bool) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", bike.ErrInvalidInput)
	}
	if err := s.repo.SetBroken(ctx, id, broken); err != nil {
		s.logger.Error("failed to set broken", "id", id, "value", broken, "error", err)
		return fmt.Errorf("set broken: %w", err)
	}
	s.logger.Info("bike broken status changed", "id", id, "is_broken", broken)
	return nil
}

func validateCreateInput(input CreateInput) error {
	if strings.TrimSpace(input.SerialNumber) == "" {
		return fmt.Errorf("%w: serial_number is required", bike.ErrInvalidInput)
	}
	return nil
}
