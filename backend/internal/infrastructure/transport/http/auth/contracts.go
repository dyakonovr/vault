package auth

import (
	"context"
	authapp "vault/internal/application/auth"
	"vault/internal/domain"
)

type authService interface {
	Login(ctx context.Context, command authapp.LoginCommand) (authapp.Session, error)
	Register(ctx context.Context, command authapp.RegisterCommand) error
	Me(ctx context.Context, sessionID string) (domain.User, error)
	Logout(ctx context.Context, sessionID string) error
}
