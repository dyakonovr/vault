package transaction

import (
	"context"
	"vault/internal/domain"
)

type TransactionUsecase struct {
	transactionRepo        transactionRepository
	walletOwnershipChecker walletOwnershipChecker
}

func New(transactionRepo transactionRepository, walletOwnershipChecker walletOwnershipChecker) *TransactionUsecase {
	return &TransactionUsecase{
		transactionRepo:        transactionRepo,
		walletOwnershipChecker: walletOwnershipChecker,
	}
}

func (u *TransactionUsecase) List(ctx context.Context, command ListTransactionsCommand) ([]domain.Transaction, int64, error) {
	if err := u.walletOwnershipChecker.IsOwnedBy(ctx, command.WalletID, command.UserID); err != nil {
		return []domain.Transaction{}, 0, err
	}

	return u.transactionRepo.List(ctx, command)
}

func (u *TransactionUsecase) GetById(ctx context.Context, id int64, command GetTransactionByIdCommand) (domain.Transaction, error) {
	if err := u.walletOwnershipChecker.IsOwnedBy(ctx, command.WalletID, command.UserID); err != nil {
		return domain.Transaction{}, err
	}

	return u.transactionRepo.GetById(ctx, id)
}

func (u *TransactionUsecase) Create(ctx context.Context, command CreateTransactionCommand) (domain.Transaction, error) {
	if err := u.walletOwnershipChecker.IsOwnedBy(ctx, command.WalletID, command.UserID); err != nil {
		return domain.Transaction{}, err
	}

	transaction, err := domain.NewTransaction(command.WalletID, command.Type, command.Amount, command.IdempotencyKey)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := u.transactionRepo.Create(ctx, transaction); err != nil {
		return domain.Transaction{}, err
	}
	return *transaction, nil
}
