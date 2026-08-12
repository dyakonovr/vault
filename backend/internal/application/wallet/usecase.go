package wallet

import (
	"context"
	"vault/internal/domain"
)

type WalletUsecase struct {
	walletRepository WalletReader
}

func New(walletRepository WalletReader) *WalletUsecase {
	return &WalletUsecase{
		walletRepository: walletRepository,
	}
}

func (u *WalletUsecase) GetById(ctx context.Context, id, userID int64) (domain.Wallet, error) {
	if err := u.IsOwnedBy(ctx, id, userID); err != nil {
		return domain.Wallet{}, err
	}

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

func (u *WalletUsecase) IsOwnedBy(ctx context.Context, walletID, userID int64) error {
	wallet, err := u.walletRepository.GetById(ctx, walletID)
	if err != nil {
		return err
	}

	if wallet.UserID != userID {
		return ErrWalletAccessDenied
	}

	return nil
}
