package postgres

import (
	"context"
	"time"
	"vault/internal/domain"
	"vault/internal/infrastructure/persistence"

	"gorm.io/gorm"
)

var userDomainErrors = DbDomainErrorsMap{
	persistence.ErrDBNoRows:          domain.ErrUserNotFound,
	persistence.ErrDBUniqueViolation: domain.ErrUserAlreadyExists,
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) List(ctx context.Context, params ListUsersParams) ([]domain.User, int64, error) {
	var users []UserModel
	var count int64

	query := r.db.WithContext(ctx).Model(&UserModel{})

	if params.Login != nil {
		query = query.Where("login = ?", params.Login)
	}

	// Сначала считаем с фильтром
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, mapDbErrorToDomain(err, userDomainErrors, false)
	}

	// Потом получаем страницу
	err := query.
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Order("id ASC").
		Find(&users).Error
	if err != nil {
		return nil, 0, mapDbErrorToDomain(err, userDomainErrors, true)
	}

	res := make([]domain.User, 0, len(users))
	for _, u := range users {
		res = append(res, mapUserToDomain(u))
	}
	return res, count, nil
}

func (r *UserRepository) GetById(ctx context.Context, id int64) (domain.User, error) {
	var user UserModel

	query := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user)
	if query.Error != nil {
		return domain.User{}, mapDbErrorToDomain(query.Error, userDomainErrors, false)
	}

	return mapUserToDomain(user), nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (domain.User, error) {
	var user UserModel

	query := r.db.WithContext(ctx).
		Where("login = ?", login).
		Take(&user)
	if query.Error != nil {
		return domain.User{}, mapDbErrorToDomain(query.Error, userDomainErrors, false)
	}

	return mapUserToDomain(user), nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	userModel := mapDomainToUserModel(user)

	query := r.db.WithContext(ctx).Create(&userModel)
	if query.Error != nil {
		return mapDbErrorToDomain(query.Error, userDomainErrors, false)
	}

	user.ID = userModel.ID
	user.CreatedAt = userModel.CreatedAt

	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	userModel := mapDomainToUserModel(user)
	if err := r.db.WithContext(ctx).Save(&userModel).Error; err != nil {
		return mapDbErrorToDomain(err, userDomainErrors, false)
	}

	user.UpdatedAt = time.Now()

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&UserModel{})
	return mapDbErrorToDomain(result.Error, userDomainErrors, false)
}

func mapDomainToUserModel(u *domain.User) UserModel {
	return UserModel{
		ID:           u.ID,
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func mapUserToDomain(u UserModel) domain.User {
	return domain.User{
		ID:           u.ID,
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
