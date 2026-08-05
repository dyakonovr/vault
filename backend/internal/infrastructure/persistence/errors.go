package persistence

import "errors"

var (
	ErrDBConnection          = errors.New("database connection error")
	ErrDBTimeout             = errors.New("database timeout")
	ErrDBUniqueViolation     = errors.New("unique constraint violation")
	ErrDBForeignKeyViolation = errors.New("foreign key violation")
	ErrDBNoRows              = errors.New("record(s) not found")
)
