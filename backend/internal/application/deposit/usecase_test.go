package deposit_test

import (
	"context"
	"testing"
	"vault/internal/application/deposit"
	depositmocks "vault/internal/application/deposit/mocks"
	walletapp "vault/internal/application/wallet"
	walletmocks "vault/internal/application/wallet/mocks"
	"vault/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDo_Success(t *testing.T) {
	ownershipChecker := depositmocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	walletRepo := walletmocks.NewMockWalletRepository(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := deposit.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	cmd := deposit.DepositCommand{UserID: 1, WalletID: 10, IdempotencyKey: key, Amount: 100}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		err := caller(resources)
		return err == nil
	})).Return(nil)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(domain.Transaction{}, domain.ErrTransactionNotFound)
	resources.On("TransactionRepository").Return(txRepo)

	w := domain.Wallet{ID: 10, UserID: 1, Balance: 0}
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(10)).Return(w, nil)
	walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
		return w.ID == 10 && w.UserID == 1 && w.Balance == 100
	})).Return(nil)
	resources.On("WalletRepository").Return(walletRepo)

	txRepo.On("Create", context.Background(), mock.AnythingOfType("*domain.Transaction")).Return(nil)

	got, err := ucase.Do(context.Background(), cmd)
	require.NoError(t, err)
	require.Equal(t, domain.TransactionTypeDeposit, got.Type)
	require.Equal(t, int64(100), got.Amount)
	require.Equal(t, key, got.IdempotencyKey)
}

func TestDo_Idempotent(t *testing.T) {
	ownershipChecker := depositmocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := deposit.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	existingTx := domain.Transaction{ID: 42, WalletID: 10, Type: domain.TransactionTypeDeposit, Amount: 100, IdempotencyKey: key}

	cmd := deposit.DepositCommand{UserID: 1, WalletID: 10, IdempotencyKey: key, Amount: 100}

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

func TestDo_OwnershipError(t *testing.T) {
	ownershipChecker := depositmocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)

	ucase := deposit.New(ownershipChecker, unitOfWork)

	cmd := deposit.DepositCommand{UserID: 1, WalletID: 10, IdempotencyKey: uuid.New(), Amount: 100}
	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(domain.ErrWalletNotFound)

	_, err := ucase.Do(context.Background(), cmd)
	require.ErrorIs(t, err, domain.ErrWalletNotFound)
}

func TestDo_WalletNotFound(t *testing.T) {
	ownershipChecker := depositmocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	walletRepo := walletmocks.NewMockWalletRepository(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := deposit.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	cmd := deposit.DepositCommand{UserID: 1, WalletID: 10, IdempotencyKey: key, Amount: 100}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		_ = caller(resources)
		return true
	})).Return(domain.ErrWalletNotFound)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(domain.Transaction{}, domain.ErrTransactionNotFound)
	resources.On("TransactionRepository").Return(txRepo)
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(10)).Return(domain.Wallet{}, domain.ErrWalletNotFound)
	resources.On("WalletRepository").Return(walletRepo)

	_, err := ucase.Do(context.Background(), cmd)
	require.ErrorIs(t, err, domain.ErrWalletNotFound)
}
