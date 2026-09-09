// Package secrets
package secrets

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type mockRepository struct {
	createFunc              func(ctx context.Context, secret Secret) error
	updateFunc              func(ctx context.Context, req UpdateSecretRequest) error
	deleteFunc              func(ctx context.Context, id string) error
	findByIDFunc            func(ctx context.Context, id string) (*Secret, error)
	findAllFunc             func(ctx context.Context) ([]Secret, error)
	findAllByConsumerIDFunc func(ctx context.Context, consumerID string) ([]Secret, error)
}

func (m *mockRepository) Create(ctx context.Context, secret Secret) error {
	return m.createFunc(ctx, secret)
}

func (m *mockRepository) Update(ctx context.Context, req UpdateSecretRequest) error {
	return m.updateFunc(ctx, req)
}

func (m *mockRepository) DeleteByID(ctx context.Context, id string) error {
	return m.deleteFunc(ctx, id)
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*Secret, error) {
	return m.findByIDFunc(ctx, id)
}

func (m *mockRepository) FindAll(ctx context.Context) ([]Secret, error) {
	return m.findAllFunc(ctx)
}

func (m *mockRepository) FindAllByConsumerID(ctx context.Context, consumerID string) ([]Secret, error) {
	return m.findAllByConsumerIDFunc(ctx, consumerID)
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
		findAllByConsumerIDFunc: func(_ context.Context, gotConsumerID string) ([]Secret, error) {
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
		findAllByConsumerIDFunc: func(_ context.Context, _ string) ([]Secret, error) {
			return nil, expectErr
		},
	})

	_, err := service.FindAllByConsumerID(context.Background(), uuid.New().String())
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}

func TestUpdateSecretRequest_Validate(t *testing.T) {
	validate := validator.New()

	key := "db.password"
	value := "s3cret"
	empty := ""

	tests := []struct {
		name    string
		req     UpdateSecretRequest
		wantErr bool
	}{
		{
			name:    "valid full",
			req:     UpdateSecretRequest{ID: uuid.New().String(), Key: &key, Value: &value},
			wantErr: false,
		},
		{
			name:    "valid only key",
			req:     UpdateSecretRequest{ID: uuid.New().String(), Key: &key},
			wantErr: false,
		},
		{
			name:    "valid only value",
			req:     UpdateSecretRequest{ID: uuid.New().String(), Value: &value},
			wantErr: false,
		},
		{
			name:    "missing id",
			req:     UpdateSecretRequest{Key: &key, Value: &value},
			wantErr: true,
		},
		{
			name:    "invalid id",
			req:     UpdateSecretRequest{ID: "not-a-uuid", Key: &key},
			wantErr: true,
		},
		{
			name:    "empty key pointer allowed by omitempty",
			req:     UpdateSecretRequest{ID: uuid.New().String(), Key: &empty},
			wantErr: false,
		},
		{
			name:    "nil key and value allowed",
			req:     UpdateSecretRequest{ID: uuid.New().String()},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.req)
			if tt.wantErr && err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no validation error, got %v", err)
			}
		})
	}
}
