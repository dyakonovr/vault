package common

import (
	"context"
	"vault/pkg/ctxkeys"
)

func GetRequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxkeys.RequestIDKey).(string)
	return id
}

func GetUserIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(ctxkeys.UserIDKey).(int64)
	return id
}
