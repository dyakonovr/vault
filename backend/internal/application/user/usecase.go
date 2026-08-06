package user

import (
	"context"
	"time"
	"vault/internal/domain"
	"vault/pkg/hash"
)

type UserUsecase struct {
	userRepo userRepository
}

func New(userRepo userRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (u *UserUsecase) List(ctx context.Context, command ListUsersCommand) ([]domain.User, int64, error) {
	return u.userRepo.List(ctx, command)
}

func (u *UserUsecase) GetById(ctx context.Context, id int64) (domain.User, error) {
	return u.userRepo.GetById(ctx, id)
}

func (u *UserUsecase) GetByLogin(ctx context.Context, login string) (domain.User, error) {
	return u.userRepo.GetByLogin(ctx, login)
}

func (u *UserUsecase) Create(ctx context.Context, command CreateUserCommand) (domain.User, error) {
	passwordHash, err := hash.HashArgon2(command.Password)
	if err != nil {
		return domain.User{}, err
	}

	user, err := domain.NewUser(command.Login, passwordHash)
	if err != nil {
		return domain.User{}, err
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return domain.User{}, err
	}
	return *user, nil
}

func (u *UserUsecase) Update(ctx context.Context, id int64, command UpdateUserCommand) (domain.User, error) {
	user, err := u.userRepo.GetById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	passwordEquals, err := hash.CompareArgon2(command.OldPassword, user.PasswordHash)
	if err != nil {
		return domain.User{}, err
	}
	if !passwordEquals {
		return domain.User{}, domain.ErrWrongPassword
	}

	newPasswordHash, err := hash.HashArgon2(command.NewPassword)
	if err != nil {
		return domain.User{}, err
	}

	user.PasswordHash = newPasswordHash
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, &user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (u *UserUsecase) Delete(ctx context.Context, id int64) error {
	return u.userRepo.Delete(ctx, id)
}
