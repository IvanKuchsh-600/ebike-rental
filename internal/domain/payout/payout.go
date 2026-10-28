package payout

import "time"

type Payout struct {
	ID                int64
	PeriodStart       time.Time
	PeriodEnd         time.Time
	TotalRevenue      int64 // выручка за период (в рублях)
	OwnerShare        int64 // доля владельца (40%)
	PartnerShare      int64 // доля партнёра (40%)
	AmortizationShare int64 // амортизация (20%)
	IsPaid            bool
	PaidAt            *time.Time
	Comment           *string
	CreatedAt         time.Time
}
