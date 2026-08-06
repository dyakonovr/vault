package wallet

import (
	"context"
	walletapp "vault/internal/application/wallet"
	"vault/internal/domain"
)

type walletService interface {
	GetById(ctx context.Context, id int64) (domain.Wallet, error)
	Create(ctx context.Context, command walletapp.CreateWalletCommand) (domain.Wallet, error)
	Deposit(ctx context.Context, id int64, command walletapp.WalletDepositCommand) (domain.Wallet, error)
	Withdraw(ctx context.Context, id int64, command walletapp.WalletWithdrawCommand) (domain.Wallet, error)
}
