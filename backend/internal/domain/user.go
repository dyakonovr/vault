package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrWrongPassword          = errors.New("wrong password")
	ErrSessionNotFound        = errors.New("session not found")
	ErrUserEmptyLogin         = errors.New("user login can't be empty")
	ErrUserEmptyPasswordHash  = errors.New("user password can't be empty")
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(login, passwordHash string) (*User, error) {
	if login == "" {
		return nil, ErrUserEmptyLogin
	}

	if passwordHash == "" {
		return nil, ErrUserEmptyPasswordHash
	}

	return &User{
		Login:        login,
		PasswordHash: passwordHash,
	}, nil
}
