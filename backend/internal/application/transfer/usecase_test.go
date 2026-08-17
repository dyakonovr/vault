package transfer_test

import (
	"context"
	"testing"
	"vault/internal/application/transfer"
	transfermocks "vault/internal/application/transfer/mocks"
	walletapp "vault/internal/application/wallet"
	walletmocks "vault/internal/application/wallet/mocks"
	"vault/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDo_Success(t *testing.T) {
	ownershipChecker := transfermocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	walletRepo := walletmocks.NewMockWalletRepository(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := transfer.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	cmd := transfer.TransferCommand{UserID: 1, WalletFromID: 10, WalletToID: 20, IdempotencyKey: key, Amount: 10}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		err := caller(resources)
		return err == nil
	})).Return(nil)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(domain.Transaction{}, domain.ErrTransactionNotFound)
	resources.On("TransactionRepository").Return(txRepo)

	wFrom := domain.Wallet{ID: 10, UserID: 1, Balance: 100}
	wTo := domain.Wallet{ID: 20, UserID: 2, Balance: 50}
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(10)).Return(wFrom, nil)
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(20)).Return(wTo, nil)
	resources.On("WalletRepository").Return(walletRepo)

	walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
		return w.ID == 10 && w.UserID == 1 && w.Balance == 90
	})).Return(nil)
	walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
		return w.ID == 20 && w.UserID == 2 && w.Balance == 60
	})).Return(nil)
	txRepo.On("Create", context.Background(), mock.AnythingOfType("*domain.Transaction")).Return(nil).Maybe()

	got, err := ucase.Do(context.Background(), cmd)
	require.NoError(t, err)
	require.Equal(t, int64(10), got.WalletID)
	require.Equal(t, domain.TransactionTypeTransferOut, got.Type)
	require.Equal(t, int64(10), got.Amount)
	require.Equal(t, key, got.IdempotencyKey)
}

func TestDo_Idempotent(t *testing.T) {
	ownershipChecker := transfermocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := transfer.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	existingTx := domain.Transaction{ID: 42, WalletID: 10, Type: domain.TransactionTypeTransferOut, Amount: 10, IdempotencyKey: key}

	cmd := transfer.TransferCommand{UserID: 1, WalletFromID: 10, WalletToID: 20, IdempotencyKey: key, Amount: 10}

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
	ownershipChecker := transfermocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)

	ucase := transfer.New(ownershipChecker, unitOfWork)

	cmd := transfer.TransferCommand{UserID: 1, WalletFromID: 10, WalletToID: 20, IdempotencyKey: uuid.New(), Amount: 10}
	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(domain.ErrWalletNotFound)

	_, err := ucase.Do(context.Background(), cmd)
	require.ErrorIs(t, err, domain.ErrWalletNotFound)
}

func TestDo_InsufficientFunds(t *testing.T) {
	ownershipChecker := transfermocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	walletRepo := walletmocks.NewMockWalletRepository(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := transfer.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	cmd := transfer.TransferCommand{UserID: 1, WalletFromID: 10, WalletToID: 20, IdempotencyKey: key, Amount: 200}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		_ = caller(resources)
		return true
	})).Return(domain.ErrInsufficientFunds)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(domain.Transaction{}, domain.ErrTransactionNotFound)
	resources.On("TransactionRepository").Return(txRepo)

	wFrom := domain.Wallet{ID: 10, UserID: 1, Balance: 100}
	wTo := domain.Wallet{ID: 20, UserID: 2, Balance: 50}
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(10)).Return(wFrom, nil)
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(20)).Return(wTo, nil)
	resources.On("WalletRepository").Return(walletRepo)

	_, err := ucase.Do(context.Background(), cmd)
	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestDo_WalletNotFound(t *testing.T) {
	ownershipChecker := transfermocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	walletRepo := walletmocks.NewMockWalletRepository(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := transfer.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	cmd := transfer.TransferCommand{UserID: 1, WalletFromID: 10, WalletToID: 20, IdempotencyKey: key, Amount: 10}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(10), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		_ = caller(resources)
		return true
	})).Return(domain.ErrWalletNotFound)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(domain.Transaction{}, domain.ErrTransactionNotFound)
	resources.On("TransactionRepository").Return(txRepo)
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(20)).Return(domain.Wallet{}, domain.ErrWalletNotFound)
	resources.On("WalletRepository").Return(walletRepo)

	_, err := ucase.Do(context.Background(), cmd)
	require.ErrorIs(t, err, domain.ErrWalletNotFound)
}

func TestDo_LockOrdering(t *testing.T) {
	ownershipChecker := transfermocks.NewMockWalletOwnershipChecker(t)
	unitOfWork := walletmocks.NewMockUnitOfWork(t)
	resources := walletmocks.NewMockTransactionalResources(t)
	walletRepo := walletmocks.NewMockWalletRepository(t)
	txRepo := walletmocks.NewMockTransactionRepository(t)

	ucase := transfer.New(ownershipChecker, unitOfWork)

	key := uuid.New()
	cmd := transfer.TransferCommand{UserID: 1, WalletFromID: 20, WalletToID: 10, IdempotencyKey: key, Amount: 10}

	ownershipChecker.On("IsOwnedBy", context.Background(), int64(20), int64(1)).Return(nil)
	unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(func(caller func(walletapp.TransactionalResources) error) bool {
		err := caller(resources)
		return err == nil
	})).Return(nil)

	txRepo.On("GetByIdempotencyKey", context.Background(), key).Return(domain.Transaction{}, domain.ErrTransactionNotFound)
	resources.On("TransactionRepository").Return(txRepo)

	wFrom := domain.Wallet{ID: 20, UserID: 1, Balance: 100}
	wTo := domain.Wallet{ID: 10, UserID: 2, Balance: 50}
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(20)).Return(wFrom, nil)
	walletRepo.On("GetByIdForUpdate", context.Background(), int64(10)).Return(wTo, nil)
	resources.On("WalletRepository").Return(walletRepo)

	walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
		return w.ID == 20 && w.Balance == 90
	})).Return(nil)
	walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
		return w.ID == 10 && w.Balance == 60
	})).Return(nil)
	txRepo.On("Create", context.Background(), mock.AnythingOfType("*domain.Transaction")).Return(nil).Maybe()

	got, err := ucase.Do(context.Background(), cmd)
	require.NoError(t, err)
	require.Equal(t, int64(20), got.WalletID)
	require.Equal(t, domain.TransactionTypeTransferOut, got.Type)
}
