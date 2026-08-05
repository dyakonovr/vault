package domain

import "time"

type Wallet struct {
	ID        int64
	UserID    int64
	Balance   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
