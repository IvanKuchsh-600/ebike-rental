package handlers

import (
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
	bikeusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/bike"
)

// ============ Request DTOs ============

type CreateBikeRequest struct {
	SerialNumber string  `json:"serial_number" binding:"required"`
	Model        *string `json:"model"`
	Comment      *string `json:"comment"`
}

type UpdateBikeRequest struct {
	SerialNumber *string `json:"serial_number"`
	Model        *string `json:"model"`
	Comment      *string `json:"comment"`
}

// ============ Response DTOs ============

type BikeResponse struct {
	ID           int64     `json:"id"`
	SerialNumber string    `json:"serial_number"`
	Model        *string   `json:"model"`
	IsRented     bool      `json:"is_rented"`
	IsBroken     bool      `json:"is_broken"`
	Comment      *string   `json:"comment"`
	CreatedAt    time.Time `json:"created_at"`
}

// ============ Mappers ============

func toBikeResponse(b *bike.Bike) BikeResponse {
	return BikeResponse{
		ID:           b.ID,
		SerialNumber: b.SerialNumber,
		Model:        b.Model,
		IsRented:     b.IsRented,
		IsBroken:     b.IsBroken,
		Comment:      b.Comment,
		CreatedAt:    b.CreatedAt,
	}
}

func toBikeResponseList(bikes []bike.Bike) []BikeResponse {
	result := make([]BikeResponse, 0, len(bikes))
	for i := range bikes {
		result = append(result, toBikeResponse(&bikes[i]))
	}
	return result
}

func toCreateBikeInput(req CreateBikeRequest) bikeusecase.CreateInput {
	return bikeusecase.CreateInput{
		SerialNumber: req.SerialNumber,
		Model:        req.Model,
		Comment:      req.Comment,
	}
}

func toUpdateBikeInput(req UpdateBikeRequest) bikeusecase.UpdateInput {
	return bikeusecase.UpdateInput{
		SerialNumber: req.SerialNumber,
		Model:        req.Model,
		Comment:      req.Comment,
	}
}
