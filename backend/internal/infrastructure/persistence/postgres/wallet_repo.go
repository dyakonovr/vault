package postgres

import (
	"context"
	"time"
	"vault/internal/domain"
	"vault/internal/infrastructure/persistence"

	"gorm.io/gorm"
)

var walletDomainErrors = DbDomainErrorsMap{
	persistence.ErrDBNoRows:          domain.ErrWalletNotFound,
	persistence.ErrDBUniqueViolation: domain.ErrWalletAlreadyExists,
}

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{
		db: db,
	}
}

func (r *WalletRepository) GetById(ctx context.Context, id int64) (domain.Wallet, error) {
	var wallet WalletModel

	query := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&wallet)
	if query.Error != nil {
		return domain.Wallet{}, mapDbErrorToDomain(query.Error, walletDomainErrors, false)
	}

	return mapWalletToDomain(wallet), nil
}

func (r *WalletRepository) GetByLogin(ctx context.Context, login string) (domain.Wallet, error) {
	var wallet WalletModel

	query := r.db.WithContext(ctx).
		Where("login = ?", login).
		Take(&wallet)
	if query.Error != nil {
		return domain.Wallet{}, mapDbErrorToDomain(query.Error, walletDomainErrors, false)
	}

	return mapWalletToDomain(wallet), nil
}

func (r *WalletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	walletModel := mapDomainToWalletModel(wallet)

	query := r.db.WithContext(ctx).Create(&walletModel)
	if query.Error != nil {
		return mapDbErrorToDomain(query.Error, walletDomainErrors, false)
	}

	wallet.ID = walletModel.ID
	wallet.CreatedAt = walletModel.CreatedAt

	return nil
}

func (r *WalletRepository) Update(ctx context.Context, wallet *domain.Wallet) error {
	walletModel := mapDomainToWalletModel(wallet)
	if err := r.db.WithContext(ctx).Save(&walletModel).Error; err != nil {
		return mapDbErrorToDomain(err, walletDomainErrors, false)
	}

	wallet.UpdatedAt = time.Now()

	return nil
}

func mapDomainToWalletModel(w *domain.Wallet) WalletModel {
	return WalletModel{
		ID:        w.ID,
		UserID:    w.UserID,
		Balance:   w.Balance,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func mapWalletToDomain(w WalletModel) domain.Wallet {
	return domain.Wallet{
		ID:        w.ID,
		UserID:    w.UserID,
		Balance:   w.Balance,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}
