// Package roles
package roles

import (
	"context"
	"errors"
	"testing"

	"secrets-vault/internal/tests"
)

func setupRoleRepo(t *testing.T) *RoleRepository {
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

	return NewRoleRepository(conn)
}

func TestRoleRepository_Create(t *testing.T) {
	repo := setupRoleRepo(t)
	ctx := context.Background()

	role, err := NewRole("admin")
	if err != nil {
		t.Fatalf("new role: %v", err)
	}

	if err := repo.Create(ctx, *role); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.FindByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if *got != *role {
		t.Fatalf("expected %+v, got %+v", *role, *got)
	}
}

func TestRoleRepository_Create_DuplicateName(t *testing.T) {
	repo := setupRoleRepo(t)
	ctx := context.Background()

	role, err := NewRole("admin")
	if err != nil {
		t.Fatalf("new role: %v", err)
	}

	if err := repo.Create(ctx, *role); err != nil {
		t.Fatalf("create: %v", err)
	}

	dup, err := NewRole("admin")
	if err != nil {
		t.Fatalf("new role: %v", err)
	}

	if err := repo.Create(ctx, *dup); err == nil {
		t.Fatal("expected error on duplicate name, got nil")
	}
}

func TestRoleRepository_FindByID(t *testing.T) {
	repo := setupRoleRepo(t)
	ctx := context.Background()

	role, err := NewRole("superadmin")
	if err != nil {
		t.Fatalf("new role: %v", err)
	}

	if err := repo.Create(ctx, *role); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.FindByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if *got != *role {
		t.Fatalf("expected %+v, got %+v", *role, *got)
	}
}

func TestRoleRepository_FindByID_NotFound(t *testing.T) {
	repo := setupRoleRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrorNotFound) {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestRoleRepository_FindByID_EmptyID(t *testing.T) {
	repo := setupRoleRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}
