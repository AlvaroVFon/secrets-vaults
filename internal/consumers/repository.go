package consumers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrorNotFound = errors.New("not found")

type ConsumersRepository struct {
	store *pgx.Conn
}

func NewConsumersRepository(store *pgx.Conn) *ConsumersRepository {
	return &ConsumersRepository{
		store: store,
	}
}

func (r *ConsumersRepository) Create(ctx context.Context, consumer Consumer) error {
	query := "INSERT INTO consumers (ID, name, apikey, role_id) VALUES ($1, $2, $3, $4)"
	_, err := r.store.Exec(ctx, query, consumer.ID, consumer.Name, consumer.Apikey, consumer.RoleID)
	if err != nil {
		return err
	}
	return nil
}

func (r *ConsumersRepository) FindByID(ctx context.Context, id string) (*Consumer, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidUUID, id)
	}

	query := "SELECT ID, name, apikey, role_id, active FROM consumers WHERE id=$1"

	var consumer Consumer
	err := r.store.QueryRow(ctx, query, id).Scan(&consumer.ID, &consumer.Name, &consumer.Apikey, &consumer.RoleID, &consumer.Active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w consumer with id: %q", ErrorNotFound, id)
		}
		return nil, err
	}

	return &consumer, nil
}

func (r *ConsumersRepository) Update(ctx context.Context, req UpdateConsumerRequest) error {
	if req.ID == "" {
		return fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}
	if _, err := uuid.Parse(req.ID); err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidUUID, req.ID)
	}

	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)

	if req.Name != nil {
		args = append(args, *req.Name)
		sets = append(sets, fmt.Sprintf("name=$%d", len(args)))
	}
	if req.Apikey != nil {
		args = append(args, *req.Apikey)
		sets = append(sets, fmt.Sprintf("apikey=$%d", len(args)))
	}
	if req.RoleID != nil {
		args = append(args, *req.RoleID)
		sets = append(sets, fmt.Sprintf("role_id=$%d", len(args)))
	}
	if req.Active != nil {
		args = append(args, *req.Active)
		sets = append(sets, fmt.Sprintf("active=$%d", len(args)))
	}

	if len(sets) == 0 {
		return nil
	}

	args = append(args, req.ID)
	query := fmt.Sprintf("UPDATE consumers SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))

	tag, err := r.store.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w consumer with id: %q", ErrorNotFound, req.ID)
	}

	return nil
}

func (r *ConsumersRepository) DeleteByID(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: %q", ErrEmptyArgument, "id")
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidUUID, id)
	}

	query := "DELETE FROM consumers WHERE id=$1"

	tag, err := r.store.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w consumer with id: %q", ErrorNotFound, id)
	}

	return nil
}

func (r *ConsumersRepository) FindByApikey(ctx context.Context, apikey string) (*Consumer, error) {
	if apikey == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyArgument, "apikey")
	}

	query := "SELECT ID, name, apikey, role_id, active FROM consumers WHERE apikey=$1"

	var consumer Consumer
	err := r.store.QueryRow(ctx, query, apikey).Scan(&consumer.ID, &consumer.Name, &consumer.Apikey, &consumer.RoleID, &consumer.Active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w consumer with apikey: %q", ErrorNotFound, apikey)
		}
		return nil, err
	}

	return &consumer, nil
}
