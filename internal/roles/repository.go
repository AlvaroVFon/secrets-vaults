package roles

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrorNotFound = errors.New("not found")

type RoleRepository struct {
	store *pgxpool.Pool
}

func NewRoleRepository(conn *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{
		store: conn,
	}
}

func (r *RoleRepository) Create(ctx context.Context, role Role) error {
	query := "INSERT INTO roles (ID, name) VALUES ($1, $2)"
	if _, err := r.store.Exec(ctx, query, role.ID, role.Name); err != nil {
		return err
	}

	return nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id string) (*Role, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}

	query := "SELECT * FROM roles WHERE ID=$1"

	var role Role
	err := r.store.QueryRow(ctx, query, id).Scan(&role.ID, &role.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w role with ID: %q", ErrorNotFound, id)
		}
		return nil, err
	}

	return &role, nil
}
