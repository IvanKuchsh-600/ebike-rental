package rental

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/rental"
)

type Service struct {
	rentalRepo Repository
	bikeRepo   BikeRepository
	transactor Transactor
	logger     *slog.Logger
}

func NewService(
	rentalRepo Repository,
	bikeRepo BikeRepository,
	transactor Transactor,
	logger *slog.Logger,
) *Service {
	return &Service{
		rentalRepo: rentalRepo,
		bikeRepo:   bikeRepo,
		transactor: transactor,
		logger:     logger,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*rental.Rental, error) {
	if err := validateCreateInput(input); err != nil {
		s.logger.Warn("validation failed", "error", err)
		return nil, err
	}

	var created *rental.Rental

	err := s.transactor.WithinTransaction(ctx, func(ctx context.Context, repos Repos) error {
		// 1. Заблокировать и загрузить велосипед
		b, err := repos.Bikes.GetByIDForUpdate(ctx, input.BikeID)
		if err != nil {
			if errors.Is(err, bike.ErrNotFound) {
				return rental.ErrBikeNotFound
			}
			return fmt.Errorf("get bike for update: %w", err)
		}

		// 2. Проверить доступность
		if !b.IsAvailable() {
			return rental.ErrBikeNotAvailable
		}

		// 3. Сформировать аренду
		startedAt := time.Now().UTC()
		if input.StartedAt != nil {
			startedAt = *input.StartedAt
		}

		r := &rental.Rental{
			BikeID:       input.BikeID,
			RenterName:   strings.TrimSpace(input.RenterName),
			RenterPhone:  strings.TrimSpace(input.RenterPhone),
			StartedAt:    startedAt,
			PlannedEndAt: input.PlannedEndAt,
			TotalAmount:  input.TotalAmount,
			IsPaid:       false,
			Prepayment:   input.Prepayment,
			Deposit:      input.Deposit,
			Comment:      input.Comment,
		}

		// 4. Создать аренду
		result, err := repos.Rentals.Create(ctx, r)
		if err != nil {
			return fmt.Errorf("create rental: %w", err)
		}
		created = result

		// 5. Пометить велосипед сданным
		if err := repos.Bikes.UpdateStatus(ctx, input.BikeID, bike.StatusRented); err != nil {
			return fmt.Errorf("set bike rented: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to create rental", "bike_id", input.BikeID, "error", err)
		return nil, err
	}

	s.logger.Info("rental created", "id", created.ID, "bike_id", created.BikeID)
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*rental.Rental, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", rental.ErrInvalidInput)
	}

	r, err := s.rentalRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get rental", "id", id, "error", err)
		return nil, err
	}
	return r, nil
}

func (s *Service) List(ctx context.Context) ([]rental.Rental, error) {
	items, err := s.rentalRepo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list rentals", "error", err)
		return nil, fmt.Errorf("list rentals: %w", err)
	}
	return items, nil
}

func (s *Service) ListActive(ctx context.Context) ([]rental.Rental, error) {
	items, err := s.rentalRepo.ListActive(ctx)
	if err != nil {
		s.logger.Error("failed to list active rentals", "error", err)
		return nil, fmt.Errorf("list active rentals: %w", err)
	}
	return items, nil
}

func (s *Service) ListByBike(ctx context.Context, bikeID int64) ([]rental.Rental, error) {
	if bikeID <= 0 {
		return nil, fmt.Errorf("%w: bike_id must be positive", rental.ErrInvalidInput)
	}

	items, err := s.rentalRepo.ListByBike(ctx, bikeID)
	if err != nil {
		s.logger.Error("failed to list rentals by bike", "bike_id", bikeID, "error", err)
		return nil, fmt.Errorf("list rentals by bike: %w", err)
	}
	return items, nil
}

