package user_test

import (
	"context"
	"testing"
	"vault/internal/application/user"
	"vault/internal/application/user/mocks"
	"vault/internal/domain"
	"vault/pkg/hash"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate_Success(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	ucase := user.New(repo)

	repo.On("Create", context.Background(), mock.AnythingOfType("*domain.User")).Return(nil)

	got, err := ucase.Create(context.Background(), user.CreateUserCommand{Login: "alice", Password: "secret123"})
	require.NoError(t, err)
	require.Equal(t, "alice", got.Login)
	require.NotEmpty(t, got.PasswordHash)
	require.NotEqual(t, "secret123", got.PasswordHash)
}

func TestGetById_Success(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	ucase := user.New(repo)

	u := domain.User{ID: 1, Login: "alice", PasswordHash: "hash"}
	repo.On("GetById", context.Background(), int64(1)).Return(u, nil)

	got, err := ucase.GetById(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, u, got)
}

func TestGetById_NotFound(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	ucase := user.New(repo)

	repo.On("GetById", context.Background(), int64(99)).Return(domain.User{}, domain.ErrUserNotFound)

	_, err := ucase.GetById(context.Background(), 99)
	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUpdate_Success(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	ucase := user.New(repo)

	passwordHash, err := hash.HashArgon2("oldpassword")
	require.NoError(t, err)

	u := domain.User{ID: 1, Login: "alice", PasswordHash: passwordHash}
	repo.On("GetById", context.Background(), int64(1)).Return(u, nil)
	repo.On("Update", context.Background(), mock.AnythingOfType("*domain.User")).Return(nil)

	got, err := ucase.Update(context.Background(), 1, user.UpdateUserCommand{OldPassword: "oldpassword", NewPassword: "newpassword"})
	require.NoError(t, err)
	require.Equal(t, "alice", got.Login)
	require.NotEqual(t, passwordHash, got.PasswordHash)
}

func TestUpdate_WrongPassword(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	ucase := user.New(repo)

	passwordHash, err := hash.HashArgon2("correctpassword")
	require.NoError(t, err)

	u := domain.User{ID: 1, Login: "alice", PasswordHash: passwordHash}
	repo.On("GetById", context.Background(), int64(1)).Return(u, nil)

	_, err = ucase.Update(context.Background(), 1, user.UpdateUserCommand{OldPassword: "wrongpassword", NewPassword: "newpassword"})
	require.ErrorIs(t, err, domain.ErrWrongPassword)
}
