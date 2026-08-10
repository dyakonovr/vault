package transfer

import "github.com/google/uuid"

type TransferCommand struct {
	WalletFromID   int64
	WalletToID     int64
	UserID         int64 // owner of walletFrom
	IdempotencyKey uuid.UUID
	Amount         int64
}