// ============ Update ============

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*rental.Rental, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", rental.ErrInvalidInput)
	}

	existing, err := s.rentalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Нельзя редактировать возвращённую аренду
	if !existing.IsActive() {
		return nil, fmt.Errorf("%w: cannot update returned rental", rental.ErrInvalidInput)
	}

	if input.RenterName != nil {
		name := strings.TrimSpace(*input.RenterName)
		if name == "" {
			return nil, fmt.Errorf("%w: renter_name cannot be empty", rental.ErrInvalidInput)
		}
		existing.RenterName = name
	}
	if input.RenterPhone != nil {
		phone := strings.TrimSpace(*input.RenterPhone)
		if phone == "" {
			return nil, fmt.Errorf("%w: renter_phone cannot be empty", rental.ErrInvalidInput)
		}
		existing.RenterPhone = phone
	}
	if input.PlannedEndAt != nil {
		if !input.PlannedEndAt.After(existing.StartedAt) {
			return nil, fmt.Errorf("%w: planned_end_at must be after started_at", rental.ErrInvalidInput)
		}
		existing.PlannedEndAt = *input.PlannedEndAt
	}
	if input.TotalAmount != nil {
		if *input.TotalAmount <= 0 {
			return nil, fmt.Errorf("%w: total_amount must be positive", rental.ErrInvalidInput)
		}
		existing.TotalAmount = *input.TotalAmount
	}
	if input.Prepayment != nil {
		if *input.Prepayment < 0 {
			return nil, fmt.Errorf("%w: prepayment cannot be negative", rental.ErrInvalidInput)
		}
		existing.Prepayment = input.Prepayment
	}
	if input.Deposit != nil {
		if *input.Deposit < 0 {
			return nil, fmt.Errorf("%w: deposit cannot be negative", rental.ErrInvalidInput)
		}
		existing.Deposit = input.Deposit
	}
	if input.Comment != nil {
		existing.Comment = input.Comment
	}

	updated, err := s.rentalRepo.Update(ctx, existing)
	if err != nil {
		s.logger.Error("failed to update rental", "id", id, "error", err)
		return nil, fmt.Errorf("update rental: %w", err)
	}

	s.logger.Info("rental updated", "id", updated.ID)
	return updated, nil
}

// ============ Return ============

// Return — вернуть велосипед: транзакция (rental.MarkReturned + bike.SetRented(false)).
func (s *Service) Return(ctx context.Context, id int64) (*rental.Rental, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", rental.ErrInvalidInput)
	}

	var result *rental.Rental

	err := s.transactor.WithinTransaction(ctx, func(ctx context.Context, repos Repos) error {
		// 1. Загрузить аренду
		r, err := repos.Rentals.GetByID(ctx, id)
		if err != nil {
			return err
		}

		// 2. Проверить, что аренда активна
		if !r.IsActive() {
			return rental.ErrAlreadyReturned
		}

		// 3. Пометить возвращённой
		if err := repos.Rentals.MarkReturned(ctx, id); err != nil {
			return fmt.Errorf("mark returned: %w", err)
		}

		// 4. Освободить велосипед
		if err := repos.Bikes.UpdateStatus(ctx, r.BikeID, bike.StatusAvailable); err != nil {
			return fmt.Errorf("set bike free: %w", err)
		}

		// 5. Вернуть актуальное состояние
		updated, err := repos.Rentals.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("get updated rental: %w", err)
		}
		result = updated

		return nil
	})

	if err != nil {
		s.logger.Error("failed to return rental", "id", id, "error", err)
		return nil, err
	}

	s.logger.Info("rental returned", "id", result.ID, "bike_id", result.BikeID)
	return result, nil
}

// ============ MarkPaid ============

func (s *Service) MarkPaid(ctx context.Context, id int64) (*rental.Rental, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", rental.ErrInvalidInput)
	}

	if err := s.rentalRepo.MarkPaid(ctx, id); err != nil {
		s.logger.Error("failed to mark paid", "id", id, "error", err)
		return nil, fmt.Errorf("mark paid: %w", err)
	}

	updated, err := s.rentalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get updated rental: %w", err)
	}

	s.logger.Info("rental marked as paid", "id", id)
	return updated, nil
}

