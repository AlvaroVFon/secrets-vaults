//go:build integration

// Package secrets
package secrets

import (
	"context"
	"errors"
	"testing"

	"secrets-vault/internal/consumers"
	"secrets-vault/internal/roles"
	"secrets-vault/internal/tests"

	"github.com/google/uuid"
)

func setupSecretsRepo(t *testing.T) *SecretsRepository {
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

	return NewSecretsRepository(conn)
}

func createConsumerForSecret(t *testing.T, repo *SecretsRepository) string {
	t.Helper()
	ctx := context.Background()

	roleRepo := roles.NewRoleRepository(repo.store)
	role, err := roles.NewRole("admin")
	if err != nil {
		t.Fatalf("new role: %v", err)
	}
	if err := roleRepo.Create(ctx, *role); err != nil {
		t.Fatalf("create role: %v", err)
	}

	consumerRepo := consumers.NewConsumersRepository(repo.store)
	consumer, err := consumers.NewConsumer("secret-consumer", "secret-api-key", role.ID)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	if err := consumerRepo.Create(ctx, *consumer); err != nil {
		t.Fatalf("create consumer: %v", err)
	}

	return consumer.ID
}

func TestSecretsRepository_CreateAndFindByKeyAndConsumerID(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	consumerID := createConsumerForSecret(t, repo)

	secret, err := NewSecret("db.password", "s3cret", consumerID)
	if err != nil {
		t.Fatalf("new secret: %v", err)
	}

	if err := repo.Create(ctx, *secret); err != nil {
		t.Fatalf("create secret: %v", err)
	}

	got, err := repo.FindByKeyAndConsumerID(ctx, secret.Key, consumerID)
	if err != nil {
		t.Fatalf("find by key and consumerID: %v", err)
	}
	if got.Key != secret.Key {
		t.Errorf("expected Key %q, got %q", secret.Key, got.Key)
	}
	if got.Value != secret.Value {
		t.Errorf("expected Value %q, got %q", secret.Value, got.Value)
	}
	if got.ConsumerID != secret.ConsumerID {
		t.Errorf("expected ConsumerID %q, got %q", secret.ConsumerID, got.ConsumerID)
	}
}

func TestSecretsRepository_FindByKeyAndConsumerID_NotFound(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	consumerID := createConsumerForSecret(t, repo)

	_, err := repo.FindByKeyAndConsumerID(ctx, "nonexistent-key", consumerID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSecretsRepository_FindByKeyAndConsumerID_EmptyKey(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	_, err := repo.FindByKeyAndConsumerID(ctx, "", uuid.New().String())
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestSecretsRepository_FindByKeyAndConsumerID_EmptyConsumerID(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	_, err := repo.FindByKeyAndConsumerID(ctx, "db.password", "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestSecretsRepository_FindByKeyAndConsumerID_InvalidConsumerID(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	_, err := repo.FindByKeyAndConsumerID(ctx, "db.password", "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid consumerID, got nil")
	}
}
