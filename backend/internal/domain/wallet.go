package domain

import (
	"errors"
	"time"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletAlreadyExists = errors.New("wallet already exists")
)

type Wallet struct {
	ID        int64
	UserID    int64
	Balance   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
