package domain

import (
	"errors"
	"time"
)

var (
	ErrWalletNotFound        = errors.New("wallet not found")
	ErrWalletAlreadyExists   = errors.New("wallet already exists")
	ErrWalletInvalidAmount   = errors.New("amount must be greater than zero")
	ErrInsufficientFunds     = errors.New("insufficient funds on the balance")
	ErrWalletIncorrectUserID = errors.New("wallet userID should be non-zero value")
)

type Wallet struct {
	ID        int64
	UserID    int64
	Balance   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewWallet(userID int64) (*Wallet, error) {
	if userID <= 0 {
		return nil, ErrWalletIncorrectUserID
	}

	return &Wallet{
		UserID:  userID,
		Balance: 0,
	}, nil
}

// Начисление денег на баланс
func (w *Wallet) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrWalletInvalidAmount
	}

	w.Balance += amount
	w.UpdatedAt = time.Now()
	return nil
}

// Списание денег с баланса
func (w *Wallet) Withdrawal(amount int64) error {
	if amount <= 0 {
		return ErrWalletInvalidAmount
	}

	newBalance := w.Balance - amount
	if newBalance < 0 {
		return ErrInsufficientFunds
	}

	w.Balance = newBalance
	w.UpdatedAt = time.Now()
	return nil
}
