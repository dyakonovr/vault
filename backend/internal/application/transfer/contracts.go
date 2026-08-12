package transfer

import (
	"context"
)

//go:generate mockery
type WalletOwnershipChecker interface {
	IsOwnedBy(ctx context.Context, walletID, userID int64) error
}
