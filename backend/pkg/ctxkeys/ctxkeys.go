package ctxkeys

type contextKey string

const (
	RequestIDKey   contextKey = "requestID"
	UserIDKey      contextKey = "userID"
	IdempotencyKey contextKey = "userID"
)
