package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start container: %v\n", err)
		os.Exit(1)
	}

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get host: %v\n", err)
		os.Exit(1)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get port: %v\n", err)
		os.Exit(1)
	}

	dsn := fmt.Sprintf("host=%s user=test password=test dbname=test port=%s sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	migrationsPath := filepath.Join("..", "..", "..", "..", "migrations")
	migrationSQL, err := os.ReadFile(filepath.Join(migrationsPath, "000001_init.up.sql"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read migration file: %v\n", err)
		os.Exit(1)
	}

	if err := db.Exec(string(migrationSQL)).Error; err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migration: %v\n", err)
		os.Exit(1)
	}

	testDB = db

	code := m.Run()

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	container.Terminate(ctx)

	os.Exit(code)
}

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testDB == nil {
		t.Fatal("test DB not initialized")
	}
	return testDB
}

func cleanTable(t *testing.T, db *gorm.DB, table string) {
	t.Helper()
	db.Exec("DELETE FROM " + table)
}
