package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Type for database storage
type DatabaseStorage struct {
	dbPool *pgxpool.Pool
}

// Return a new object of URLStorage (database type)
func NewDatabaseStorage(pool *pgxpool.Pool) *DatabaseStorage {
	return &DatabaseStorage{
		dbPool: pool,
	}
}

// Writes a new link of token --> long URL in DB
func (ds *DatabaseStorage) AddURL(token, longURL string) error {
	if ds.dbPool == nil {
		return errors.New("database not initialized")
	}
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ds.dbPool.Exec(ctx,
		"INSERT INTO urls (token, original_url) VALUES ($1, $2)",
		token, longURL)
	return err
}

// Gets a long URL by token from DB
func (ds *DatabaseStorage) GetURL(token string) (string, error) {
	if ds.dbPool == nil {
		return "", errors.New("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var longURL string
	err := ds.dbPool.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE token = $1",
		token).Scan(&longURL)
	if err != nil {
		return "", err
	}
	return longURL, nil
}

// Empty method from URLStorage interface
func (ds *DatabaseStorage) SaveToFile() error {
	return nil
}

// Empty method from URLStorage interface
func (ds *DatabaseStorage) LoadFromFile(filepath string) error {
	return nil
}

// Set pool for making queries to DB
func (ds *DatabaseStorage) SetDBPool(pool *pgxpool.Pool) {
	ds.dbPool = pool
}

// Get URLStorage type (inherited method)
func (ds *DatabaseStorage) GetStorageType() StorageType {
	return DatabaseStorageType
}

// Get pool for making queries to DB
func (ds *DatabaseStorage) GetDBPool() *pgxpool.Pool {
	return ds.dbPool
}

func (ds *DatabaseStorage) AddURLBatch(urls map[string]string) error {
	if ds.dbPool == nil {
		return errors.New("database not initialized")
	}

	if len(urls) == 0 {
		return errors.New("batch cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := ds.dbPool.Begin(ctx)
	if err != nil {
		zap.L().Sugar().Errorw("Begin transaction failed", "error", err)
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && rErr != pgx.ErrTxClosed {
			zap.L().Sugar().Errorw("Rollback failed", "error", rErr)
		}
	}()

	zap.L().Sugar().Infow("Transaction started")

	batch := &pgx.Batch{}
	for token, longURL := range urls {
		zap.L().Sugar().Info("Queueing INSERT in batch", "token", token, "url", longURL)
		batch.Queue("INSERT INTO urls (token, original_url) VALUES ($1, $2)", token, longURL)
	}

	br := tx.SendBatch(ctx, batch)
	defer func() {
		if cErr := br.Close(); cErr != nil {
			zap.L().Sugar().Errorw("Batch close failed", "error", cErr)
		}
	}()

	zap.L().Sugar().Infow("Batch sent to DB", "batch_len", batch.Len())

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			zap.L().Sugar().Errorw("Exec in batch failed", "index", i, "error", err)
			return err
		}
		zap.L().Sugar().Debugw("Exec in batch succeeded", "index", i)
	}

	if err := tx.Commit(ctx); err != nil {
		zap.L().Sugar().Errorw("Commit failed", "error", err)
		return err
	}
	zap.L().Sugar().Infow("Transaction committed successfully")

	return nil
}
