package handlers

import (
	"time"

	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/user"
	authusecase "github.com/IvanKuchsh-600/ebike-rental/internal/usecase/auth"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Name      *string   `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

func toUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
	}
}

func toLoginInput(req LoginRequest) authusecase.LoginInput {
	return authusecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}
}

func toChangePasswordInput(req ChangePasswordRequest) authusecase.ChangePasswordInput {
	return authusecase.ChangePasswordInput{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}
}
