package common

import (
	"context"
	"vault/pkg/ctxkeys"
)

func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxkeys.RequestIDKey).(string)
	return id, ok
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxkeys.UserIDKey).(int64)
	return id, ok
}
