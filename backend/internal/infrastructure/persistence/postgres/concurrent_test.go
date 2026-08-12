package postgres

import (
	"context"
	"sync"
	"testing"
	"vault/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestConcurrentDeposits(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "transactions")
	cleanTable(t, db, "wallets")
	cleanTable(t, db, "users")
	walletRepo := NewWalletRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "conc_user", PasswordHash: "hash"}
	userRepo.Create(ctx, user)

	w := &domain.Wallet{UserID: user.ID, Balance: 0}
	walletRepo.Create(ctx, w)

	var wg sync.WaitGroup
	numGoroutines := 10
	amountPerDeposit := int64(100)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx := db.Begin()
			defer tx.Rollback()

			var wallet domain.Wallet
			err := tx.Clauses().Raw("SELECT * FROM wallets WHERE id = ? FOR UPDATE", w.ID).Scan(&wallet).Error
			require.NoError(t, err)

			wallet.Balance += amountPerDeposit
			err = tx.Save(&WalletModel{
				ID:        wallet.ID,
				UserID:    wallet.UserID,
				Balance:   wallet.Balance,
				CreatedAt: wallet.CreatedAt,
				UpdatedAt: wallet.UpdatedAt,
			}).Error
			require.NoError(t, err)

			err = tx.Commit().Error
			require.NoError(t, err)
		}()
	}

	wg.Wait()

	found, err := walletRepo.GetById(ctx, w.ID)
	require.NoError(t, err)
	require.Equal(t, amountPerDeposit*int64(numGoroutines), found.Balance)
}
