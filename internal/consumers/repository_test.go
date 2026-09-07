// Package consumers
package consumers

import (
	"context"
	"errors"
	"testing"

	"secrets-vault/internal/roles"

	"secrets-vault/internal/tests"
)

func setupConsumersRepo(t *testing.T) *ConsumersRepository {
	t.Helper()

	conn, err := tests.Connect(t)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	t.Cleanup(func() {
		ctx := context.Background()
		if err := tests.CleanDB(t, conn); err != nil {
			t.Logf("cleanup DB: %v", err)
		}
		if err := conn.Close(ctx); err != nil {
			t.Logf("close conn: %v", err)
		}
	})

	return NewConsumersRepository(conn)
}

func createRoleForConsumer(t *testing.T, repo *ConsumersRepository) string {
	t.Helper()

	ctx := context.Background()
	role, err := roles.NewRole("admin")
	if err != nil {
		t.Fatalf("new role: %v", err)
	}

	roleRepo := roles.NewRoleRepository(repo.store)
	if err := roleRepo.Create(ctx, *role); err != nil {
		t.Fatalf("create role: %v", err)
	}

	return role.ID
}

func TestConsumersRepository_CreateAndFindByApikey(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	roleID := createRoleForConsumer(t, repo)
	consumer, err := NewConsumer("consumer1", "secret-key-123", roleID)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}

	if err := repo.Create(ctx, *consumer); err != nil {
		t.Fatalf("create consumer: %v", err)
	}

	got, err := repo.FindByApikey(ctx, consumer.Apikey)
	if err != nil {
		t.Fatalf("find by apikey: %v", err)
	}
	if *got != *consumer {
		t.Fatalf("expected %+v, got %+v", *consumer, *got)
	}
}

func TestConsumersRepository_Create_DuplicateApikey(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	roleID := createRoleForConsumer(t, repo)
	first, err := NewConsumer("consumer1", "duplicate-key", roleID)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	if err := repo.Create(ctx, *first); err != nil {
		t.Fatalf("create consumer: %v", err)
	}

	dup, err := NewConsumer("consumer2", "duplicate-key", roleID)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	if err := repo.Create(ctx, *dup); err == nil {
		t.Fatal("expected error on duplicate apikey, got nil")
	}
}

func TestConsumersRepository_FindByApikey_NotFound(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByApikey(ctx, "nonexistent-key")
	if !errors.Is(err, ErrorNotFound) {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestConsumersRepository_FindByApikey_EmptyApikey(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByApikey(ctx, "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}
