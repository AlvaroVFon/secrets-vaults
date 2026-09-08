package secrets

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("not found")

type SecretsRepository struct {
	store *pgx.Conn
}

func NewSecretsRepository(store *pgx.Conn) *SecretsRepository {
	return &SecretsRepository{
		store: store,
	}
}

func (r *SecretsRepository) Create(ctx context.Context, secret Secret) error {
	query := "INSERT INTO secrets (id, key, value, consumer_id) VALUES ($1, $2, $3, $4)"
	if _, err := r.store.Exec(ctx, query, secret.ID, secret.Key, secret.Value, secret.ConsumerID); err != nil {
		return err
	}
	return nil
}

func (r *SecretsRepository) FindAllByConsumerID(ctx context.Context, consumerID string) ([]Secret, error) {
	if consumerID == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "consumerID")
	}
	if _, err := uuid.Parse(consumerID); err != nil {
		return nil, err
	}

	query := "SELECT key, value FROM secrets WHERE consumer_id=$1"
	secrets := make([]Secret, 0)

	rows, err := r.store.Query(ctx, query, consumerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var secret Secret
		if err := rows.Scan(&secret.Key, &secret.Value); err != nil {
			return nil, err
		}

		secrets = append(secrets, secret)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return secrets, nil
}
