// Package users
package users

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	createFunc     func(ctx context.Context, user User) error
	findAllFunc    func(ctx context.Context) ([]User, error)
	findByIDFunc   func(ctx context.Context, id string) (*User, error)
	findByUsername func(ctx context.Context, username string) (*User, error)
	countFunc      func(ctx context.Context) (int, error)
}

func (m *mockRepository) Create(ctx context.Context, user User) error {
	return m.createFunc(ctx, user)
}

func (m *mockRepository) FindAll(ctx context.Context) ([]User, error) {
	return m.findAllFunc(ctx)
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*User, error) {
	return m.findByIDFunc(ctx, id)
}

func (m *mockRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	return m.findByUsername(ctx, username)
}

func (m *mockRepository) Count(ctx context.Context) (int, error) {
	return m.countFunc(ctx)
}

func hashedUser(t *testing.T, username, password string) *User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	u, err := NewUser(username, string(hash))
	if err != nil {
		t.Fatalf("new user: %v", err)
	}
	return u
}

func TestUsersService_Create(t *testing.T) {
	var got User
	service := NewUsersService(&mockRepository{
		createFunc: func(_ context.Context, user User) error {
			got = user
			return nil
		},
	})

	user, err := service.Create(context.Background(), CreateUserRequest{
		Username: "admin",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if user == nil {
		t.Fatal("expected created user, got nil")
	}
	if user.ID == "" {
		t.Error("expected non-empty ID")
	}
	if user.Username != "admin" {
		t.Errorf("expected Username %q, got %q", "admin", user.Username)
	}
	if user.PasswordHash == "" {
		t.Error("expected PasswordHash to be set")
	}
	if got.ID != user.ID {
		t.Errorf("repository received user with ID %q, returned %q", got.ID, user.ID)
	}
}

func TestUsersService_Create_EmptyUsername(t *testing.T) {
	service := NewUsersService(&mockRepository{})

	_, err := service.Create(context.Background(), CreateUserRequest{
		Password: "secret123",
	})
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestUsersService_Create_EmptyPassword(t *testing.T) {
	service := NewUsersService(&mockRepository{})

	_, err := service.Create(context.Background(), CreateUserRequest{
		Username: "admin",
	})
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestUsersService_Create_RepositoryError(t *testing.T) {
	expectErr := errors.New("repo failed")
	service := NewUsersService(&mockRepository{
		createFunc: func(_ context.Context, _ User) error {
			return expectErr
		},
	})

	_, err := service.Create(context.Background(), CreateUserRequest{
		Username: "admin",
		Password: "secret123",
	})
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}

func TestUsersService_Authenticate_Valid(t *testing.T) {
	user := hashedUser(t, "admin", "secret123")

	service := NewUsersService(&mockRepository{
		findByUsername: func(_ context.Context, username string) (*User, error) {
			return user, nil
		},
	})

	got, err := service.Authenticate(context.Background(), "admin", "secret123")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if got == nil {
		t.Fatal("expected authenticated user, got nil")
	}
	if got.ID != user.ID {
		t.Errorf("expected ID %q, got %q", user.ID, got.ID)
	}
}

func TestUsersService_Authenticate_WrongPassword(t *testing.T) {
	user := hashedUser(t, "admin", "secret123")

	service := NewUsersService(&mockRepository{
		findByUsername: func(_ context.Context, username string) (*User, error) {
			return user, nil
		},
	})

	got, err := service.Authenticate(context.Background(), "admin", "wrong-password")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil user, got %+v", got)
	}
}

func TestUsersService_Authenticate_UnknownUser(t *testing.T) {
	service := NewUsersService(&mockRepository{
		findByUsername: func(_ context.Context, username string) (*User, error) {
			return nil, ErrorNotFound
		},
	})

	got, err := service.Authenticate(context.Background(), "ghost", "secret123")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil user, got %+v", got)
	}
}

func TestUsersService_Authenticate_EmptyCredentials(t *testing.T) {
	service := NewUsersService(&mockRepository{})

	_, err := service.Authenticate(context.Background(), "", "")
	if !errors.Is(err, ErrEmptyArgument) {
		t.Fatalf("expected ErrEmptyArgument, got %v", err)
	}
}

func TestUsersService_Authenticate_RepositoryError(t *testing.T) {
	expectErr := errors.New("db down")
	service := NewUsersService(&mockRepository{
		findByUsername: func(_ context.Context, username string) (*User, error) {
			return nil, expectErr
		},
	})

	_, err := service.Authenticate(context.Background(), "admin", "secret123")
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected %v, got %v", expectErr, err)
	}
}

func TestUsersService_FindByID(t *testing.T) {
	expected := hashedUser(t, "admin", "secret123")
	service := NewUsersService(&mockRepository{
		findByIDFunc: func(_ context.Context, id string) (*User, error) {
			return expected, nil
		},
	})

	got, err := service.FindByID(context.Background(), expected.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
}

func TestUsersService_Count(t *testing.T) {
	service := NewUsersService(&mockRepository{
		countFunc: func(_ context.Context) (int, error) {
			return 0, nil
		},
	})

	count, err := service.Count(context.Background())
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestUsersService_NewUsersService(t *testing.T) {
	repo := &mockRepository{}
	service := NewUsersService(repo)

	if service == nil {
		t.Fatal("expected non-nil service")
	}
	if service.repository != repo {
		t.Errorf("expected repository %+v, got %+v", repo, service.repository)
	}
}
