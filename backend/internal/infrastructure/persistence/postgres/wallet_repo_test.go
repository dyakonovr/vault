package postgres

import (
	"context"
	"testing"
	"vault/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestCreateWallet(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user := &domain.User{Login: "wallet_user1", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 100}
	err := repo.Create(ctx, w)
	require.NoError(t, err)
	require.NotZero(t, w.ID)
	require.Equal(t, int64(100), w.Balance)
}

func TestFindByID_Wallet(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user := &domain.User{Login: "wallet_user2", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 200}
	repo.Create(ctx, w)

	found, err := repo.GetById(ctx, w.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, found.UserID)
	require.Equal(t, int64(200), found.Balance)
}

func TestFindByUserID(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user := &domain.User{Login: "wallet_user3", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 300}
	repo.Create(ctx, w)

	found, err := repo.GetByUserId(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, w.ID, found.ID)
	require.Equal(t, int64(300), found.Balance)
}

func TestUpdateBalance(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user := &domain.User{Login: "wallet_user4", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 500}
	repo.Create(ctx, w)

	w.Balance = 750
	err := repo.Update(ctx, w)
	require.NoError(t, err)

	found, err := repo.GetById(ctx, w.ID)
	require.NoError(t, err)
	require.Equal(t, int64(750), found.Balance)
}

func TestCreateWallet_DuplicateUserID(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user := &domain.User{Login: "wallet_user5", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w1 := &domain.Wallet{UserID: user.ID, Balance: 100}
	err := repo.Create(ctx, w1)
	require.NoError(t, err)

	w2 := &domain.Wallet{UserID: user.ID, Balance: 200}
	err = repo.Create(ctx, w2)
	require.ErrorIs(t, err, domain.ErrWalletAlreadyExists)
}
