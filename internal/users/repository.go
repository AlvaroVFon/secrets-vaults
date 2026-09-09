package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrorNotFound = errors.New("not found")

type UsersRepository struct {
	store *pgxpool.Pool
}

func NewUsersRepository(store *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{
		store: store,
	}
}

func (r *UsersRepository) Create(ctx context.Context, user User) error {
	query := "INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3)"
	_, err := r.store.Exec(ctx, query, user.ID, user.Username, user.PasswordHash)
	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository) FindAll(ctx context.Context) ([]User, error) {
	query := "SELECT id, username, password_hash, created_at FROM users ORDER BY username"

	users := make([]User, 0)

	rows, err := r.store.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		var createdAt time.Time
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAt); err != nil {
			return nil, err
		}
		user.CreatedAt = createdAt.Format(time.RFC3339)
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UsersRepository) FindByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidUUID, id)
	}

	query := "SELECT id, username, password_hash, created_at FROM users WHERE id=$1"

	var user User
	var createdAt time.Time
	err := r.store.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w user with id: %q", ErrorNotFound, id)
		}
		return nil, err
	}
	user.CreatedAt = createdAt.Format(time.RFC3339)

	return &user, nil
}

func (r *UsersRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "username")
	}

	query := "SELECT id, username, password_hash, created_at FROM users WHERE username=$1"

	var user User
	var createdAt time.Time
	err := r.store.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w user with username: %q", ErrorNotFound, username)
		}
		return nil, err
	}
	user.CreatedAt = createdAt.Format(time.RFC3339)

	return &user, nil
}

func (r *UsersRepository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM users"

	var count int
	err := r.store.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
