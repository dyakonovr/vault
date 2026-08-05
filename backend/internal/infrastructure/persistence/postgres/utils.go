package postgres

import (
	"errors"
	"vault/internal/infrastructure/persistence"
)

// DbDomainErrorsMap — маппинг storage-ошибок в доменные
type DbDomainErrorsMap map[error]error

// mapDbErrorToDomain преобразует ошибку БД в доменную.
// Если ignoreNoRows == true, то ErrDBNoRows превращается в nil (списки, где пустой результат не ошибка).
func mapDbErrorToDomain(dbErr error, errorsMap DbDomainErrorsMap, ignoreNoRows bool) error {
	if dbErr == nil {
		return nil
	}

	mappedErr := mapGormError(dbErr)

	if ignoreNoRows && errors.Is(mappedErr, persistence.ErrDBNoRows) {
		return nil
	}

	if e, ok := errorsMap[mappedErr]; ok {
		return e
	}
	return mappedErr
}
