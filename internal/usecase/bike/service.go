package bike

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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

// UploadPhoto — сохраняет фото велосипеда.
// file — содержимое файла (байты).
// filename — оригинальное имя (для определения типа).
func (s *Service) UploadPhoto(ctx context.Context, id int64, file []byte, filename string) (*bike.Bike, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", bike.ErrInvalidInput)
	}

	// Проверка размера — макс 5 MB
	const maxSize = 5 * 1024 * 1024
	if len(file) > maxSize {
		return nil, fmt.Errorf("%w: file too large (max 5MB)", bike.ErrInvalidInput)
	}

	// Проверка типа файла по расширению
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		return nil, fmt.Errorf("%w: unsupported file type (allowed: jpg, jpeg, png, webp)", bike.ErrInvalidInput)
	}

	// Проверяем, что велосипед существует
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Формируем путь: uploads/bikes/{id}{ext}
	photoPath := fmt.Sprintf("uploads/bikes/%d%s", id, ext)

	// Сохраняем файл на диск
	if err := os.MkdirAll("uploads/bikes", 0755); err != nil {
		s.logger.Error("failed to create uploads dir", "error", err)
		return nil, fmt.Errorf("create uploads dir: %w", err)
	}

	fullPath := filepath.Join(".", photoPath)
	if err := os.WriteFile(fullPath, file, 0644); err != nil {
		s.logger.Error("failed to save photo", "id", id, "error", err)
		return nil, fmt.Errorf("save photo: %w", err)
	}

	// Если было старое фото с другим расширением — удаляем
	if existing.PhotoPath != nil && *existing.PhotoPath != photoPath {
		oldPath := filepath.Join(".", *existing.PhotoPath)
		_ = os.Remove(oldPath) // игнорируем ошибку — файла может не быть
	}

	// Обновляем путь в БД
	if err := s.repo.UpdatePhoto(ctx, id, photoPath); err != nil {
		s.logger.Error("failed to update photo path in db", "id", id, "error", err)
		return nil, fmt.Errorf("update photo in db: %w", err)
	}

	s.logger.Info("bike photo uploaded", "id", id, "path", photoPath)

	// Возвращаем обновлённый велосипед
	return s.repo.GetByID(ctx, id)
}

// DeletePhoto — удаляет фото велосипеда.
func (s *Service) DeletePhoto(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", bike.ErrInvalidInput)
	}

	// Загружаем велосипед, чтобы узнать путь
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.PhotoPath == nil {
		// Фото нет — не ошибка
		return nil
	}

	// Удаляем файл с диска
	oldPath := filepath.Join(".", *existing.PhotoPath)
	if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("failed to remove photo file", "id", id, "path", oldPath, "error", err)
	}

	// Обновляем БД
	if err := s.repo.ClearPhoto(ctx, id); err != nil {
		s.logger.Error("failed to clear photo in db", "id", id, "error", err)
		return fmt.Errorf("clear photo in db: %w", err)
	}

	s.logger.Info("bike photo deleted", "id", id)
	return nil
}
