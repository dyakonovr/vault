package common

import (
	"context"
	"vault/pkg/ctxkeys"

	"github.com/google/uuid"
)

func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxkeys.RequestIDKey).(string)
	return id, ok
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxkeys.UserIDKey).(int64)
	return id, ok
}

func GetIdempotencyKeyFromContext(ctx context.Context) (uuid.UUID, bool) {
	key, ok := ctx.Value(ctxkeys.IdempotencyKey).(uuid.UUID)
	return key, ok
}
