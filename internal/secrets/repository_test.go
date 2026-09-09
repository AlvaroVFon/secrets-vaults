// Package secrets
package secrets

import (
	"context"
	"errors"
	"testing"

	"secrets-vault/internal/consumers"
	"secrets-vault/internal/roles"
	"secrets-vault/internal/tests"
)

func setupSecretsRepo(t *testing.T) *SecretsRepository {
	t.Helper()

	conn, err := tests.Connect(t)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	t.Cleanup(func() {
		if err := tests.CleanDB(t, conn); err != nil {
			t.Logf("cleanup DB: %v", err)
		}
		conn.Close()
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

func TestSecretsRepository_CreateAndFindAllByConsumerID(t *testing.T) {
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

	got, err := repo.FindAllByConsumerID(ctx, consumerID)
	if err != nil {
		t.Fatalf("find all by consumerID: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(got))
	}
	if got[0].Key != secret.Key {
		t.Errorf("expected Key %q, got %q", secret.Key, got[0].Key)
	}
	if got[0].Value != secret.Value {
		t.Errorf("expected Value %q, got %q", secret.Value, got[0].Value)
	}
}

func TestSecretsRepository_FindAllByConsumerID(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	consumerID := createConsumerForSecret(t, repo)

	expected := []Secret{
		{Key: "db.password", Value: "s3cret"},
		{Key: "db.host", Value: "localhost"},
		{Key: "api.key", Value: "key-123"},
	}

	for _, s := range expected {
		secret, err := NewSecret(s.Key, s.Value, consumerID)
		if err != nil {
			t.Fatalf("new secret: %v", err)
		}
		if err := repo.Create(ctx, *secret); err != nil {
			t.Fatalf("create secret: %v", err)
		}
	}

	got, err := repo.FindAllByConsumerID(ctx, consumerID)
	if err != nil {
		t.Fatalf("find all by consumerID: %v", err)
	}

	if len(got) != len(expected) {
		t.Fatalf("expected %d secrets, got %d", len(expected), len(got))
	}

	for i, s := range expected {
		if got[i].Key != s.Key {
			t.Errorf("secret[%d] expected Key %q, got %q", i, s.Key, got[i].Key)
		}
		if got[i].Value != s.Value {
			t.Errorf("secret[%d] expected Value %q, got %q", i, s.Value, got[i].Value)
		}
	}
}

func TestSecretsRepository_FindAllByConsumerID_EmptyConsumerID(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	_, err := repo.FindAllByConsumerID(ctx, "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestSecretsRepository_FindAllByConsumerID_InvalidConsumerID(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	_, err := repo.FindAllByConsumerID(ctx, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid consumerID, got nil")
	}
}

func TestSecretsRepository_FindAllByConsumerID_NoSecrets(t *testing.T) {
	repo := setupSecretsRepo(t)
	ctx := context.Background()

	consumerID := createConsumerForSecret(t, repo)

	got, err := repo.FindAllByConsumerID(ctx, consumerID)
	if err != nil {
		t.Fatalf("find all by consumerID: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected 0 secrets, got %d", len(got))
	}
}
