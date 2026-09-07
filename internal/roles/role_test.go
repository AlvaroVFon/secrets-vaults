// Package roles
package roles

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewRole(t *testing.T) {
	role, err := NewRole("admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if role.ID == "" {
		t.Error("expected non-empty ID")
	}
	if _, err := uuid.Parse(role.ID); err != nil {
		t.Errorf("expected valid UUID, got %q: %v", role.ID, err)
	}
	if role.Name != "admin" {
		t.Errorf("expected Name %q, got %q", "admin", role.Name)
	}
}

func TestNewRole_Superadmin(t *testing.T) {
	role, err := NewRole("superadmin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "superadmin" {
		t.Errorf("expected Name %q, got %q", "superadmin", role.Name)
	}
}

func TestNewRole_EmptyName(t *testing.T) {
	_, err := NewRole("")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestNewRole_InvalidRole(t *testing.T) {
	_, err := NewRole("editor")
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}
