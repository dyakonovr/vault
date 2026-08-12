package withdrawal_test

import (
	"context"
	"testing"
	walletapp "vault/internal/application/wallet"
	walletmocks "vault/internal/application/wallet/mocks"
	"vault/internal/application/withdrawal"
	withdrawalmocks "vault/internal/application/withdrawal/mocks"
	"vault/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDo_Success(t *testing.T) {
	ownershipChecker := withdrawalmocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	walletRepo := walletmocks.NewMockWalletRepository(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := withdrawal.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	cmd := withdrawal.WithdrawalCommand{UserID: 1, WalletID: 10, IdempotencyKey: key, Amount: 50}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		err := caller(resources)
		return err == nil
	})).Return(nil)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(domain.Transaction{}, domain.ErrTransactionNotFound)
	resources.On("TransactionRepository").Return(txRepo)

	w := domain.Wallet{ID: 10, UserID: 1, Balance: 200}
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(10)).Return(w, nil)
	walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
		return w.ID == 10 && w.UserID == 1 && w.Balance == 150
	})).Return(nil)
	resources.On("WalletRepository").Return(walletRepo)

	txRepo.On("Create", context.Background(), mock.AnythingOfType("*domain.Transaction")).Return(nil)

	got, err := ucase.Do(context.Background(), cmd)
	require.NoError(t, err)
	require.Equal(t, domain.TransactionTypeWithdrawal, got.Type)
	require.Equal(t, int64(50), got.Amount)
	require.Equal(t, key, got.IdempotencyKey)
}

func TestDo_Idempotent(t *testing.T) {
	ownershipChecker := withdrawalmocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := withdrawal.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	existingTx := domain.Transaction{ID: 42, WalletID: 10, Type: domain.TransactionTypeWithdrawal, Amount: 50, IdempotencyKey: key}

	cmd := withdrawal.WithdrawalCommand{UserID: 1, WalletID: 10, IdempotencyKey: key, Amount: 50}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		err := caller(resources)
		return err == nil
	})).Return(nil)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(existingTx, nil)
	resources.On("TransactionRepository").Return(txRepo)

	got, err := ucase.Do(context.Background(), cmd)
	require.NoError(t, err)
	require.Equal(t, existingTx, got)
}

func TestDo_InsufficientFunds(t *testing.T) {
	ownershipChecker := withdrawalmocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	walletRepo := walletmocks.NewMockWalletRepository(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := withdrawal.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	cmd := withdrawal.WithdrawalCommand{UserID: 1, WalletID: 10, IdempotencyKey: key, Amount: 1000}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		_ = caller(resources)
		return true
	})).Return(domain.ErrInsufficientFunds)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(domain.Transaction{}, domain.ErrTransactionNotFound)
	resources.On("TransactionRepository").Return(txRepo)

	w := domain.Wallet{ID: 10, UserID: 1, Balance: 50}
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(10)).Return(w, nil)
	resources.On("WalletRepository").Return(walletRepo)

	_, err := ucase.Do(context.Background(), cmd)
	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestDo_OwnershipError(t *testing.T) {
	ownershipChecker := withdrawalmocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)

	ucase := withdrawal.New(ownershipChecker, unitOfWork)

	cmd := withdrawal.WithdrawalCommand{UserID: 1, WalletID: 10, IdempotencyKey: uuid.New(), Amount: 50}
	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(domain.ErrWalletNotFound)

	_, err := ucase.Do(context.Background(), cmd)
	require.ErrorIs(t, err, domain.ErrWalletNotFound)
}
