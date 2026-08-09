package app

import (
	"vault/internal/application/deposit"
	"vault/internal/infrastructure/persistence/postgres"
	"vault/pkg/logger"

	gormPostgres "gorm.io/driver/postgres"

	"gorm.io/gorm"
)

func Run() {
	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable"
	db, err := gorm.Open(gormPostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("database not opened: %w", err)
	}

	unitOfWork := postgres.NewUnitOfWork(db)

	depositUsecase := deposit.New(nil, unitOfWork)
}
