// Package users
package users

import (
	"context"
	"errors"
	"testing"

	"secrets-vault/internal/tests"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupUsersRepo(t *testing.T) *UsersRepository {
	t.Helper()

	conn, err := tests.Connect(t)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	if err := tests.CleanDB(t, conn); err != nil {
		t.Fatalf("cleanup DB: %v", err)
	}

	t.Cleanup(func() {
		if err := tests.CleanDB(t, conn); err != nil {
			t.Logf("cleanup DB: %v", err)
		}
		conn.Close()
	})

	return NewUsersRepository(conn)
}

func newTestUser(t *testing.T, username string) User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := NewUser(username, string(hash))
	if err != nil {
		t.Fatalf("new user: %v", err)
	}
	return *user
}

func TestUsersRepository_CreateAndFindByUsername(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	user := newTestUser(t, "admin")

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	got, err := repo.FindByUsername(ctx, user.Username)
	if err != nil {
		t.Fatalf("find by username: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("expected ID %q, got %q", user.ID, got.ID)
	}
	if got.Username != user.Username {
		t.Errorf("expected Username %q, got %q", user.Username, got.Username)
	}
	if got.PasswordHash != user.PasswordHash {
		t.Errorf("expected PasswordHash %q, got %q", user.PasswordHash, got.PasswordHash)
	}
}

func TestUsersRepository_Create_DuplicateUsername(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	user := newTestUser(t, "admin")
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	dup := newTestUser(t, "admin")
	if err := repo.Create(ctx, dup); err == nil {
		t.Fatal("expected error on duplicate username, got nil")
	}
}

func TestUsersRepository_FindByUsername_NotFound(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByUsername(ctx, "nobody")
	if !errors.Is(err, ErrorNotFound) {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestUsersRepository_FindByUsername_EmptyUsername(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByUsername(ctx, "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestUsersRepository_FindByID(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	user := newTestUser(t, "admin")
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	got, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("expected ID %q, got %q", user.ID, got.ID)
	}
}

func TestUsersRepository_FindByID_NotFound(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.New().String())
	if !errors.Is(err, ErrorNotFound) {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestUsersRepository_FindByID_EmptyID(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestUsersRepository_FindByID_InvalidID(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "not-a-uuid")
	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID, got %v", err)
	}
}

func TestUsersRepository_FindAll(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	first := newTestUser(t, "alice")
	second := newTestUser(t, "bob")
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("create user: %v", err)
	}

	got, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 users, got %d", len(got))
	}
}

func TestUsersRepository_Count(t *testing.T) {
	repo := setupUsersRepo(t)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 users, got %d", count)
	}

	user := newTestUser(t, "admin")
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 user, got %d", count)
	}
}
