package domain

import (
	"errors"
	"time"
	"vault/pkg/utils"

	"github.com/google/uuid"
)

// ------------ ERRORS ------------ 

var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrTransactionAlreadyExists = errors.New("transaction already exists")
	ErrTransactionIncorrectType = errors.New("incorrect transaction type")
	ErrTransactionInvalidAmount = errors.New("transaction amount must be greater than zero")
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
	availableTransactionTypes    = []TransactionType{TransactionTypeDeposit, TransactionTypeWithdrawal, TransactionTypeTransferOut, TransactionTypeTransferIn}
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
	Status         TransactionStatus
	CreatedAt      time.Time
}

func NewTransaction(walletID int64, type_ string, amount int64, idempotencyKey uuid.UUID) (*Transaction, error) {
	if !utils.Contains(availableTransactionTypes, TransactionType(type_)) {
		return nil, ErrTransactionIncorrectType
	}

	if amount <= 0 {
		return nil, ErrTransactionInvalidAmount
	}

	return &Transaction{
		WalletID:       walletID,
		Type:           TransactionType(type_),
		Amount:         amount,
		IdempotencyKey: idempotencyKey,
		Status:         TransactionStatusPending,
	}, nil
}
