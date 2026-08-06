package wallet

import (
	"time"
	"vault/internal/domain"
)

type BalanceResponse struct {
	Balance int64 `json:"balance"`
}

func NewBalanceResponse(balance int64) BalanceResponse {
	return BalanceResponse{
		Balance: balance,
	}
}

type WalletResponse struct {
	ID        int64     `json:"id"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewWalletResponse(w domain.Wallet) WalletResponse {
	return WalletResponse{
		ID:        w.ID,
		Balance:   w.Balance,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}
