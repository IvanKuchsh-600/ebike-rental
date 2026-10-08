package payout

import "time"

type Payout struct {
	ID           int64
	PeriodStart  time.Time
	PeriodEnd    time.Time
	TotalRevenue int64 // выручка за период (в рублях)
	PartnerShare int64 // доля партнёра (10%)
	IsPaid       bool
	PaidAt       *time.Time // nullable
	Comment      *string    // nullable
	CreatedAt    time.Time
}
