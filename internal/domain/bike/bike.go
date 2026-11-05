package bike

import "time"

const (
	StatusAvailable = "available" // Свободен
	StatusRented    = "rented"    // Сдан
	StatusBroken    = "broken"    // Сломан
	StatusAssembly  = "assembly"  // Сборка
)

type Bike struct {
	ID             int64
	SerialNumber   string
	Model          *string
	Status         string // <-- НОВОЕ ПОЛЕ
	Comment        *string
	PhotoPath      *string
	PhotoUpdatedAt *time.Time
	CreatedAt      time.Time
}

// Хелперы для удобства
func (b *Bike) IsAvailable() bool { return b.Status == StatusAvailable }
func (b *Bike) IsRented() bool    { return b.Status == StatusRented }
