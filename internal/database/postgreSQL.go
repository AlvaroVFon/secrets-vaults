package database

import (
	"context"

	"secrets-vault/internal/config"

	"github.com/jackc/pgx/v5"
)

func Connect(cfg *config.DatabaseConfig) (*pgx.Conn, error) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.URL)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
