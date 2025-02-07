package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseManagerInterface interface {
	Ping(ctx context.Context) error
	Close()
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
