package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWallet(t *testing.T) {
	wallet, err := NewWallet(1)
	require.NoError(t, err)
	require.Equal(t, int64(1), wallet.UserID)
	require.Equal(t, int64(0), wallet.Balance)
}

func TestNewWallet_InvalidUserID(t *testing.T) {
	_, err := NewWallet(0)
	require.ErrorIs(t, err, ErrWalletIncorrectUserID)

	_, err = NewWallet(-1)
	require.ErrorIs(t, err, ErrWalletIncorrectUserID)
}

func TestDeposit_Success(t *testing.T) {
	wallet, _ := NewWallet(1)
	err := wallet.Deposit(100)
	require.NoError(t, err)
	require.Equal(t, int64(100), wallet.Balance)
	require.False(t, wallet.UpdatedAt.IsZero())
}

func TestDeposit_ZeroAmount(t *testing.T) {
	wallet, _ := NewWallet(1)
	err := wallet.Deposit(0)
	require.ErrorIs(t, err, ErrWalletInvalidAmount)
	require.Equal(t, int64(0), wallet.Balance)
}

func TestDeposit_NegativeAmount(t *testing.T) {
	wallet, _ := NewWallet(1)
	err := wallet.Deposit(-50)
	require.ErrorIs(t, err, ErrWalletInvalidAmount)
	require.Equal(t, int64(0), wallet.Balance)
}

func TestWithdraw_Success(t *testing.T) {
	wallet, _ := NewWallet(1)
	wallet.Deposit(200)
	err := wallet.Withdraw(50)
	require.NoError(t, err)
	require.Equal(t, int64(150), wallet.Balance)
	require.False(t, wallet.UpdatedAt.IsZero())
}

func TestWithdraw_ZeroAmount(t *testing.T) {
	wallet, _ := NewWallet(1)
	wallet.Deposit(100)
	err := wallet.Withdraw(0)
	require.ErrorIs(t, err, ErrWalletInvalidAmount)
	require.Equal(t, int64(100), wallet.Balance)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	wallet, _ := NewWallet(1)
	wallet.Deposit(50)
	err := wallet.Withdraw(100)
	require.ErrorIs(t, err, ErrInsufficientFunds)
	require.Equal(t, int64(50), wallet.Balance)
}

func TestWithdraw_EntireBalance(t *testing.T) {
	wallet, _ := NewWallet(1)
	wallet.Deposit(100)
	err := wallet.Withdraw(100)
	require.NoError(t, err)
	require.Equal(t, int64(0), wallet.Balance)
}
