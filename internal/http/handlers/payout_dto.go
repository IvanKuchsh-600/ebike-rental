package handlers

import (
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/payout"
	payoutusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/payout"
)

// ============ Request DTOs ============

type CreatePayoutRequest struct {
	Month   string  `json:"month" binding:"required"`
	Comment *string `json:"comment"`
}

// ============ Response DTOs ============

type PayoutResponse struct {
	ID           int64      `json:"id"`
	PeriodStart  time.Time  `json:"period_start"`
	PeriodEnd    time.Time  `json:"period_end"`
	TotalRevenue int64      `json:"total_revenue"`
	PartnerShare int64      `json:"partner_share"`
	IsPaid       bool       `json:"is_paid"`
	PaidAt       *time.Time `json:"paid_at"`
	Comment      *string    `json:"comment"`
	CreatedAt    time.Time  `json:"created_at"`
}

type PayoutPreviewResponse struct {
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	TotalRevenue int64     `json:"total_revenue"`
	PartnerShare int64     `json:"partner_share"`
}

// ============ Mappers ============

func toPayoutResponse(p *payout.Payout) PayoutResponse {
	return PayoutResponse{
		ID:           p.ID,
		PeriodStart:  p.PeriodStart,
		PeriodEnd:    p.PeriodEnd,
		TotalRevenue: p.TotalRevenue,
		PartnerShare: p.PartnerShare,
		IsPaid:       p.IsPaid,
		PaidAt:       p.PaidAt,
		Comment:      p.Comment,
		CreatedAt:    p.CreatedAt,
	}
}

func toPayoutResponseList(items []payout.Payout) []PayoutResponse {
	result := make([]PayoutResponse, 0, len(items))
	for i := range items {
		result = append(result, toPayoutResponse(&items[i]))
	}
	return result
}

func toPayoutPreviewResponse(r *payoutusecase.PreviewResult) PayoutPreviewResponse {
	return PayoutPreviewResponse{
		PeriodStart:  r.PeriodStart,
		PeriodEnd:    r.PeriodEnd,
		TotalRevenue: r.TotalRevenue,
		PartnerShare: r.PartnerShare,
	}
}

func toCreatePayoutInput(req CreatePayoutRequest) payoutusecase.CreateInput {
	return payoutusecase.CreateInput{
		Month:   req.Month,
		Comment: req.Comment,
	}
}
