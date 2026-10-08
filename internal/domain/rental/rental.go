package rental

import "time"

type Rental struct {
	ID          int64
	BikeID      int64
	RenterName  string
	RenterPhone string

	StartedAt    time.Time
	PlannedEndAt time.Time
	ReturnedAt   *time.Time

	TotalAmount int64
	IsPaid      bool

	Prepayment        *int64
	Deposit           *int64
	DepositReturned   *int64
	IsDepositReturned bool

	Comment   *string
	CreatedAt time.Time
}

func (r *Rental) IsActive() bool {
	return r.ReturnedAt == nil
}

func (r *Rental) IsOverdue(now time.Time) bool {
	return r.IsActive() && now.After(r.PlannedEndAt)
}
