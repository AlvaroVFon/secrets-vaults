package secrets

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type SecretsRepository struct {
	store *pgxpool.Pool
}

func NewSecretsRepository(store *pgxpool.Pool) *SecretsRepository {
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

func (r *SecretsRepository) FindAll(ctx context.Context) ([]Secret, error) {
	query := "SELECT id, key, value, consumer_id FROM secrets"
	secrets := make([]Secret, 0)

	rows, err := r.store.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var secret Secret
		if err := rows.Scan(&secret.ID, &secret.Key, &secret.Value, &secret.ConsumerID); err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return secrets, nil
}

func (r *SecretsRepository) FindByID(ctx context.Context, id string) (*Secret, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidUUID, id)
	}

	query := "SELECT id, key, value, consumer_id FROM secrets WHERE id=$1"

	var secret Secret
	err := r.store.QueryRow(ctx, query, id).Scan(&secret.ID, &secret.Key, &secret.Value, &secret.ConsumerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w secret with id: %q", ErrNotFound, id)
		}
		return nil, err
	}

	return &secret, nil
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

func (r *SecretsRepository) Update(ctx context.Context, req UpdateSecretRequest) error {
	if req.ID == "" {
		return fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}
	if _, err := uuid.Parse(req.ID); err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidUUID, req.ID)
	}

	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)

	if req.Key != nil {
		args = append(args, *req.Key)
		sets = append(sets, fmt.Sprintf("key=$%d", len(args)))
	}
	if req.Value != nil {
		args = append(args, *req.Value)
		sets = append(sets, fmt.Sprintf("value=$%d", len(args)))
	}

	if len(sets) == 0 {
		return nil
	}

	args = append(args, req.ID)
	query := fmt.Sprintf("UPDATE secrets SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))

	tag, err := r.store.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w secret with id: %q", ErrNotFound, req.ID)
	}

	return nil
}

func (r *SecretsRepository) DeleteByID(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidUUID, id)
	}

	query := "DELETE FROM secrets WHERE id=$1"

	tag, err := r.store.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w secret with id: %q", ErrNotFound, id)
	}

	return nil
}
