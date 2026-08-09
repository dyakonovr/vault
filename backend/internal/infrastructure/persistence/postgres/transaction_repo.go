package postgres

import (
	"context"
	transactionapp "vault/internal/application/transaction"
	"vault/internal/domain"
	"vault/internal/infrastructure/persistence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var transactionDomainErrors = DbDomainErrorsMap{
	persistence.ErrDBNoRows:          domain.ErrTransactionNotFound,
	persistence.ErrDBUniqueViolation: domain.ErrTransactionAlreadyExists,
}

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (r *TransactionRepository) List(ctx context.Context, params transactionapp.ListTransactionsCommand) ([]domain.Transaction, int64, error) {
	var transactions []TransactionModel
	var count int64

	query := r.db.WithContext(ctx).Model(&TransactionModel{}).Where("wallet_id = ?", params.WalletID)

	if params.Status != nil {
		query = query.Where("status = ?", params.Status)
	}
	if params.Type != nil {
		query = query.Where("type = ?", params.Type)
	}

	// Сначала считаем с фильтром
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, mapDbErrorToDomain(err, transactionDomainErrors, false)
	}

	// Потом получаем страницу
	err := query.
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Order("id ASC").
		Find(&transactions).Error
	if err != nil {
		return nil, 0, mapDbErrorToDomain(err, transactionDomainErrors, true)
	}

	res := make([]domain.Transaction, 0, len(transactions))
	for _, u := range transactions {
		res = append(res, mapTransactionToDomain(u))
	}
	return res, count, nil
}

func (r *TransactionRepository) GetById(ctx context.Context, id int64) (domain.Transaction, error) {
	var transaction TransactionModel

	query := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&transaction)
	if query.Error != nil {
		return domain.Transaction{}, mapDbErrorToDomain(query.Error, transactionDomainErrors, false)
	}

	return mapTransactionToDomain(transaction), nil
}

func (r *TransactionRepository) GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (domain.Transaction, error) {
	var transaction TransactionModel

	query := r.db.WithContext(ctx).
		Where("idempotency_key = ?", key).
		First(&transaction)
	if query.Error != nil {
		return domain.Transaction{}, mapDbErrorToDomain(query.Error, transactionDomainErrors, false)
	}

	return mapTransactionToDomain(transaction), nil
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	transactionModel := mapDomainToTransactionModel(transaction)

	query := r.db.WithContext(ctx).Create(&transactionModel)
	if query.Error != nil {
		return mapDbErrorToDomain(query.Error, transactionDomainErrors, false)
	}

	transaction.ID = transactionModel.ID
	transaction.CreatedAt = transactionModel.CreatedAt

	return nil
}

func mapDomainToTransactionModel(t *domain.Transaction) TransactionModel {
	return TransactionModel{
		ID:             t.ID,
		WalletID:       t.WalletID,
		Type:           string(t.Type),
		Amount:         t.Amount,
		IdempotencyKey: t.IdempotencyKey,
		Status:         string(t.Status),
		CreatedAt:      t.CreatedAt,
	}
}

func mapTransactionToDomain(t TransactionModel) domain.Transaction {
	return domain.Transaction{
		ID:             t.ID,
		WalletID:       t.WalletID,
		Type:           domain.TransactionType(t.Type),
		Amount:         t.Amount,
		IdempotencyKey: t.IdempotencyKey,
		Status:         domain.TransactionStatus(t.Status),
		CreatedAt:      t.CreatedAt,
	}
}
