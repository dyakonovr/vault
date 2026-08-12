package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewTransaction_Deposit(t *testing.T) {
	key := uuid.New()
	tx, err := NewTransaction(1, TransactionTypeDeposit, 100, key)
	require.NoError(t, err)
	require.Equal(t, int64(1), tx.WalletID)
	require.Equal(t, TransactionTypeDeposit, tx.Type)
	require.Equal(t, int64(100), tx.Amount)
	require.Equal(t, key, tx.IdempotencyKey)
	require.Equal(t, TransactionStatusCompleted, tx.Status)
}

func TestNewTransaction_TransferOut(t *testing.T) {
	key := uuid.New()
	tx, err := NewTransaction(1, TransactionTypeTransferOut, 50, key)
	require.NoError(t, err)
	require.Equal(t, TransactionTypeTransferOut, tx.Type)
	require.Equal(t, int64(50), tx.Amount)
}

func TestNewTransaction_TransferIn(t *testing.T) {
	key := uuid.New()
	tx, err := NewTransaction(2, TransactionTypeTransferIn, 50, key)
	require.NoError(t, err)
	require.Equal(t, TransactionTypeTransferIn, tx.Type)
	require.Equal(t, int64(50), tx.Amount)
}

func TestNewTransaction_Withdrawal(t *testing.T) {
	key := uuid.New()
	tx, err := NewTransaction(1, TransactionTypeWithdrawal, 30, key)
	require.NoError(t, err)
	require.Equal(t, TransactionTypeWithdrawal, tx.Type)
	require.Equal(t, int64(30), tx.Amount)
}

func TestNewTransaction_InvalidType(t *testing.T) {
	key := uuid.New()
	_, err := NewTransaction(1, "invalid_type", 100, key)
	require.ErrorIs(t, err, ErrTransactionIncorrectType)
}

func TestNewTransaction_ZeroAmount(t *testing.T) {
	key := uuid.New()
	_, err := NewTransaction(1, TransactionTypeDeposit, 0, key)
	require.ErrorIs(t, err, ErrTransactionInvalidAmount)
}

func TestNewTransaction_NegativeAmount(t *testing.T) {
	key := uuid.New()
	_, err := NewTransaction(1, TransactionTypeDeposit, -100, key)
	require.ErrorIs(t, err, ErrTransactionInvalidAmount)
}

func TestNewTransaction_InvalidWalletID(t *testing.T) {
	key := uuid.New()
	_, err := NewTransaction(0, TransactionTypeDeposit, 100, key)
	require.ErrorIs(t, err, ErrTransactionIncorrectWalletID)

	_, err = NewTransaction(-1, TransactionTypeDeposit, 100, key)
	require.ErrorIs(t, err, ErrTransactionIncorrectWalletID)
}
