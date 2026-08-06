package wallet

import (
	"context"
	"vault/internal/domain"
)

type WalletUsecase struct {
	walletRepository walletRepository
}

func New(walletRepository walletRepository) *WalletUsecase {
	return &WalletUsecase{
		walletRepository: walletRepository,
	}
}

func (u *WalletUsecase) GetById(ctx context.Context, id int64) (domain.Wallet, error) {
	return u.walletRepository.GetById(ctx, id)
}

func (u *WalletUsecase) GetByUserId(ctx context.Context, userID int64) (domain.Wallet, error) {
	return u.walletRepository.GetByUserId(ctx, userID)
}

func (u *WalletUsecase) Create(ctx context.Context, command CreateWalletCommand) (domain.Wallet, error) {
	wallet, err := domain.NewWallet(command.UserID)
	if err != nil {
		return domain.Wallet{}, err
	}

	if err = u.walletRepository.Create(ctx, wallet); err != nil {
		return domain.Wallet{}, err
	}
	return *wallet, nil
}

func (u *WalletUsecase) Deposit(ctx context.Context, id int64, command WalletDepositCommand) (domain.Wallet, error) {
	wallet, err := u.walletRepository.GetById(ctx, id)
	if err != nil {
		return domain.Wallet{}, err
	}

	if wallet.UserID != command.UserID {
		return domain.Wallet{}, ErrWalletAccessDenied
	}

	err = wallet.Deposit(command.Amount)
	if err != nil {
		return domain.Wallet{}, err
	}

	if err := u.walletRepository.Update(ctx, &wallet); err != nil {
		return domain.Wallet{}, err
	}
	return wallet, nil
}

func (u *WalletUsecase) Withdraw(ctx context.Context, id int64, command WalletWithdrawCommand) (domain.Wallet, error) {
	wallet, err := u.walletRepository.GetById(ctx, id)
	if err != nil {
		return domain.Wallet{}, err
	}

	if wallet.UserID != command.UserID {
		return domain.Wallet{}, ErrWalletAccessDenied
	}

	err = wallet.Withdraw(command.Amount)
	if err != nil {
		return domain.Wallet{}, err
	}

	if err := u.walletRepository.Update(ctx, &wallet); err != nil {
		return domain.Wallet{}, err
	}
	return wallet, nil
}

func (u *WalletUsecase) IsOwnedBy(ctx context.Context, walletID, userID int64) (bool, error) {
	wallet, err := u.walletRepository.GetById(ctx, walletID)
	if err != nil {
		return false, err
	}

	if wallet.UserID != userID {
		return false, nil
	}

	return true, nil
}