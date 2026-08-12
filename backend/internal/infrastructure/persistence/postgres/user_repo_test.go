package postgres

import (
	"context"
	"testing"
	"vault/internal/application"
	userapp "vault/internal/application/user"
	"vault/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "users")
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		Login:        "alice",
		PasswordHash: "hashed_password",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	require.NotZero(t, user.ID)
	require.False(t, user.CreatedAt.IsZero())
	require.False(t, user.UpdatedAt.IsZero())
}

func TestFindByID(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "users")
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "bob", PasswordHash: "hash"}
	repo.Create(ctx, user)

	found, err := repo.GetById(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "bob", found.Login)
	require.Equal(t, "hash", found.PasswordHash)
}

func TestFindByLogin(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "users")
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "charlie", PasswordHash: "hash"}
	repo.Create(ctx, user)

	found, err := repo.GetByLogin(ctx, "charlie")
	require.NoError(t, err)
	require.Equal(t, user.ID, found.ID)
	require.Equal(t, "charlie", found.Login)
}

func TestFindByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "users")
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetById(ctx, 999)
	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestCreate_DuplicateLogin(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "users")
	repo := NewUserRepository(db)
	ctx := context.Background()

	user1 := &domain.User{Login: "duplicate", PasswordHash: "hash1"}
	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	user2 := &domain.User{Login: "duplicate", PasswordHash: "hash2"}
	err = repo.Create(ctx, user2)
	require.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestUpdate(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "users")
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "dave", PasswordHash: "old_hash"}
	repo.Create(ctx, user)

	user.PasswordHash = "new_hash"
	err := repo.Update(ctx, user)
	require.NoError(t, err)

	found, err := repo.GetById(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "new_hash", found.PasswordHash)
}

func TestDelete(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "users")
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{Login: "eve", PasswordHash: "hash"}
	repo.Create(ctx, user)

	err := repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	_, err = repo.GetById(ctx, user.ID)
	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestList(t *testing.T) {
	db := SetupTestDB(t)
	cleanTable(t, db, "users")
	repo := NewUserRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.User{Login: "user1", PasswordHash: "h1"})
	repo.Create(ctx, &domain.User{Login: "user2", PasswordHash: "h2"})
	repo.Create(ctx, &domain.User{Login: "user3", PasswordHash: "h3"})

	users, total, err := repo.List(ctx, userapp.ListUsersCommand{
		PaginationCommand: application.PaginationCommand{Offset: 0, Limit: 2},
	})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, users, 2)
}
