package domain

import (
	"errors"
	"time"
	"vault/pkg/utils"

	"github.com/google/uuid"
)

// ------------ ERRORS ------------

var (
	ErrTransactionNotFound          = errors.New("transaction not found")
	ErrTransactionAlreadyExists     = errors.New("transaction already exists")
	ErrTransactionIncorrectWalletID = errors.New("incorrect transaction walletID")
	ErrTransactionIncorrectType     = errors.New("incorrect transaction type")
	ErrTransactionInvalidAmount     = errors.New("transaction amount must be greater than zero")
	ErrInvalidStatusTransition      = errors.New("invalid transaction status transition")
)

// ------------ TRANSACTION TYPES ------------

type TransactionType string

const (
	TransactionTypeDeposit     TransactionType = "deposit"
	TransactionTypeWithdrawal  TransactionType = "withdrawal"
	TransactionTypeTransferOut TransactionType = "transfer_out"
	TransactionTypeTransferIn  TransactionType = "transfer_in"
)

var (
	availableTransactionTypes = []TransactionType{TransactionTypeDeposit, TransactionTypeWithdrawal, TransactionTypeTransferOut, TransactionTypeTransferIn}
)

// ------------ TRANSACTION STATUSES ------------

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

var (
	availableTransactionStatuses = []TransactionStatus{TransactionStatusPending, TransactionStatusCompleted, TransactionStatusFailed}
)

// ------------ MODEL AND CONSTRUCTOR ------------

type Transaction struct {
	ID             int64
	WalletID       int64
	Type           TransactionType
	Amount         int64
	IdempotencyKey uuid.UUID
	// Статус не используется. При необходимости
	// исправить присвоение в конструкторе
	Status         TransactionStatus
	CreatedAt      time.Time
}

func NewTransaction(walletID int64, type_ TransactionType, amount int64, idempotencyKey uuid.UUID) (*Transaction, error) {
	if walletID <= 0 {
		return nil, ErrTransactionIncorrectWalletID
	}

	if !utils.Contains(availableTransactionTypes, type_) {
		return nil, ErrTransactionIncorrectType
	}

	if amount <= 0 {
		return nil, ErrTransactionInvalidAmount
	}

	return &Transaction{
		WalletID:       walletID,
		Type:           type_,
		Amount:         amount,
		IdempotencyKey: idempotencyKey,
		Status:         TransactionStatusCompleted,
	}, nil
}

// func (t *Transaction) MarkCompleted() error {
// 	if t.Status != TransactionStatusPending {
// 		return ErrInvalidStatusTransition
// 	}
// 	t.Status = TransactionStatusCompleted
// 	return nil
// }
