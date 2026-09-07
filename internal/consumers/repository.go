package consumers

import (
	"context"
	"errors"
	"fmt"

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
