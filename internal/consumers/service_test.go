// Package consumers
package consumers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockRepository struct {
	createFunc       func(ctx context.Context, consumer Consumer) error
	findByIDFunc     func(ctx context.Context, id string) (*Consumer, error)
	findByApikeyFunc func(ctx context.Context, apikey string) (*Consumer, error)
	updateFunc       func(ctx context.Context, req UpdateConsumerRequest) error
	deleteFunc       func(ctx context.Context, id string) error
}

func (m *mockRepository) Create(ctx context.Context, consumer Consumer) error {
	return m.createFunc(ctx, consumer)
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*Consumer, error) {
	return m.findByIDFunc(ctx, id)
}

func (m *mockRepository) FindByApikey(ctx context.Context, apikey string) (*Consumer, error) {
	return m.findByApikeyFunc(ctx, apikey)
}

func (m *mockRepository) Update(ctx context.Context, req UpdateConsumerRequest) error {
	return m.updateFunc(ctx, req)
}

func (m *mockRepository) DeleteByID(ctx context.Context, id string) error {
	return m.deleteFunc(ctx, id)
}

func TestConsumersService_Create(t *testing.T) {
	roleID := uuid.New().String()

	var got Consumer
	service := NewConsumersService(&mockRepository{
		createFunc: func(_ context.Context, consumer Consumer) error {
			got = consumer
			return nil
		},
	})

	consumer, err := service.Create(context.Background(), CreateConsumerRequest{
		Name:   "consumer1",
		Apikey: "secret-key-123",
		RoleID: roleID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if consumer == nil {
		t.Fatal("expected created consumer, got nil")
	}

	if consumer.ID == "" {
		t.Error("expected non-empty ID")
	}
	if consumer.Name != "consumer1" {
		t.Errorf("expected Name %q, got %q", "consumer1", consumer.Name)
	}
	if consumer.Apikey != "secret-key-123" {
		t.Errorf("expected Apikey %q, got %q", "secret-key-123", consumer.Apikey)
	}
	if consumer.RoleID != roleID {
		t.Errorf("expected RoleID %q, got %q", roleID, consumer.RoleID)
	}
	if got.ID != consumer.ID {
		t.Errorf("repository received consumer with ID %q, returned %q", got.ID, consumer.ID)
	}
}

func TestConsumersService_Create_EmptyName(t *testing.T) {
	service := NewConsumersService(&mockRepository{})

	_, err := service.Create(context.Background(), CreateConsumerRequest{
		Apikey: "secret-key-123",
		RoleID: uuid.New().String(),
	})
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestConsumersService_Create_RepositoryError(t *testing.T) {
	expectErr := errors.New("repo failed")
	service := NewConsumersService(&mockRepository{
		createFunc: func(_ context.Context, _ Consumer) error {
			return expectErr
		},
	})

	_, err := service.Create(context.Background(), CreateConsumerRequest{
		Name:   "consumer1",
		Apikey: "secret-key-123",
		RoleID: uuid.New().String(),
	})
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}

func TestConsumersService_DeleteByID(t *testing.T) {
	id := uuid.New().String()

	service := NewConsumersService(&mockRepository{
		deleteFunc: func(_ context.Context, gotID string) error {
			if gotID != id {
				t.Errorf("expected id %q, got %q", id, gotID)
			}
			return nil
		},
	})

	if err := service.DeleteByID(context.Background(), id); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestConsumersService_DeleteByID_InvalidID(t *testing.T) {
	service := NewConsumersService(&mockRepository{
		deleteFunc: func(_ context.Context, _ string) error {
			return ErrInvalidUUID
		},
	})

	err := service.DeleteByID(context.Background(), "not-a-uuid")
	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID, got %v", err)
	}
}

func TestConsumersService_DeleteByID_RepositoryError(t *testing.T) {
	expectErr := errors.New("repo failed")
	service := NewConsumersService(&mockRepository{
		deleteFunc: func(_ context.Context, _ string) error {
			return expectErr
		},
	})

	err := service.DeleteByID(context.Background(), uuid.New().String())
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}

func TestConsumersService_Update(t *testing.T) {
	id := uuid.New().String()
	roleID := uuid.New().String()
	name := "renamed"

	expected := &Consumer{
		ID:     id,
		Name:   name,
		Apikey: "secret-key-123",
		RoleID: roleID,
		Active: true,
	}

	var gotReq UpdateConsumerRequest
	service := NewConsumersService(&mockRepository{
		updateFunc: func(_ context.Context, req UpdateConsumerRequest) error {
			gotReq = req
			return nil
		},
		findByIDFunc: func(_ context.Context, gotID string) (*Consumer, error) {
			if gotID != id {
				t.Errorf("expected id %q, got %q", id, gotID)
			}
			return expected, nil
		},
	})

	consumer, err := service.Update(context.Background(), UpdateConsumerRequest{
		ID:   id,
		Name: &name,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if consumer == nil {
		t.Fatal("expected updated consumer, got nil")
	}

	if gotReq.ID != id {
		t.Errorf("expected ID %q, got %q", id, gotReq.ID)
	}
	if gotReq.Name == nil || *gotReq.Name != name {
		t.Errorf("expected Name %q, got %+v", name, gotReq.Name)
	}
	if gotReq.Apikey != nil {
		t.Errorf("expected Apikey nil, got %+v", *gotReq.Apikey)
	}
	if gotReq.RoleID != nil {
		t.Errorf("expected RoleID nil, got %+v", *gotReq.RoleID)
	}
	if gotReq.Active != nil {
		t.Errorf("expected Active nil, got %+v", *gotReq.Active)
	}
}

func TestConsumersService_Update_InvalidID(t *testing.T) {
	service := NewConsumersService(&mockRepository{
		updateFunc: func(_ context.Context, _ UpdateConsumerRequest) error {
			return ErrInvalidUUID
		},
	})

	_, err := service.Update(context.Background(), UpdateConsumerRequest{ID: "not-a-uuid"})
	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID, got %v", err)
	}
}

func TestConsumersService_Update_EmptyName(t *testing.T) {
	service := NewConsumersService(&mockRepository{})

	name := ""
	_, err := service.Update(context.Background(), UpdateConsumerRequest{
		ID:   uuid.New().String(),
		Name: &name,
	})
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestConsumersService_Update_InvalidRoleID(t *testing.T) {
	service := NewConsumersService(&mockRepository{})

	roleID := "not-a-uuid"
	_, err := service.Update(context.Background(), UpdateConsumerRequest{
		ID:     uuid.New().String(),
		RoleID: &roleID,
	})
	if !errors.Is(err, ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID, got %v", err)
	}
}

func TestConsumersService_Update_RepositoryError(t *testing.T) {
	expectErr := errors.New("repo failed")
	service := NewConsumersService(&mockRepository{
		updateFunc: func(_ context.Context, _ UpdateConsumerRequest) error {
			return expectErr
		},
	})

	name := "renamed"
	_, err := service.Update(context.Background(), UpdateConsumerRequest{
		ID:   uuid.New().String(),
		Name: &name,
	})
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}

func TestConsumersService_Update_FindByIDError(t *testing.T) {
	expectErr := errors.New("repo failed")
	service := NewConsumersService(&mockRepository{
		updateFunc: func(_ context.Context, _ UpdateConsumerRequest) error {
			return nil
		},
		findByIDFunc: func(_ context.Context, _ string) (*Consumer, error) {
			return nil, expectErr
		},
	})

	_, err := service.Update(context.Background(), UpdateConsumerRequest{
		ID: uuid.New().String(),
	})
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}

func TestConsumersService_NewConsumersService(t *testing.T) {
	repo := &mockRepository{}
	service := NewConsumersService(repo)

	if service == nil {
		t.Fatal("expected non-nil service")
	}
	if service.repository != repo {
		t.Errorf("expected repository %+v, got %+v", repo, service.repository)
	}
}
