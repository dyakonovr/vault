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
