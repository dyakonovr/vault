package wallet_test

import (
	"context"
	"testing"
	"vault/internal/application/wallet"
	"vault/internal/application/wallet/mocks"
	"vault/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestGetById_Owner(t *testing.T) {
	repo := mocks.NewMockWalletReader(t)
	ucase := wallet.New(repo)

	w := domain.Wallet{ID: 1, UserID: 10, Balance: 500}
	repo.On("GetById", context.Background(), int64(1)).Return(w, nil)

	got, err := ucase.GetById(context.Background(), 1, 10)
	require.NoError(t, err)
	require.Equal(t, w, got)
}

func TestGetById_NotOwner(t *testing.T) {
	repo := mocks.NewMockWalletReader(t)
	ucase := wallet.New(repo)

	w := domain.Wallet{ID: 1, UserID: 10, Balance: 500}
	repo.On("GetById", context.Background(), int64(1)).Return(w, nil)

	_, err := ucase.GetById(context.Background(), 1, 99)
	require.ErrorIs(t, err, wallet.ErrWalletAccessDenied)
}

func TestGetByUserId(t *testing.T) {
	repo := mocks.NewMockWalletReader(t)
	ucase := wallet.New(repo)

	w := domain.Wallet{ID: 5, UserID: 10, Balance: 200}
	repo.On("GetByUserId", context.Background(), int64(10)).Return(w, nil)

	got, err := ucase.GetByUserId(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, w, got)
}

func TestCreate_Success(t *testing.T) {
	repo := mocks.NewMockWalletReader(t)
	ucase := wallet.New(repo)

	repo.On("Create", context.Background(), &domain.Wallet{UserID: 10}).Return(nil)

	got, err := ucase.Create(context.Background(), wallet.CreateWalletCommand{UserID: 10})
	require.NoError(t, err)
	require.Equal(t, int64(10), got.UserID)
	require.Equal(t, int64(0), got.Balance)
}

func TestCreate_InvalidUserID(t *testing.T) {
	repo := mocks.NewMockWalletReader(t)
	ucase := wallet.New(repo)

	_, err := ucase.Create(context.Background(), wallet.CreateWalletCommand{UserID: 0})
	require.ErrorIs(t, err, domain.ErrWalletIncorrectUserID)
}
