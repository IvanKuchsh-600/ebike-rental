package handlers

import (
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/rental"
	rentalusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/rental"
)

// ============ Request DTOs ============

type CreateRentalRequest struct {
	BikeID       int64      `json:"bike_id" binding:"required"`
	RenterName   string     `json:"renter_name" binding:"required"`
	RenterPhone  string     `json:"renter_phone" binding:"required"`
	StartedAt    *time.Time `json:"started_at"`
	PlannedEndAt time.Time  `json:"planned_end_at" binding:"required"`
	TotalAmount  int64      `json:"total_amount" binding:"required"`
	Prepayment   *int64     `json:"prepayment"`
	Deposit      *int64     `json:"deposit"`
	Comment      *string    `json:"comment"`
}

type UpdateRentalRequest struct {
	RenterName   *string    `json:"renter_name"`
	RenterPhone  *string    `json:"renter_phone"`
	PlannedEndAt *time.Time `json:"planned_end_at"`
	TotalAmount  *int64     `json:"total_amount"`
	Comment      *string    `json:"comment"`
}

// ============ Response DTOs ============

type RentalResponse struct {
	ID          int64  `json:"id"`
	BikeID      int64  `json:"bike_id"`
	RenterName  string `json:"renter_name"`
	RenterPhone string `json:"renter_phone"`

	StartedAt    time.Time  `json:"started_at"`
	PlannedEndAt time.Time  `json:"planned_end_at"`
	ReturnedAt   *time.Time `json:"returned_at"`

	TotalAmount int64 `json:"total_amount"`
	IsPaid      bool  `json:"is_paid"`

	Prepayment        *int64 `json:"prepayment"`
	Deposit           *int64 `json:"deposit"`
	DepositReturned   *int64 `json:"deposit_returned"`
	IsDepositReturned bool   `json:"is_deposit_returned"`

	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// ============ Mappers ============

func toRentalResponse(r *rental.Rental) RentalResponse {
	return RentalResponse{
		ID:                r.ID,
		BikeID:            r.BikeID,
		RenterName:        r.RenterName,
		RenterPhone:       r.RenterPhone,
		StartedAt:         r.StartedAt,
		PlannedEndAt:      r.PlannedEndAt,
		ReturnedAt:        r.ReturnedAt,
		TotalAmount:       r.TotalAmount,
		IsPaid:            r.IsPaid,
		Prepayment:        r.Prepayment,
		Deposit:           r.Deposit,
		DepositReturned:   r.DepositReturned,
		IsDepositReturned: r.IsDepositReturned,
		Comment:           r.Comment,
		CreatedAt:         r.CreatedAt,
	}
}

func toRentalResponseList(items []rental.Rental) []RentalResponse {
	result := make([]RentalResponse, 0, len(items))
	for i := range items {
		result = append(result, toRentalResponse(&items[i]))
	}
	return result
}

func toCreateRentalInput(req CreateRentalRequest) rentalusecase.CreateInput {
	return rentalusecase.CreateInput{
		BikeID:       req.BikeID,
		RenterName:   req.RenterName,
		RenterPhone:  req.RenterPhone,
		StartedAt:    req.StartedAt,
		PlannedEndAt: req.PlannedEndAt,
		TotalAmount:  req.TotalAmount,
		Prepayment:   req.Prepayment,
		Deposit:      req.Deposit,
		Comment:      req.Comment,
	}
}

func toUpdateRentalInput(req UpdateRentalRequest) rentalusecase.UpdateInput {
	return rentalusecase.UpdateInput{
		RenterName:   req.RenterName,
		RenterPhone:  req.RenterPhone,
		PlannedEndAt: req.PlannedEndAt,
		TotalAmount:  req.TotalAmount,
		Comment:      req.Comment,
	}
}
