package withdrawal

import "github.com/google/uuid"

type WithdrawalCommand struct {
	UserID         int64
	WalletID       int64
	IdempotencyKey uuid.UUID
	Amount         int64
}
