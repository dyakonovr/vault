package deposit

import (
	"context"
	"time"
)

//go:generate mockery
type WalletOwnershipChecker interface {
	IsOwnedBy(ctx context.Context, walletID, userID int64) error
}

//go:generate mockery
type WalletLocker interface {
	Lock(ctx context.Context, walletID int64, timeout time.Duration) (func(), error)
}
