// Package tests
package tests

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

func Connect(t *testing.T) (*pgx.Conn, error) {
	t.Helper()

	ctx := context.Background()
	dbURL := os.Getenv("POSTGRES_URL")
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
