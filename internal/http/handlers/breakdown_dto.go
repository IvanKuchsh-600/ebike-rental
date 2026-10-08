package handlers

import (
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/breakdown"
	breakdownusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/breakdown"
)

// ============ Request DTOs ============

type CreateBreakdownRequest struct {
	BikeID   int64      `json:"bike_id" binding:"required"`
	Reason   string     `json:"reason" binding:"required"`
	Cost     int64      `json:"cost"`
	BrokenAt *time.Time `json:"broken_at"`
	Comment  *string    `json:"comment"`
}

type UpdateBreakdownRequest struct {
	Reason  *string `json:"reason"`
	Cost    *int64  `json:"cost"`
	Comment *string `json:"comment"`
}

// ============ Response DTOs ============

type BreakdownResponse struct {
	ID        int64      `json:"id"`
	BikeID    int64      `json:"bike_id"`
	Reason    string     `json:"reason"`
	Cost      int64      `json:"cost"`
	BrokenAt  time.Time  `json:"broken_at"`
	IsFixed   bool       `json:"is_fixed"`
	FixedAt   *time.Time `json:"fixed_at"`
	Comment   *string    `json:"comment"`
	CreatedAt time.Time  `json:"created_at"`
}

// ============ Mappers ============

func toBreakdownResponse(b *breakdown.Breakdown) BreakdownResponse {
	return BreakdownResponse{
		ID:        b.ID,
		BikeID:    b.BikeID,
		Reason:    b.Reason,
		Cost:      b.Cost,
		BrokenAt:  b.BrokenAt,
		IsFixed:   b.IsFixed,
		FixedAt:   b.FixedAt,
		Comment:   b.Comment,
		CreatedAt: b.CreatedAt,
	}
}

func toBreakdownResponseList(items []breakdown.Breakdown) []BreakdownResponse {
	result := make([]BreakdownResponse, 0, len(items))
	for i := range items {
		result = append(result, toBreakdownResponse(&items[i]))
	}
	return result
}

func toCreateBreakdownInput(req CreateBreakdownRequest) breakdownusecase.CreateInput {
	return breakdownusecase.CreateInput{
		BikeID:   req.BikeID,
		Reason:   req.Reason,
		Cost:     req.Cost,
		BrokenAt: req.BrokenAt,
		Comment:  req.Comment,
	}
}

func toUpdateBreakdownInput(req UpdateBreakdownRequest) breakdownusecase.UpdateInput {
	return breakdownusecase.UpdateInput{
		Reason:  req.Reason,
		Cost:    req.Cost,
		Comment: req.Comment,
	}
}
