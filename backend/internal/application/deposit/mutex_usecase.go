package deposit

import (
	"context"
	"errors"
	"time"
	walletapp "vault/internal/application/wallet"
	"vault/internal/domain"
)

type MutexDepositUsecase struct {
	walletOwnershipChecker WalletOwnershipChecker
	unitOfWork             walletapp.UnitOfWork
	walletLocker           WalletLocker
}

func NewMutex(walletOwnershipChecker WalletOwnershipChecker, unitOfWork walletapp.UnitOfWork, walletLocker WalletLocker) *MutexDepositUsecase {
	return &MutexDepositUsecase{
		walletOwnershipChecker: walletOwnershipChecker,
		unitOfWork:             unitOfWork,
		walletLocker:           walletLocker,
	}
}

func (u *MutexDepositUsecase) Do(ctx context.Context, command DepositCommand) (domain.Transaction, error) {
	var tx domain.Transaction

	if err := u.walletOwnershipChecker.IsOwnedBy(ctx, command.WalletID, command.UserID); err != nil {
		return tx, err
	}

	unlock, err := u.walletLocker.Lock(ctx, command.WalletID, 5*time.Second)
	if err != nil {
		return tx, err
	}
	defer unlock()

	err = u.unitOfWork.StartTransaction(ctx, func(repos walletapp.TransactionalResources) error {
		existsTransaction, err := repos.TransactionRepository().GetByIdempotencyKey(ctx, command.IdempotencyKey)
		if err == nil { // Transaction already completed
			tx = existsTransaction
			return nil
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
			domain.TransactionTypeDeposit,
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
