package transaction_test

import (
	"context"
	"testing"
	"vault/internal/application/transaction"
	"vault/internal/application/transaction/mocks"
	"vault/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestList_Success(t *testing.T) {
	txRepo := mocks.NewMockTransactionRepository(t)
	ownershipChecker := mocks.NewMockWalletOwnershipChecker(t)
	ucase := transaction.New(txRepo, ownershipChecker)

	cmd := transaction.ListTransactionsCommand{WalletID: 10, UserID: 1}
	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)

	txns := []domain.Transaction{
		{ID: 1, WalletID: 10, Type: domain.TransactionTypeDeposit, Amount: 100},
		{ID: 2, WalletID: 10, Type: domain.TransactionTypeWithdrawal, Amount: 50},
	}
	txRepo.On("List", context.Background(), cmd).Return(txns, int64(2), nil)

	got, total, err := ucase.List(context.Background(), cmd)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, got, 2)
}

func TestList_OwnershipError(t *testing.T) {
	txRepo := mocks.NewMockTransactionRepository(t)
	ownershipChecker := mocks.NewMockWalletOwnershipChecker(t)
	ucase := transaction.New(txRepo, ownershipChecker)

	cmd := transaction.ListTransactionsCommand{WalletID: 10, UserID: 1}
	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(domain.ErrWalletNotFound)

	_, _, err := ucase.List(context.Background(), cmd)
	require.ErrorIs(t, err, domain.ErrWalletNotFound)
}

func TestGetById_Success(t *testing.T) {
	txRepo := mocks.NewMockTransactionRepository(t)
	ownershipChecker := mocks.NewMockWalletOwnershipChecker(t)
	ucase := transaction.New(txRepo, ownershipChecker)

	cmd := transaction.GetTransactionByIdCommand{WalletID: 10, UserID: 1}
	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)

	tx := domain.Transaction{ID: 5, WalletID: 10, Type: domain.TransactionTypeDeposit, Amount: 100}
	txRepo.On("GetById", context.Background(), int64(5)).Return(tx, nil)

	got, err := ucase.GetById(context.Background(), 5, cmd)
	require.NoError(t, err)
	require.Equal(t, tx, got)
}

func TestCreate_Success(t *testing.T) {
	txRepo := mocks.NewMockTransactionRepository(t)
	ownershipChecker := mocks.NewMockWalletOwnershipChecker(t)
	ucase := transaction.New(txRepo, ownershipChecker)

	key := uuid.New()
	cmd := transaction.CreateTransactionCommand{WalletID: 10, UserID: 1, Type: "deposit", Amount: 100, IdempotencyKey: key}
	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	txRepo.On("Create", context.Background(), mock.AnythingOfType("*domain.Transaction")).Return(nil)

	got, err := ucase.Create(context.Background(), cmd)
	require.NoError(t, err)
	require.Equal(t, domain.TransactionTypeDeposit, got.Type)
	require.Equal(t, int64(100), got.Amount)
	require.Equal(t, key, got.IdempotencyKey)
}
