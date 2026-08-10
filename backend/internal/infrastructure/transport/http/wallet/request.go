package wallet

import "github.com/google/uuid"

type DepositRequest struct {
	Amount         int64     `json:"amount"          validate:"required,min=1"`
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required,uuid"`
}

type WithdrawRequest struct {
	Amount         int64     `json:"amount"          validate:"required,min=1"`
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required,uuid"`
}

type TransferRequest struct {
	WalletFromID   int64     `json:"wallet_from_id" validate:"required,min=1"`
	WalletToID     int64     `json:"wallet_to_id" validate:"required,min=1"`
	Amount         int64     `json:"amount"          validate:"required,min=1"`
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required,uuid"`
}
