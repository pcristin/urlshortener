package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pcristin/urlshortener/internal/models"
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
func (ds *DatabaseStorage) AddURL(token, longURL string, userID string) error {
	if ds.dbPool == nil {
		return errors.New("database not initialized")
	}
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ds.dbPool.Exec(ctx,
		"INSERT INTO urls (token, original_url, user_id) VALUES ($1, $2, $3)",
		token, longURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// Check if it's the original_url unique constraint
			if pgErr.ConstraintName == "idx_urls_original_url" {
				return ErrURLExists
			}
		}
		return err
	}
	return nil
}

// Gets a long URL by token from DB
func (ds *DatabaseStorage) GetURL(token string) (string, error) {
	if ds.dbPool == nil {
		return "", errors.New("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var longURL string
	var isDeleted bool
	err := ds.dbPool.QueryRow(ctx,
		"SELECT original_url, is_deleted FROM urls WHERE token = $1",
		token).Scan(&longURL, &isDeleted)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", errors.New("URL not found")
		}
		return "", err
	}

	if isDeleted {
		return "", ErrURLDeleted
	}

	return longURL, nil
}

// Gets a token by original URL from DB
func (ds *DatabaseStorage) GetTokenByURL(longURL string) (string, error) {
	if ds.dbPool == nil {
		return "", errors.New("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var token string
	err := ds.dbPool.QueryRow(ctx,
		"SELECT token FROM urls WHERE original_url = $1",
		longURL).Scan(&token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", errors.New("url not found")
		}
		return "", err
	}
	return token, nil
}

// GetUserURLs returns all URLs shortened by a specific user
func (ds *DatabaseStorage) GetUserURLs(userID string) ([]models.URLStorageNode, error) {
	if ds.dbPool == nil {
		return nil, errors.New("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := ds.dbPool.Query(ctx,
		"SELECT id, token, original_url FROM urls WHERE user_id = $1",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []models.URLStorageNode
	for rows.Next() {
		var node models.URLStorageNode
		var id string
		err := rows.Scan(&id, &node.ShortURL, &node.OriginalURL)
		if err != nil {
			return nil, err
		}
		node.UUID, err = uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		node.UserID = userID
		urls = append(urls, node)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
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

// AddURLBatch adds multiple URLs to the database in a single transaction
func (ds *DatabaseStorage) AddURLBatch(urls map[string]string) error {
	if ds.dbPool == nil {
		return errors.New("database not initialized")
	}
	if len(urls) == 0 {
		return errors.New("batch cannot be empty")
	}

	ctx := context.Background()
	tx, err := ds.dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && rErr != pgx.ErrTxClosed {
			zap.L().Sugar().Errorw("Rollback failed", "error", rErr)
		}
	}()

	batch := &pgx.Batch{}
	for token, originalURL := range urls {
		zap.L().Sugar().Infof("Queueing INSERT for token=%s, original_url=%s", token, originalURL)
		batch.Queue(`
			INSERT INTO urls (token, original_url, user_id) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (original_url) DO NOTHING`,
			token, originalURL, uuid.New().String()) // Generate a new UUID for each URL in batch
	}

	br := tx.SendBatch(ctx, batch)
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				// If it's a unique violation on the token, we should return an error
				if pgErr.ConstraintName == "urls_token_key" {
					zap.L().Sugar().Errorf("batch execution error at item %d: %v", i, err)
					_ = br.Close()
					return err
				}
				// If it's a unique violation on the original_url, we can ignore it
				continue
			}
			zap.L().Sugar().Errorf("batch execution error at item %d: %v", i, err)
			_ = br.Close()
			return err
		}
	}

	if err := br.Close(); err != nil {
		zap.L().Sugar().Errorw("Failed to close batch", "error", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		zap.L().Sugar().Errorf("commit failed: %v", err)
		return err
	}

	return nil
}

// DeleteURLs marks multiple URLs as deleted for a specific user
func (ds *DatabaseStorage) DeleteURLs(userID string, tokens []string) error {
	if ds.dbPool == nil {
		return errors.New("database not initialized")
	}
	if len(tokens) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start a transaction
	tx, err := ds.dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && rErr != pgx.ErrTxClosed {
			zap.L().Sugar().Errorw("Rollback failed", "error", rErr)
		}
	}()

	// Create a batch for updating multiple URLs
	batch := &pgx.Batch{}
	for _, token := range tokens {
		batch.Queue(`
			UPDATE urls 
			SET is_deleted = TRUE 
			WHERE token = $1 AND user_id = $2`,
			token, userID)
	}

	// Execute the batch
	br := tx.SendBatch(ctx, batch)
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			zap.L().Sugar().Errorf("batch execution error at item %d: %v", i, err)
			_ = br.Close()
			return err
		}
	}

	if err := br.Close(); err != nil {
		zap.L().Sugar().Errorw("Failed to close batch", "error", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		zap.L().Sugar().Errorf("commit failed: %v", err)
		return err
	}

	return nil
}
