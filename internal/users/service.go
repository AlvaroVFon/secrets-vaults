// Package users
package users

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Password string `json:"password" validate:"required,min=6,max=255"`
}

type repository interface {
	Create(ctx context.Context, user User) error
	FindAll(ctx context.Context) ([]User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Count(ctx context.Context) (int, error)
}

type UsersService struct {
	repository repository
}

func NewUsersService(repo repository) *UsersService {
	return &UsersService{
		repository: repo,
	}
}

func (s *UsersService) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
	if strings.TrimSpace(req.Username) == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "username")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := NewUser(req.Username, string(hash))
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, *user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersService) Authenticate(ctx context.Context, username, password string) (*User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "credentials")
	}

	user, err := s.repository.FindByUsername(ctx, username)
	if err != nil {
		if !isNotFound(err) {
			return nil, err
		}
		// No revelar si el usuario existe: compara contra un hash dummy.
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$invalidhashforcomparison"), []byte(password))
		return nil, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil
	}

	return user, nil
}

func (s *UsersService) FindByID(ctx context.Context, id string) (*User, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *UsersService) FindAll(ctx context.Context) ([]User, error) {
	return s.repository.FindAll(ctx)
}

func (s *UsersService) Count(ctx context.Context) (int, error) {
	return s.repository.Count(ctx)
}

func isNotFound(err error) bool {
	for err != nil {
		if err == ErrorNotFound {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return false
}
