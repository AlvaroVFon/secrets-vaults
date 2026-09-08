package secrets

import "context"

type CreateSecretRequest struct {
	Key        string `json:"key" validate:"required"`
	Value      string `json:"value" validate:"required"`
	ConsumerID string `json:"consumer_id" validate:"required,uuid"`
}

type repository interface {
	Create(ctx context.Context, secret Secret) error
	FindAllByConsumerID(ctx context.Context, consummerID string) ([]Secret, error)
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

func (s *SecretsService) FindAllByConsumerID(ctx context.Context, consumerID string) ([]Secret, error) {
	return s.repository.FindAllByConsumerID(ctx, consumerID)
}
