package breakdown

import "time"

type Breakdown struct {
	ID        int64
	BikeID    int64
	Reason    string
	Cost      int64 // рубли (без копеек)
	BrokenAt  time.Time
	IsFixed   bool
	FixedAt   *time.Time // nullable
	Comment   *string    // nullable
	CreatedAt time.Time
}
