package auth

import (
	"time"
	"vault/internal/domain"
)

type UserResponse struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserResponse(u domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Login:     u.Login,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
