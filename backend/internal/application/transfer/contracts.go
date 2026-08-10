package transfer

import (
	"context"
)

type walletOwnershipChecker interface {
	IsOwnedBy(ctx context.Context, walletID, userID int64) error
}
