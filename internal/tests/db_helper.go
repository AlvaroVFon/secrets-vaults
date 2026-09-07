// Package tests
package tests

import (
	"context"
	"fmt"
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

func CleanDB(t *testing.T, db *pgx.Conn) error {
	t.Helper()

	ctx := context.Background()

	rows, err := db.Query(ctx, `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = $1 AND tablename <> 'schema_migrations'
	`, "public")
	if err != nil {
		return err
	}

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		tables = append(tables, name)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(tables) == 0 {
		return nil
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, name := range tables {
		if _, err := tx.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE "%s" RESTART IDENTITY CASCADE`, name)); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
