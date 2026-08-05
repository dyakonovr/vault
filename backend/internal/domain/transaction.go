package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeDeposit     TransactionType = "deposit"
	TransactionTypeWithdrawal  TransactionType = "withdrawal"
	TransactionTypeTransferOut TransactionType = "transfer_out"
	TransactionTypeTransferIn  TransactionType = "transfer_in"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

type Transaction struct {
	ID             int64
	WalletID       int64
	Type           TransactionType
	Amount         float64
	IdempotencyKey uuid.UUID
	Status         TransactionStatus
	CreatedAt      time.Time
}
