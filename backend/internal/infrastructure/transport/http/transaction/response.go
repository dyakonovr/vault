package transaction

import (
	"time"
	"vault/internal/domain"
)

type TransactionResponse struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

func NewTransactionResponse(t domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:        t.ID,
		Type:      string(t.Type),
		Status:    string(t.Status),
		Amount:    t.Amount,
		CreatedAt: t.CreatedAt,
	}
}
