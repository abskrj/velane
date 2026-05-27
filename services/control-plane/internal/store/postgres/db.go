package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_initial.sql
var migrationSQL string

// Store wraps a pgxpool.Pool and provides all database operations.
type Store struct {
	pool *pgxpool.Pool
}

// New connects to Postgres using the provided DSN, runs the embedded migration
// SQL to ensure the schema is up to date, and returns a ready Store.
func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	if _, err := pool.Exec(ctx, migrationSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close releases all pool connections.
func (s *Store) Close() {
	s.pool.Close()
}
