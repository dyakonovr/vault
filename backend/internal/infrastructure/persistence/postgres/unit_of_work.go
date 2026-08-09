package postgres

import (
	"context"
	"vault/internal/application/deposit"

	"gorm.io/gorm"
)

type UnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{
		db: db,
	}
}

func (uow *UnitOfWork) StartTransaction(ctx context.Context, caller func(resources deposit.TransactionalResources) error) error {
	return uow.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return caller(DepositRepositories{
			walletRepo:      NewWalletRepository(tx),
			transactionRepo: NewTransactionRepository(tx),
		})
	})
}

type DepositRepositories struct {
	walletRepo      *WalletRepository
	transactionRepo *TransactionRepository
}

func (r DepositRepositories) WalletRepository() deposit.WalletRepository           { return r.walletRepo }
func (r DepositRepositories) TransactionRepository() deposit.TransactionRepository { return r.transactionRepo }
