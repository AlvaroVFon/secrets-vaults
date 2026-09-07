// Package consumers
package consumers

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewConsumer(t *testing.T) {
	roleID := uuid.New().String()

	consumer, err := NewConsumer("consumer1", "secret", roleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if consumer.ID == "" {
		t.Error("expected non-empty ID")
	}
	if _, err := uuid.Parse(consumer.ID); err != nil {
		t.Errorf("expected valid UUID, got %q: %v", consumer.ID, err)
	}
	if consumer.Name != "consumer1" {
		t.Errorf("expected Name %q, got %q", "consumer1", consumer.Name)
	}
	if consumer.Apikey != "secret" {
		t.Errorf("expected Apikey %q, got %q", "secret", consumer.Apikey)
	}
	if consumer.RoleID != roleID {
		t.Errorf("expected RoleID %q, got %q", roleID, consumer.RoleID)
	}
}

func TestNewConsumer_EmptyName(t *testing.T) {
	_, err := NewConsumer("", "secret", uuid.New().String())
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestNewConsumer_EmptyApikey(t *testing.T) {
	_, err := NewConsumer("consumer1", "", uuid.New().String())
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestNewConsumer_EmptyRoleID(t *testing.T) {
	_, err := NewConsumer("consumer1", "secret", "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestNewConsumer_InvalidRoleID(t *testing.T) {
	_, err := NewConsumer("consumer1", "secret", "not-a-uuid")
	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID, got %v", err)
	}
}
