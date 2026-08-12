package postgres

import (
	"context"
	"testing"
	"vault/internal/application"
	transactionapp "vault/internal/application/transaction"
	"vault/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateTransaction(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	txRepo := NewTransactionRepository(db)
	userRepo := NewUserRepository(db)
	walletRepo := NewWalletRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "tx_user1", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 1000}
	walletRepo.Create(ctx, w)

	key := uuid.New()
	tx := &domain.Transaction{
		WalletID:       w.ID,
		Type:           domain.TransactionTypeDeposit,
		Amount:         100,
		IdempotencyKey: key,
		Status:         domain.TransactionStatusCompleted,
	}

	err := txRepo.Create(ctx, tx)
	require.NoError(t, err)
	require.NotZero(t, tx.ID)
	require.False(t, tx.CreatedAt.IsZero())
}

func TestFindByIdempotencyKey(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	txRepo := NewTransactionRepository(db)
	userRepo := NewUserRepository(db)
	walletRepo := NewWalletRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "tx_user2", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 1000}
	walletRepo.Create(ctx, w)

	key := uuid.New()
	tx := &domain.Transaction{
		WalletID:       w.ID,
		Type:           domain.TransactionTypeDeposit,
		Amount:         50,
		IdempotencyKey: key,
		Status:         domain.TransactionStatusCompleted,
	}
	txRepo.Create(ctx, tx)

	found, err := txRepo.GetByIdempotencyKey(ctx, key)
	require.NoError(t, err)
	require.Equal(t, int64(50), found.Amount)
	require.Equal(t, domain.TransactionTypeDeposit, found.Type)
}

func TestCreateTransaction_DuplicateKey(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	txRepo := NewTransactionRepository(db)
	userRepo := NewUserRepository(db)
	walletRepo := NewWalletRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "tx_user3", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 1000}
	walletRepo.Create(ctx, w)

	key := uuid.New()
	tx1 := &domain.Transaction{
		WalletID:       w.ID,
		Type:           domain.TransactionTypeDeposit,
		Amount:         50,
		IdempotencyKey: key,
		Status:         domain.TransactionStatusCompleted,
	}
	err := txRepo.Create(ctx, tx1)
	require.NoError(t, err)

	tx2 := &domain.Transaction{
		WalletID:       w.ID,
		Type:           domain.TransactionTypeDeposit,
		Amount:         50,
		IdempotencyKey: key,
		Status:         domain.TransactionStatusCompleted,
	}
	err = txRepo.Create(ctx, tx2)
	require.ErrorIs(t, err, domain.ErrTransactionAlreadyExists)
}

func TestListByWalletID(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	txRepo := NewTransactionRepository(db)
	userRepo := NewUserRepository(db)
	walletRepo := NewWalletRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "tx_user4", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 1000}
	walletRepo.Create(ctx, w)

	txRepo.Create(ctx, &domain.Transaction{
		WalletID:       w.ID,
		Type:           domain.TransactionTypeDeposit,
		Amount:         100,
		IdempotencyKey: uuid.New(),
		Status:         domain.TransactionStatusCompleted,
	})
	txRepo.Create(ctx, &domain.Transaction{
		WalletID:       w.ID,
		Type:           domain.TransactionTypeWithdrawal,
		Amount:         50,
		IdempotencyKey: uuid.New(),
		Status:         domain.TransactionStatusCompleted,
	})

	txns, total, err := txRepo.List(ctx, transactionapp.ListTransactionsCommand{
		WalletID:          w.ID,
		PaginationCommand: application.PaginationCommand{Offset: 0, Limit: 10},
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, txns, 2)
}
