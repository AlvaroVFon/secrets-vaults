// Package secrets
package secrets

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrEmptyArgument = errors.New("invalid empty argument")
	ErrInvalidUUID   = errors.New("invalid UUID provided")
)

type Secret struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Value      string `json:"value"`
	ConsumerID string `json:"consumerId"`
}

func NewSecret(key, value, consumerID string) (*Secret, error) {
	if key == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "key")
	}
	if value == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "value")
	}
	if consumerID == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "consumerID")
	}
	if _, err := uuid.Parse(consumerID); err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidUUID, consumerID)
	}

	id := uuid.New().String()

	return &Secret{
		ID:         id,
		Key:        key,
		Value:      value,
		ConsumerID: consumerID,
	}, nil
}
