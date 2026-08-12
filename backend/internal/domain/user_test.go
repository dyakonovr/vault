package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	user, err := NewUser("alice", "hashed_password")
	require.NoError(t, err)
	require.Equal(t, "alice", user.Login)
	require.Equal(t, "hashed_password", user.PasswordHash)
	require.Zero(t, user.ID)
}

func TestNewUser_EmptyLogin(t *testing.T) {
	_, err := NewUser("", "hashed_password")
	require.ErrorIs(t, err, ErrUserEmptyLogin)
}

func TestNewUser_EmptyPassword(t *testing.T) {
	_, err := NewUser("alice", "")
	require.ErrorIs(t, err, ErrUserEmptyPasswordHash)
}
