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
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(login, passwordHash string) (*User, error) {
	return &User{
		Login:        login,
		PasswordHash: passwordHash,
	}, nil
}
