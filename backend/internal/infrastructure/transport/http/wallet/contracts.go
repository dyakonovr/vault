package wallet

import (
	"context"
	depositapp "vault/internal/application/deposit"
	transferapp "vault/internal/application/transfer"
	walletapp "vault/internal/application/wallet"
	withdrawalapp "vault/internal/application/withdrawal"
	"vault/internal/domain"
)

//go:generate mockery
type WalletService interface {
	GetById(ctx context.Context, id, userID int64) (domain.Wallet, error)
	Create(ctx context.Context, command walletapp.CreateWalletCommand) (domain.Wallet, error)
}

//go:generate mockery
type DepositService interface {
	Do(ctx context.Context, command depositapp.DepositCommand) (domain.Transaction, error)
}

//go:generate mockery
type WithdrawalService interface {
	Do(ctx context.Context, command withdrawalapp.WithdrawalCommand) (domain.Transaction, error)
}

//go:generate mockery
type TransferService interface {
	Do(ctx context.Context, command transferapp.TransferCommand) (domain.Transaction, error)
}
