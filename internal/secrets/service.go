package secrets

import (
	"context"
	"fmt"
)

type CreateSecretRequest struct {
	Key        string `json:"key" validate:"required"`
	Value      string `json:"value" validate:"required"`
	ConsumerID string `json:"consumer_id" validate:"required,uuid"`
}

type UpdateSecretRequest struct {
	ID    string  `json:"id" validate:"required,uuid"`
	Key   *string `json:"key" validate:"omitempty"`
	Value *string `json:"value" validate:"omitempty"`
}

type repository interface {
	Create(ctx context.Context, secret Secret) error
	FindAll(ctx context.Context) ([]Secret, error)
	FindByID(ctx context.Context, id string) (*Secret, error)
	FindAllByConsumerID(ctx context.Context, consumerID string) ([]Secret, error)
	Update(ctx context.Context, req UpdateSecretRequest) error
	DeleteByID(ctx context.Context, id string) error
}

type SecretsService struct {
	repository repository
}

func NewSecretsService(repo repository) *SecretsService {
	return &SecretsService{
		repository: repo,
	}
}

func (s *SecretsService) Create(ctx context.Context, createSecretRequest CreateSecretRequest) (*Secret, error) {
	secret, err := NewSecret(createSecretRequest.Key, createSecretRequest.Value, createSecretRequest.ConsumerID)
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, *secret); err != nil {
		return nil, err
	}

	return secret, nil
}

func (s *SecretsService) FindAll(ctx context.Context) ([]Secret, error) {
	return s.repository.FindAll(ctx)
}

func (s *SecretsService) FindByID(ctx context.Context, id string) (*Secret, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *SecretsService) FindAllByConsumerID(ctx context.Context, consumerID string) ([]Secret, error) {
	return s.repository.FindAllByConsumerID(ctx, consumerID)
}

func (s *SecretsService) Update(ctx context.Context, req UpdateSecretRequest) (*Secret, error) {
	if req.ID == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}
	if req.Key != nil && *req.Key == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "key")
	}
	if req.Value != nil && *req.Value == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "value")
	}

	if err := s.repository.Update(ctx, req); err != nil {
		return nil, err
	}

	return s.repository.FindByID(ctx, req.ID)
}

func (s *SecretsService) DeleteByID(ctx context.Context, id string) error {
	return s.repository.DeleteByID(ctx, id)
}
