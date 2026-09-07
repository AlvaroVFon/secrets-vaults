// Package consumers
package consumers

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrEmptyArgument = errors.New("invalid empty argument")
	ErrInvalidUUID   = errors.New("invalid UUID provided")
)

type Consumer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Apikey string `json:"apikey"`
	RoleID string `json:"roleId"`
	Active bool   `json:"active"`
}

func NewConsumer(name, apikey, roleID string) (*Consumer, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "name")
	}
	if apikey == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "consumerSecret")
	}
	if roleID == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "roleID")
	}
	if _, err := uuid.Parse(roleID); err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidUUID, roleID)
	}

	id := uuid.New().String()

	return &Consumer{
		ID:     id,
		Name:   name,
		Apikey: apikey,
		RoleID: roleID,
	}, nil
}
