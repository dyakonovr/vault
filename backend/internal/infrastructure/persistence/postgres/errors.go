package postgres

import (
	"context"
	"errors"
	"io"
	"net"
	"syscall"
	"vault/internal/infrastructure/persistence"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func mapGormError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return persistence.ErrDBNoRows
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return persistence.ErrDBUniqueViolation
		case "23503":
			return persistence.ErrDBForeignKeyViolation
		}
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return persistence.ErrDBTimeout
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return persistence.ErrDBTimeout
	}
	if errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) {
		return persistence.ErrDBConnection
	}

	return err // неизвестная ошибка
}
