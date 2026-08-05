package user

import (
	"context"
	"vault/internal/domain"
)

type userRepository interface {
	List(ctx context.Context, params ListUsersCommand) ([]domain.User, int64, error)
	GetById(ctx context.Context, id int64) (domain.User, error)
	GetByLogin(ctx context.Context, login string) (domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
}
