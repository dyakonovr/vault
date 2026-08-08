package transaction

import (
	"context"
	transactionapp "vault/internal/application/transaction"
	"vault/internal/domain"
)

type transactionService interface {
	List(ctx context.Context, command transactionapp.ListTransactionsCommand) ([]domain.Transaction, int64, error)
	GetById(ctx context.Context, id int64, command transactionapp.GetTransactionByIdCommand) (domain.Transaction, error)
}
