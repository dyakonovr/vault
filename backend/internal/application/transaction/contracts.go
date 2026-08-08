package transaction

import (
	"context"
	"vault/internal/domain"
)

type transactionRepository interface {
	List(ctx context.Context, params ListTransactionsCommand) ([]domain.Transaction, int64, error)
	GetById(ctx context.Context, id int64) (domain.Transaction, error)
	Create(ctx context.Context, transaction *domain.Transaction) error
}

type walletOwnershipChecker interface {
	IsOwnedBy(ctx context.Context, walletID, userID int64) error
}