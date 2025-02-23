package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseManagerInterface interface {
	Ping(ctx context.Context) error
	Close()
	GetPool() *pgxpool.Pool
}

type DatabaseManager struct {
	pool *pgxpool.Pool
}

// NewDatabaseManager creates a new database manager instance
func NewDatabaseManager(databaseDSN string) (DatabaseManagerInterface, error) {
	pool, err := pgxpool.New(context.Background(), databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	// Create the database table if it doesn't exist
	ctx := context.Background()
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			token VARCHAR(10) NOT NULL UNIQUE,
			original_url TEXT NOT NULL,
			user_id TEXT NOT NULL,
			is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original_url ON urls (original_url);
		CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls (user_id);
	`)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to create table: %v", err)
	}

	return &DatabaseManager{
		pool: pool,
	}, nil
}

// Ping checks database connectivity
func (dm *DatabaseManager) Ping(ctx context.Context) error {
	return dm.pool.Ping(ctx)
}

// Close closes the database connection pool
func (dm *DatabaseManager) Close() {
	if dm.pool != nil {
		dm.pool.Close()
	}
}

// GetPool gets pool from DatabaseManager
func (dm *DatabaseManager) GetPool() *pgxpool.Pool {
	return dm.pool
}
