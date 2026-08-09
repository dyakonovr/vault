package wallet

import (
	"context"
	depositapp "vault/internal/application/deposit"
	walletapp "vault/internal/application/wallet"
	"vault/internal/domain"
)

type walletService interface {
	GetById(ctx context.Context, id, userID int64) (domain.Wallet, error)
	Create(ctx context.Context, command walletapp.CreateWalletCommand) (domain.Wallet, error)
	Withdraw(ctx context.Context, id int64, command walletapp.WalletWithdrawCommand) (domain.Wallet, error)
}

type depositService interface {
	Do(ctx context.Context, id int64, command depositapp.DepositCommand) (domain.Transaction, error)
}
