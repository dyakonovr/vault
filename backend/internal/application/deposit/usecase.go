package deposit

import (
	"context"
	"errors"
	"vault/internal/domain"
)

type DepositUsecase struct {
	walletOwnershipChecker walletOwnershipChecker
	unitOfWork             unitOfWork
}

func New(walletOwnershipChecker walletOwnershipChecker, unitOfWork unitOfWork) *DepositUsecase {
	return &DepositUsecase{
		walletOwnershipChecker: walletOwnershipChecker,
		unitOfWork:             unitOfWork,
	}
}

func (u *DepositUsecase) Do(ctx context.Context, command DepositCommand) (domain.Transaction, error) {
	var tx domain.Transaction

	if err := u.walletOwnershipChecker.IsOwnedBy(ctx, command.WalletID, command.UserID); err != nil {
		return tx, err
	}

	err := u.unitOfWork.StartTransaction(ctx, func(repos TransactionalResources) error {
		_, err := repos.TransactionRepository().GetByIdempotencyKey(ctx, command.IdempotencyKey)
		if err == nil {
			return ErrDepositAlreadyCompleted
		} else if !errors.Is(err, domain.ErrTransactionNotFound) {
			return err
		}

		wallet, err := repos.WalletRepository().GetById(ctx, command.WalletID)
		if err != nil {
			return err
		}

		err = wallet.Deposit(command.Amount)
		if err != nil {
			return err
		}

		err = repos.WalletRepository().Update(ctx, &wallet)
		if err != nil {
			return err
		}

		transaction, err := domain.NewTransaction(
			command.WalletID,
			string(domain.TransactionTypeDeposit),
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
