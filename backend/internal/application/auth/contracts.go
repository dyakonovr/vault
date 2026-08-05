package auth

import (
	"context"
	"time"
	userapp "vault/internal/application/user"
	"vault/internal/domain"
)

type userService interface {
	GetByLogin(ctx context.Context, login string) (domain.User, error)
	GetById(ctx context.Context, id int64) (domain.User, error)
	Create(ctx context.Context, command userapp.CreateUserCommand) (domain.User, error)
}

type sessionStore interface {
	Create(ctx context.Context, userID int64, ttl time.Duration) (sessionID string, err error)
	GetUserID(ctx context.Context, sessionID string) (int64, error)
	Delete(ctx context.Context, sessionID string) error
}
