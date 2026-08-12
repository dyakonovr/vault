package wallet

import (
	"context"
	"vault/internal/domain"
)

//go:generate mockery
type WalletReader interface {
	GetById(ctx context.Context, id int64) (domain.Wallet, error)
	GetByUserId(ctx context.Context, userID int64) (domain.Wallet, error)
	Create(ctx context.Context, wallet *domain.Wallet) error
}
