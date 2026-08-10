package wallet

import (
	"context"
	depositapp "vault/internal/application/deposit"
	walletapp "vault/internal/application/wallet"
	withdrawalapp "vault/internal/application/withdrawal"
	"vault/internal/domain"
)

type walletService interface {
	GetById(ctx context.Context, id, userID int64) (domain.Wallet, error)
	Create(ctx context.Context, command walletapp.CreateWalletCommand) (domain.Wallet, error)
}

type depositService interface {
	Do(ctx context.Context, command depositapp.DepositCommand) (domain.Transaction, error)
}

type withdrawalService interface {
	Do(ctx context.Context, command withdrawalapp.WithdrawalCommand) (domain.Transaction, error)
}
