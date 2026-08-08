package postgres

import (
	"time"

	"github.com/google/uuid"
)

type UserModel struct {
	ID           int64     `gorm:"primaryKey"`
	Login        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

type WalletModel struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"uniqueIndex;not null"`
	Balance   int64     `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type TransactionModel struct {
	ID             int64     `gorm:"primaryKey"`
	WalletID       int64     `gorm:"not null;index"`
	Type           string    `gorm:"not null;size:20"`
	Amount         int64     `gorm:"not null"`
	IdempotencyKey uuid.UUID `gorm:"type:uuid;not null"`
	Status         string    `gorm:"not null;size:20"`
	CreatedAt      time.Time `gorm:"not null"`
}
