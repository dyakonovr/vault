package common

import (
	"context"
	"vault/pkg/ctxkeys"
)

func GetRequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxkeys.RequestIDKey).(string)
	return id
}
