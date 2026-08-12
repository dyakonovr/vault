package auth_test

import (
	"context"
	"testing"
	"time"
	"vault/internal/application/auth"
	"vault/internal/application/auth/mocks"
	"vault/internal/application/user"
	"vault/internal/domain"
	"vault/pkg/hash"

	"github.com/stretchr/testify/require"
)

func TestLogin_Success(t *testing.T) {
	userService := mocks.NewMockUserService(t)
	sessionStore := mocks.NewMockSessionStore(t)
	ucase := auth.New(userService, sessionStore, 30*time.Minute)

	passwordHash, _ := hash.HashArgon2("secret123")
	u := domain.User{ID: 1, Login: "alice", PasswordHash: passwordHash}
	userService.On("GetByLogin", context.Background(), "alice").Return(u, nil)
	sessionStore.On("Create", context.Background(), int64(1), 30*time.Minute).Return("session-abc", nil)

	got, err := ucase.Login(context.Background(), auth.LoginCommand{Login: "alice", Password: "secret123"})
	require.NoError(t, err)
	require.Equal(t, "session-abc", got.Value)
	require.Equal(t, 30*time.Minute, got.Ttl)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	userService := mocks.NewMockUserService(t)
	sessionStore := mocks.NewMockSessionStore(t)
	ucase := auth.New(userService, sessionStore, 30*time.Minute)

	userService.On("GetByLogin", context.Background(), "alice").Return(domain.User{}, domain.ErrUserNotFound)

	_, err := ucase.Login(context.Background(), auth.LoginCommand{Login: "alice", Password: "secret123"})
	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestRegister_Success(t *testing.T) {
	userService := mocks.NewMockUserService(t)
	sessionStore := mocks.NewMockSessionStore(t)
	ucase := auth.New(userService, sessionStore, 30*time.Minute)

	userService.On("GetByLogin", context.Background(), "alice").Return(domain.User{}, domain.ErrUserNotFound)
	userService.On("Create", context.Background(), user.CreateUserCommand{Login: "alice", Password: "secret123"}).Return(domain.User{ID: 1, Login: "alice"}, nil)

	err := ucase.Register(context.Background(), auth.RegisterCommand{Login: "alice", Password: "secret123"})
	require.NoError(t, err)
}

func TestRegister_AlreadyExists(t *testing.T) {
	userService := mocks.NewMockUserService(t)
	sessionStore := mocks.NewMockSessionStore(t)
	ucase := auth.New(userService, sessionStore, 30*time.Minute)

	existing := domain.User{ID: 1, Login: "alice"}
	userService.On("GetByLogin", context.Background(), "alice").Return(existing, nil)

	err := ucase.Register(context.Background(), auth.RegisterCommand{Login: "alice", Password: "secret123"})
	require.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestMe_Success(t *testing.T) {
	userService := mocks.NewMockUserService(t)
	sessionStore := mocks.NewMockSessionStore(t)
	ucase := auth.New(userService, sessionStore, 30*time.Minute)

	sessionStore.On("GetUserID", context.Background(), "session-abc").Return(int64(1), nil)
	u := domain.User{ID: 1, Login: "alice"}
	userService.On("GetById", context.Background(), int64(1)).Return(u, nil)

	got, err := ucase.Me(context.Background(), "session-abc")
	require.NoError(t, err)
	require.Equal(t, u, got)
}

func TestMe_InvalidSession(t *testing.T) {
	userService := mocks.NewMockUserService(t)
	sessionStore := mocks.NewMockSessionStore(t)
	ucase := auth.New(userService, sessionStore, 30*time.Minute)

	sessionStore.On("GetUserID", context.Background(), "invalid-session").Return(int64(0), domain.ErrSessionNotFound)

	_, err := ucase.Me(context.Background(), "invalid-session")
	require.ErrorIs(t, err, domain.ErrSessionNotFound)
}

func TestLogout_Success(t *testing.T) {
	userService := mocks.NewMockUserService(t)
	sessionStore := mocks.NewMockSessionStore(t)
	ucase := auth.New(userService, sessionStore, 30*time.Minute)

	sessionStore.On("Delete", context.Background(), "session-abc").Return(nil)

	err := ucase.Logout(context.Background(), "session-abc")
	require.NoError(t, err)
}
