// Package consumers
package consumers

import (
	"context"
	"errors"
	"testing"

	"secrets-vault/internal/roles"

	"secrets-vault/internal/tests"

	"github.com/google/uuid"
)

func setupConsumersRepo(t *testing.T) *ConsumersRepository {
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

func TestConsumersRepository_Create_DuplicateName(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	roleID := createRoleForConsumer(t, repo)
	first, err := NewConsumer("same-name", "key-1", roleID)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	if err := repo.Create(ctx, *first); err != nil {
		t.Fatalf("create consumer: %v", err)
	}

	dup, err := NewConsumer("same-name", "key-2", roleID)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	if err := repo.Create(ctx, *dup); err == nil {
		t.Fatal("expected error on duplicate name, got nil")
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

func TestConsumersRepository_FindByID(t *testing.T) {
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

	got, err := repo.FindByID(ctx, consumer.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if *got != *consumer {
		t.Fatalf("expected %+v, got %+v", *consumer, *got)
	}
}

func TestConsumersRepository_FindByID_NotFound(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.New().String())
	if !errors.Is(err, ErrorNotFound) {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestConsumersRepository_FindByID_EmptyID(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestConsumersRepository_FindByID_InvalidID(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "not-a-uuid")
	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID, got %v", err)
	}
}

func TestConsumersRepository_Update(t *testing.T) {
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

	name := "renamed"
	apikey := "new-key"
	active := true

	req := UpdateConsumerRequest{
		ID:     consumer.ID,
		Name:   &name,
		Apikey: &apikey,
		Active: &active,
	}
	if err := repo.Update(ctx, req); err != nil {
		t.Fatalf("update consumer: %v", err)
	}

	got, err := repo.FindByID(ctx, consumer.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.Name != name {
		t.Errorf("expected Name %q, got %q", name, got.Name)
	}
	if got.Apikey != apikey {
		t.Errorf("expected Apikey %q, got %q", apikey, got.Apikey)
	}
	if !got.Active {
		t.Error("expected Active true")
	}
	if got.RoleID != roleID {
		t.Errorf("expected RoleID %q, got %q", roleID, got.RoleID)
	}
}

func TestConsumersRepository_Update_Partial(t *testing.T) {
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

	name := "renamed"
	req := UpdateConsumerRequest{
		ID:   consumer.ID,
		Name: &name,
	}
	if err := repo.Update(ctx, req); err != nil {
		t.Fatalf("update consumer: %v", err)
	}

	got, err := repo.FindByID(ctx, consumer.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.Name != name {
		t.Errorf("expected Name %q, got %q", name, got.Name)
	}
	if got.Apikey != consumer.Apikey {
		t.Errorf("expected Apikey %q, got %q", consumer.Apikey, got.Apikey)
	}
	if got.RoleID != roleID {
		t.Errorf("expected RoleID %q, got %q", roleID, got.RoleID)
	}
}

func TestConsumersRepository_Update_NoFields(t *testing.T) {
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

	if err := repo.Update(ctx, UpdateConsumerRequest{ID: consumer.ID}); err != nil {
		t.Fatalf("update consumer: %v", err)
	}
}

func TestConsumersRepository_Update_NotFound(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	name := "rename"
	req := UpdateConsumerRequest{
		ID:   uuid.New().String(),
		Name: &name,
	}
	err := repo.Update(ctx, req)
	if !errors.Is(err, ErrorNotFound) {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestConsumersRepository_DeleteByID(t *testing.T) {
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

	if err := repo.DeleteByID(ctx, consumer.ID); err != nil {
		t.Fatalf("delete consumer: %v", err)
	}

	_, err = repo.FindByID(ctx, consumer.ID)
	if !errors.Is(err, ErrorNotFound) {
		t.Fatalf("expected ErrorNotFound after delete, got %v", err)
	}
}

func TestConsumersRepository_DeleteByID_NotFound(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	err := repo.DeleteByID(ctx, uuid.New().String())
	if !errors.Is(err, ErrorNotFound) {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestConsumersRepository_DeleteByID_EmptyID(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	err := repo.DeleteByID(ctx, "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestConsumersRepository_DeleteByID_InvalidID(t *testing.T) {
	repo := setupConsumersRepo(t)
	ctx := context.Background()

	err := repo.DeleteByID(ctx, "not-a-uuid")
	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID, got %v", err)
	}
}
