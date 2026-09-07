// Package secrets
package secrets

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewSecret(t *testing.T) {
	consumerID := uuid.New().String()

	secret, err := NewSecret("db.password", "s3cret", consumerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if secret.ID == "" {
		t.Error("expected non-empty ID")
	}
	if _, err := uuid.Parse(secret.ID); err != nil {
		t.Errorf("expected valid UUID, got %q: %v", secret.ID, err)
	}
	if secret.Key != "db.password" {
		t.Errorf("expected Key %q, got %q", "db.password", secret.Key)
	}
	if secret.Value != "s3cret" {
		t.Errorf("expected Value %q, got %q", "s3cret", secret.Value)
	}
	if secret.ConsumerID != consumerID {
		t.Errorf("expected ConsumerID %q, got %q", consumerID, secret.ConsumerID)
	}
}

func TestNewSecret_EmptyKey(t *testing.T) {
	_, err := NewSecret("", "s3cret", uuid.New().String())
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestNewSecret_EmptyValue(t *testing.T) {
	_, err := NewSecret("db.password", "", uuid.New().String())
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestNewSecret_EmptyConsumerID(t *testing.T) {
	_, err := NewSecret("db.password", "s3cret", "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestNewSecret_InvalidConsumerID(t *testing.T) {
	_, err := NewSecret("db.password", "s3cret", "not-a-uuid")
	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID, got %v", err)
	}
}
