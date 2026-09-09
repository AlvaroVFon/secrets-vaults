// Package users
package users

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrEmptyArgument = errors.New("invalid empty argument")
	ErrInvalidUUID   = errors.New("invalid UUID provided")
)

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"createdAt,omitempty"`
}

func NewUser(username, passwordHash string) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "username")
	}
	if passwordHash == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "passwordHash")
	}

	id := uuid.New().String()

	return &User{
		ID:           id,
		Username:     strings.TrimSpace(username),
		PasswordHash: passwordHash,
	}, nil
}
