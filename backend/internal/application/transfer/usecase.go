package transfer

import (
	"context"
	"errors"
	walletapp "vault/internal/application/wallet"
	"vault/internal/domain"
)

type TransferUsecase struct {
	walletOwnershipChecker walletOwnershipChecker
	unitOfWork             walletapp.UnitOfWork
}

func New(walletOwnershipChecker walletOwnershipChecker, unitOfWork walletapp.UnitOfWork) *TransferUsecase {
	return &TransferUsecase{
		walletOwnershipChecker: walletOwnershipChecker,
		unitOfWork:             unitOfWork,
	}
}

func (u *TransferUsecase) Do(ctx context.Context, command TransferCommand) (domain.Transaction, error) {
	var tx domain.Transaction
	if err := u.walletOwnershipChecker.IsOwnedBy(ctx, command.WalletFromID, command.UserID); err != nil {
		return tx, err
	}

	err := u.unitOfWork.StartTransaction(ctx, func(repos walletapp.TransactionalResources) (error) {
		// Check transaction existence by idempotency key. Type check not needed
		// Second transaction for "transfer_in" not needed (Atomicity)
		// TODO: что здесь отображают ошибки? как это отлаживать-то?
		existsTransaction, err := repos.TransactionRepository().GetByIdempotencyKey(ctx, command.IdempotencyKey)
		if err == nil {
			tx = existsTransaction
			return nil
		} else if !errors.Is(err, domain.ErrTransactionNotFound) {
			return err
		}

		walletFrom, walletTo, err := u.lockWallets(ctx, repos, command.WalletFromID, command.WalletToID)
		if err != nil {
			return err
		}

		// Withdrawal money from original wallet
		err = walletFrom.Withdraw(command.Amount)
		if err != nil {
			return err
		}

		err = repos.WalletRepository().Update(ctx, &walletFrom)
		if err != nil {
			return err
		}

		transaction, err := domain.NewTransaction(
			command.WalletFromID,
			domain.TransactionTypeTransferOut,
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

		// Deposit money to recipient wallet
		err = walletTo.Deposit(command.Amount)
		if err != nil {
			return err
		}

		err = repos.WalletRepository().Update(ctx, &walletTo)
		if err != nil {
			return err
		}

		transaction, err = domain.NewTransaction(
			command.WalletToID,
			domain.TransactionTypeTransferIn,
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

		return nil
	})

	return tx, err
}

// Lock wallets by IDs sorting
func (u *TransferUsecase) lockWallets(ctx context.Context, repos walletapp.TransactionalResources, idFrom, idTo int64) (domain.Wallet, domain.Wallet, error) {
	firstIsFrom := idFrom > idTo

	var firstID, secondID int64
	if firstIsFrom {
		firstID = idFrom
		secondID = idTo
	} else {
		firstID = idTo
		secondID = idFrom
	}

	wallet1, err := repos.WalletRepository().GetByIdForUpdate(ctx, firstID)
	if err != nil {
		return domain.Wallet{}, domain.Wallet{}, err
	}

	wallet2, err := repos.WalletRepository().GetByIdForUpdate(ctx, secondID)
	if err != nil {
		return domain.Wallet{}, domain.Wallet{}, err
	}

	if firstIsFrom {
		return wallet1, wallet2, nil
	}

	return wallet2, wallet1, nil
}
