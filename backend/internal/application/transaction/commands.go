package transaction

import (
	"vault/internal/application"

	"github.com/google/uuid"
)

type ListTransactionsCommand struct {
	WalletID int64
	UserID   int64
	Status   *string
	Type     *string
	application.PaginationCommand
}

type GetTransactionByIdCommand struct {
	WalletID int64
	UserID   int64
}

type CreateTransactionCommand struct {
	WalletID       int64
	UserID         int64
	Type           string
	Amount         int64
	IdempotencyKey uuid.UUID
}