// ============ Delete ============

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", rental.ErrInvalidInput)
	}

	// Загружаем аренду, чтобы проверить правила
	r, err := s.rentalRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Нельзя удалить возвращённую аренду
	if r.ReturnedAt != nil {
		return rental.ErrCannotDeleteReturned
	}

	// Нельзя удалить оплаченную аренду
	if r.IsPaid {
		return rental.ErrCannotDeletePaid
	}

	if err := s.rentalRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete rental", "id", id, "error", err)
		return fmt.Errorf("delete rental: %w", err)
	}

	s.logger.Info("rental deleted", "id", id)
	return nil
}

func validateCreateInput(input CreateInput) error {
	if input.BikeID <= 0 {
		return fmt.Errorf("%w: bike_id is required", rental.ErrInvalidInput)
	}
	if strings.TrimSpace(input.RenterName) == "" {
		return fmt.Errorf("%w: renter_name is required", rental.ErrInvalidInput)
	}
	if strings.TrimSpace(input.RenterPhone) == "" {
		return fmt.Errorf("%w: renter_phone is required", rental.ErrInvalidInput)
	}
	if input.PlannedEndAt.IsZero() {
		return fmt.Errorf("%w: planned_end_at is required", rental.ErrInvalidInput)
	}
	if input.TotalAmount <= 0 {
		return fmt.Errorf("%w: total_amount must be positive", rental.ErrInvalidInput)
	}

	// Проверяем, что даты согласованы
	startedAt := time.Now().UTC()
	if input.StartedAt != nil {
		startedAt = *input.StartedAt
	}
	if !input.PlannedEndAt.After(startedAt) {
		return fmt.Errorf("%w: planned_end_at must be after started_at", rental.ErrInvalidInput)
	}

	return nil
}

// ChangePaymentStatus — меняет статус оплаты аренды.
func (s *Service) ChangePaymentStatus(ctx context.Context, id int64, isPaid bool) (*rental.Rental, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", rental.ErrInvalidInput)
	}

	if err := s.rentalRepo.SetPaid(ctx, id, isPaid); err != nil {
		s.logger.Error("failed to change payment status", "id", id, "error", err)
		return nil, err
	}

	updated, err := s.rentalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get updated rental: %w", err)
	}

	s.logger.Info("rental payment status changed", "id", id, "is_paid", isPaid)
	return updated, nil
}

// ReturnDeposit — возвращает залог по аренде.
func (s *Service) ReturnDeposit(ctx context.Context, id int64, returned bool) (*rental.Rental, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", rental.ErrInvalidInput)
	}

	if err := s.rentalRepo.SetDepositReturned(ctx, id, returned); err != nil {
		s.logger.Error("failed to return deposit", "id", id, "error", err)
		return nil, err
	}

	updated, err := s.rentalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get updated rental: %w", err)
	}

	s.logger.Info("rental deposit returned", "id", id, "returned", returned)
	return updated, nil
}

// CancelReturn — отменяет возврат аренды (делает её снова активной).
func (s *Service) CancelReturn(ctx context.Context, id int64) (*rental.Rental, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", rental.ErrInvalidInput)
	}

	var result *rental.Rental

	err := s.transactor.WithinTransaction(ctx, func(ctx context.Context, repos Repos) error {
		// 1. Загрузить аренду
		r, err := repos.Rentals.GetByID(ctx, id)
		if err != nil {
			return err
		}

		// 2. Проверить, что аренда действительно была возвращена
		if r.ReturnedAt == nil {
			return fmt.Errorf("%w: rental is already active", rental.ErrInvalidInput)
		}

		// 3. Отменить возврат (обнулить returned_at)
		if err := repos.Rentals.CancelReturn(ctx, id); err != nil {
			return fmt.Errorf("cancel return: %w", err)
		}

		// 4. Занять велосипед обратно (статус "Сдан")
		if err := repos.Bikes.UpdateStatus(ctx, r.BikeID, bike.StatusRented); err != nil {
			return fmt.Errorf("set bike rented: %w", err)
		}

		// 5. Вернуть актуальное состояние
		updated, err := repos.Rentals.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("get updated rental: %w", err)
		}
		result = updated

		return nil
	})

	if err != nil {
		s.logger.Error("failed to cancel rental return", "id", id, "error", err)
		return nil, err
	}

	s.logger.Info("rental return cancelled", "id", id)
	return result, nil
}
