package deposit

import "github.com/google/uuid"

type DepositCommand struct {
	UserID         int64
	WalletID       int64
	IdempotencyKey uuid.UUID
	Amount         int64
}
