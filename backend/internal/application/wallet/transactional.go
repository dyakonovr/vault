package wallet

import (
	"context"
	"vault/internal/domain"

	"github.com/google/uuid"
)

// ----------- UNIT OF WORK -----------

type WalletRepository interface {
	GetById(ctx context.Context, id int64) (domain.Wallet, error)
	Update(ctx context.Context, wallet *domain.Wallet) error
}

type TransactionRepository interface {
	GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (domain.Transaction, error)
	Create(ctx context.Context, transaction *domain.Transaction) error
}

type TransactionalResources interface {
	WalletRepository() WalletRepository
	TransactionRepository() TransactionRepository
}

type UnitOfWork interface {
	StartTransaction(ctx context.Context, caller func(resources TransactionalResources) error) error
}
