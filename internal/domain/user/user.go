package user

import "time"

const (
	RoleOwner   = "owner"
	RolePartner = "partner"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	Name         *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) IsOwner() bool {
	return u.Role == RoleOwner
}

func (u *User) IsPartner() bool {
	return u.Role == RolePartner
}

func IsValidRole(role string) bool {
	return role == RoleOwner || role == RolePartner
}
