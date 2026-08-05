package auth

import (
	"context"
	"errors"
	"time"
	"vault/internal/application/user"
	"vault/internal/domain"
	"vault/pkg/hash"
)

type AuthUsecase struct {
	userService  userService
	sessionStore sessionStore
	sessionTtl   time.Duration
}

func New(userService userService, sessionStore sessionStore, sessionTtl time.Duration) *AuthUsecase {
	return &AuthUsecase{
		userService:  userService,
		sessionStore: sessionStore,
		sessionTtl:   sessionTtl,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, command LoginCommand) (string, error) {
	user, err := u.userService.GetByLogin(ctx, command.Login)
	if err != nil {
		return "", err
	}

	isPasswordsEqual, err := hash.CompareArgon2(command.Password, user.PasswordHash)
	if err != nil {
		return "", err
	}

	if !isPasswordsEqual {
		return "", domain.ErrInvalidLoginOrPassword
	}

	sessionID, err := u.sessionStore.Create(ctx, user.ID, u.sessionTtl)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (u *AuthUsecase) Register(ctx context.Context, command RegisterCommand) error {
	_, err := u.userService.GetByLogin(ctx, command.Login)
	if err == nil {
		return domain.ErrUserAlreadyExists
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}

	_, err = u.userService.Create(ctx, user.CreateUserCommand(command))
	if err != nil {
		return err
	}

	return nil
}

func (u *AuthUsecase) Me(ctx context.Context, sessionID string) (domain.User, error) {
	userID, err := u.sessionStore.GetUserID(ctx, sessionID)
	if err != nil {
		return domain.User{}, err
	}

	return u.userService.GetById(ctx, userID)
}

func (u *AuthUsecase) Logout(ctx context.Context, sessionID string) error {
	u.sessionStore.Delete(ctx, sessionID)
	return nil
}