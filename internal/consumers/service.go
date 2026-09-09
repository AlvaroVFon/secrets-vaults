// Package consumers
package consumers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type CreateConsumerRequest struct {
	Name   string `json:"name" validate:"required"`
	Apikey string `json:"apikey" validate:"required"`
	RoleID string `json:"roleId" validate:"required,uuid"`
}

type UpdateConsumerRequest struct {
	ID     string
	Name   *string
	Apikey *string
	RoleID *string
	Active *bool
}

type repository interface {
	Create(ctx context.Context, consumer Consumer) error
	FindAll(ctx context.Context) ([]Consumer, error)
	FindByID(ctx context.Context, id string) (*Consumer, error)
	FindByApikey(ctx context.Context, apikey string) (*Consumer, error)
	Update(ctx context.Context, req UpdateConsumerRequest) error
	DeleteByID(ctx context.Context, id string) error
}

type ConsumersService struct {
	repository repository
}

func NewConsumersService(repo repository) *ConsumersService {
	return &ConsumersService{
		repository: repo,
	}
}

func (s *ConsumersService) Create(ctx context.Context, req CreateConsumerRequest) (*Consumer, error) {
	consumer, err := NewConsumer(req.Name, req.Apikey, req.RoleID)
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, *consumer); err != nil {
		return nil, err
	}

	return consumer, nil
}

func (s *ConsumersService) DeleteByID(ctx context.Context, id string) error {
	return s.repository.DeleteByID(ctx, id)
}

func (s *ConsumersService) Update(ctx context.Context, req UpdateConsumerRequest) (*Consumer, error) {
	if req.Name != nil && *req.Name == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "name")
	}
	if req.Apikey != nil && *req.Apikey == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "apikey")
	}
	if req.RoleID != nil {
		if *req.RoleID == "" {
			return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "roleID")
		}
		if _, err := uuid.Parse(*req.RoleID); err != nil {
			return nil, fmt.Errorf("%w: %q", ErrInvalidUUID, *req.RoleID)
		}
	}

	if err := s.repository.Update(ctx, req); err != nil {
		return nil, err
	}

	return s.repository.FindByID(ctx, req.ID)
}

func (s *ConsumersService) FindByApikey(ctx context.Context, apikey string) (*Consumer, error) {
	return s.repository.FindByApikey(ctx, apikey)
}

func (s *ConsumersService) FindAll(ctx context.Context) ([]Consumer, error) {
	return s.repository.FindAll(ctx)
}

func (s *ConsumersService) FindByID(ctx context.Context, id string) (*Consumer, error) {
	return s.repository.FindByID(ctx, id)
}
