package handlers

import (
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/bike"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/breakdown"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/payout"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/rental"
	partnerusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/partner"
	payoutusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/payout"
)

// ============ Dashboard ============

type PartnerDashboardResponse struct {
	Bikes        PartnerBikesStats        `json:"bikes"`
	Rentals      PartnerRentalsStats      `json:"rentals"`
	CurrentMonth PartnerCurrentMonthStats `json:"current_month"`
}

type PartnerBikesStats struct {
	Total     int `json:"total"`
	Rented    int `json:"rented"`
	Broken    int `json:"broken"`
	Available int `json:"available"`
}

type PartnerRentalsStats struct {
	Active  int `json:"active"`
	Overdue int `json:"overdue"`
}

type PartnerCurrentMonthStats struct {
	Month             string    `json:"month"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	Revenue           int64     `json:"revenue"`
	PartnerShare      int64     `json:"partner_share"`
	AmortizationShare int64     `json:"amortization_share"`
}

func toPartnerDashboardResponse(r *partnerusecase.DashboardResult) PartnerDashboardResponse {
	return PartnerDashboardResponse{
		Bikes: PartnerBikesStats{
			Total:     r.Bikes.Total,
			Rented:    r.Bikes.Rented,
			Broken:    r.Bikes.Broken,
			Available: r.Bikes.Available,
		},
		Rentals: PartnerRentalsStats{
			Active:  r.Rentals.Active,
			Overdue: r.Rentals.Overdue,
		},
		CurrentMonth: PartnerCurrentMonthStats{
			Month:             r.CurrentMonth.Month,
			PeriodStart:       r.CurrentMonth.PeriodStart,
			PeriodEnd:         r.CurrentMonth.PeriodEnd,
			Revenue:           r.CurrentMonth.Revenue,
			PartnerShare:      r.CurrentMonth.PartnerShare,
			AmortizationShare: r.CurrentMonth.AmortizationShare,
		},
	}
}

// ============ Bike ============

// PartnerBikeResponse — БЕЗ comment.
type PartnerBikeResponse struct {
	ID           int64     `json:"id"`
	SerialNumber string    `json:"serial_number"`
	Model        *string   `json:"model"`
	IsRented     bool      `json:"is_rented"`
	IsBroken     bool      `json:"is_broken"`
	PhotoURL     *string   `json:"photo_url"`
	CreatedAt    time.Time `json:"created_at"`
}

func toPartnerBikeResponse(b *bike.Bike) PartnerBikeResponse {
	var photoURL *string
	if b.PhotoPath != nil {
		url := "/" + *b.PhotoPath
		if b.PhotoUpdatedAt != nil {
			url += "?t=" + b.PhotoUpdatedAt.Format("20060102150405")
		}
		photoURL = &url
	}

	return PartnerBikeResponse{
		ID:           b.ID,
		SerialNumber: b.SerialNumber,
		Model:        b.Model,
		IsRented:     b.IsRented,
		IsBroken:     b.IsBroken,
		PhotoURL:     photoURL,
		CreatedAt:    b.CreatedAt,
	}
}

func toPartnerBikeResponseList(items []bike.Bike) []PartnerBikeResponse {
	result := make([]PartnerBikeResponse, 0, len(items))
	for i := range items {
		result = append(result, toPartnerBikeResponse(&items[i]))
	}
	return result
}

// ============ Rental ============

// PartnerRentalResponse — serial_number вместо bike_id, БЕЗ comment.
type PartnerRentalResponse struct {
	ID          int64  `json:"id"`
	BikeSerial  string `json:"bike_serial"`
	RenterName  string `json:"renter_name"`
	RenterPhone string `json:"renter_phone"`

	StartedAt    time.Time  `json:"started_at"`
	PlannedEndAt time.Time  `json:"planned_end_at"`
	ReturnedAt   *time.Time `json:"returned_at"`

	TotalAmount int64 `json:"total_amount"`
	IsPaid      bool  `json:"is_paid"`

	CreatedAt time.Time `json:"created_at"`
}

// toPartnerRentalResponse — принимает rental + serial_number велосипеда.
// Serial нужно загрузить отдельно, потому что в rental.BikeID — только ID.
func toPartnerRentalResponse(r *rental.Rental, bikeSerial string) PartnerRentalResponse {
	return PartnerRentalResponse{
		ID:           r.ID,
		BikeSerial:   bikeSerial,
		RenterName:   r.RenterName,
		RenterPhone:  r.RenterPhone,
		StartedAt:    r.StartedAt,
		PlannedEndAt: r.PlannedEndAt,
		ReturnedAt:   r.ReturnedAt,
		TotalAmount:  r.TotalAmount,
		IsPaid:       r.IsPaid,
		CreatedAt:    r.CreatedAt,
	}
}

// ============ Breakdown ============

// PartnerBreakdownResponse — БЕЗ comment.
type PartnerBreakdownResponse struct {
	ID     int64  `json:"id"`
	BikeID int64  `json:"bike_id"`
	Reason string `json:"reason"`
	Cost   int64  `json:"cost"`

	BrokenAt time.Time  `json:"broken_at"`
	IsFixed  bool       `json:"is_fixed"`
	FixedAt  *time.Time `json:"fixed_at"`

	CreatedAt time.Time `json:"created_at"`
}

func toPartnerBreakdownResponse(b *breakdown.Breakdown) PartnerBreakdownResponse {
	return PartnerBreakdownResponse{
		ID:        b.ID,
		BikeID:    b.BikeID,
		Reason:    b.Reason,
		Cost:      b.Cost,
		BrokenAt:  b.BrokenAt,
		IsFixed:   b.IsFixed,
		FixedAt:   b.FixedAt,
		CreatedAt: b.CreatedAt,
	}
}

func toPartnerBreakdownResponseList(items []breakdown.Breakdown) []PartnerBreakdownResponse {
	result := make([]PartnerBreakdownResponse, 0, len(items))
	for i := range items {
		result = append(result, toPartnerBreakdownResponse(&items[i]))
	}
	return result
}

// ============ Payout ============

// PartnerPayoutResponse — БЕЗ comment.
type PartnerPayoutResponse struct {
	ID           int64      `json:"id"`
	PeriodStart  time.Time  `json:"period_start"`
	PeriodEnd    time.Time  `json:"period_end"`
	TotalRevenue int64      `json:"total_revenue"`
	PartnerShare int64      `json:"partner_share"`
	IsPaid       bool       `json:"is_paid"`
	PaidAt       *time.Time `json:"paid_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

func toPartnerPayoutResponse(p *payout.Payout) PartnerPayoutResponse {
	return PartnerPayoutResponse{
		ID:           p.ID,
		PeriodStart:  p.PeriodStart,
		PeriodEnd:    p.PeriodEnd,
		TotalRevenue: p.TotalRevenue,
		PartnerShare: p.PartnerShare,
		IsPaid:       p.IsPaid,
		PaidAt:       p.PaidAt,
		CreatedAt:    p.CreatedAt,
	}
}

func toPartnerPayoutResponseList(items []payout.Payout) []PartnerPayoutResponse {
	result := make([]PartnerPayoutResponse, 0, len(items))
	for i := range items {
		result = append(result, toPartnerPayoutResponse(&items[i]))
	}
	return result
}

// ============ Payout Preview ============

type PartnerPayoutPreviewResponse struct {
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	TotalRevenue int64     `json:"total_revenue"`
	PartnerShare int64     `json:"partner_share"`
}

func toPartnerPayoutPreviewResponse(r *payoutusecase.PreviewResult) PartnerPayoutPreviewResponse {
	return PartnerPayoutPreviewResponse{
		PeriodStart:  r.PeriodStart,
		PeriodEnd:    r.PeriodEnd,
		TotalRevenue: r.TotalRevenue,
		PartnerShare: r.PartnerShare,
	}
}
