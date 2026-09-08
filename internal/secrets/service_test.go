// Package secrets
package secrets

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

type mockRepository struct {
	createFunc  func(ctx context.Context, secret Secret) error
	findAllFunc func(ctx context.Context, consumerID string) ([]Secret, error)
}

func (m *mockRepository) Create(ctx context.Context, secret Secret) error {
	return m.createFunc(ctx, secret)
}

func (m *mockRepository) FindAllByConsumerID(ctx context.Context, consumerID string) ([]Secret, error) {
	return m.findAllFunc(ctx, consumerID)
}

func TestSecretsService_Create(t *testing.T) {
	consumerID := uuid.New().String()

	var got Secret
	service := NewSecretsService(&mockRepository{
		createFunc: func(_ context.Context, secret Secret) error {
			got = secret
			return nil
		},
	})

	secret, err := service.Create(context.Background(), CreateSecretRequest{
		Key:        "db.password",
		Value:      "s3cret",
		ConsumerID: consumerID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if secret == nil {
		t.Fatal("expected created secret, got nil")
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
	if secret.ID == "" {
		t.Error("expected non-empty ID")
	}
	if got.ID != secret.ID {
		t.Errorf("repository received secret with ID %q, returned %q", got.ID, secret.ID)
	}
}

func TestSecretsService_Create_EmptyKey(t *testing.T) {
	service := NewSecretsService(&mockRepository{})

	_, err := service.Create(context.Background(), CreateSecretRequest{
		Value:      "s3cret",
		ConsumerID: uuid.New().String(),
	})
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestSecretsService_Create_RepositoryError(t *testing.T) {
	expectErr := errors.New("repo failed")
	service := NewSecretsService(&mockRepository{
		createFunc: func(_ context.Context, _ Secret) error {
			return expectErr
		},
	})

	_, err := service.Create(context.Background(), CreateSecretRequest{
		Key:        "db.password",
		Value:      "s3cret",
		ConsumerID: uuid.New().String(),
	})
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}

func TestSecretsService_FindAllByConsumerID(t *testing.T) {
	consumerID := uuid.New().String()
	expected := []Secret{
		{Key: "db.password", Value: "s3cret"},
		{Key: "db.host", Value: "localhost"},
	}

	service := NewSecretsService(&mockRepository{
		findAllFunc: func(_ context.Context, gotConsumerID string) ([]Secret, error) {
			if gotConsumerID != consumerID {
				t.Errorf("expected consumerID %q, got %q", consumerID, gotConsumerID)
			}
			return expected, nil
		},
	})

	got, err := service.FindAllByConsumerID(context.Background(), consumerID)
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %+v, got %+v", expected, got)
	}
}

func TestSecretsService_FindAllByConsumerID_Error(t *testing.T) {
	expectErr := errors.New("repo failed")
	service := NewSecretsService(&mockRepository{
		findAllFunc: func(_ context.Context, _ string) ([]Secret, error) {
			return nil, expectErr
		},
	})

	_, err := service.FindAllByConsumerID(context.Background(), uuid.New().String())
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}
