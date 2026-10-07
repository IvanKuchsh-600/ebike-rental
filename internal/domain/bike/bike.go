package bike

import "time"

type Bike struct {
	ID           int64
	SerialNumber string
	Model        *string
	IsRented     bool
	IsBroken     bool
	Comment      *string
	CreatedAt    time.Time
}
