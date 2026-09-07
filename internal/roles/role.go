// Package roles
package roles

import (
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

type Role struct {
	ID   string
	Name string
}

var (
	ErrEmptyArgument          = errors.New("invalid empty argument")
	ErrInvalidRole            = errors.New("invalid role provided")
	validRoles       []string = []string{"superadmin", "admin"}
)

func NewRole(name string) (*Role, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "name")
	}
	if !slices.Contains(validRoles, name) {
		return nil, fmt.Errorf("%w: %q, roles must be one of %v", ErrInvalidRole, name, validRoles)
	}

	id := uuid.New().String()

	return &Role{
		ID:   id,
		Name: name,
	}, nil
}
