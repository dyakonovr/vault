package withdrawal

import (
	"context"
	"errors"
	walletapp "vault/internal/application/wallet"
	"vault/internal/domain"
)

type WithdrawUsecase struct {
	walletOwnershipChecker WalletOwnershipChecker
	unitOfWork             walletapp.UnitOfWork
}

func New(walletOwnershipChecker WalletOwnershipChecker, unitOfWork walletapp.UnitOfWork) *WithdrawUsecase {
	return &WithdrawUsecase{
		walletOwnershipChecker: walletOwnershipChecker,
		unitOfWork:             unitOfWork,
	}
}

func (u *WithdrawUsecase) Do(ctx context.Context, command WithdrawalCommand) (domain.Transaction, error) {
	var tx domain.Transaction

	if err := u.walletOwnershipChecker.IsOwnedBy(ctx, command.WalletID, command.UserID); err != nil {
		return tx, err
	}

	err := u.unitOfWork.StartTransaction(ctx, func(repos walletapp.TransactionalResources) error {
		existsTransaction, err := repos.TransactionRepository().GetByIdempotencyKey(ctx, command.IdempotencyKey)
		if err == nil { // Transaction already completed
			tx = existsTransaction
			return nil
		} else if !errors.Is(err, domain.ErrTransactionNotFound) {
			return err
		}

		wallet, err := repos.WalletRepository().GetByIdForUpdate(ctx, command.WalletID)
		if err != nil {
			return err
		}

		err = wallet.Withdraw(command.Amount)
		if err != nil {
			return err
		}

		err = repos.WalletRepository().Update(ctx, &wallet)
		if err != nil {
			return err
		}

		transaction, err := domain.NewTransaction(
			command.WalletID,
			domain.TransactionTypeWithdrawal,
			command.Amount,
			command.IdempotencyKey,
		)
		if err != nil {
			return err
		}

		err = repos.TransactionRepository().Create(ctx, transaction)
		if err != nil {
			return err
		}

		tx = *transaction
		return nil
	})

	return tx, err
}
