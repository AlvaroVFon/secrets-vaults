// Package roles
package roles

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Role struct {
	ID   string
	Name string
}

var ErrEmptyArgument = errors.New("invalid empty argument")

func NewRole(name string) (*Role, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "name")
	}

	id := uuid.New().String()

	return &Role{
		ID:   id,
		Name: name,
	}, nil
}
