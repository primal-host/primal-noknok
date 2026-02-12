package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// Open creates a connection pool and bootstraps the schema.
func Open(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	db := &DB{Pool: pool}
	if err := db.bootstrap(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	return db, nil
}

// Close shuts down the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// SeedOwner ensures the owner user exists with the given DID.
func (db *DB) SeedOwner(ctx context.Context, did string) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO users (did, handle, role)
		VALUES ($1, '', 'owner')
		ON CONFLICT (did) DO UPDATE SET role = 'owner', updated_at = now()
	`, did)
	return err
}

func (db *DB) bootstrap(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx, schema)
	return err
}
