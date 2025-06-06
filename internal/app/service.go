package app

import (
	"context"
	"errors"
	"net"

	"github.com/pcristin/urlshortener/internal/models"
	"github.com/pcristin/urlshortener/internal/storage"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
	"go.uber.org/zap"
)

// URLService contains the business logic for URL shortening operations
type URLService struct {
	storage       storage.URLStorager
	baseURL       string
	trustedSubnet string
	logger        *zap.Logger
}

// NewURLService creates a new instance of URLService
func NewURLService(storage storage.URLStorager, baseURL string, trustedSubnet string) *URLService {
	return &URLService{
		storage:       storage,
		baseURL:       baseURL,
		trustedSubnet: trustedSubnet,
		logger:        zap.L(),
	}
}

// ShortenURL shortens a single URL
func (s *URLService) ShortenURL(ctx context.Context, longURL string, userID string) (string, bool, error) {
	if longURL == "" {
		return "", false, errors.New("url cannot be empty")
	}

	token, err := uu.EncodeURL(longURL, s.storage, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			// URL already exists, return the existing token
			return token, true, nil
		}
		return "", false, err
	}

	return token, false, nil
}

// GetOriginalURL retrieves the original URL by token
func (s *URLService) GetOriginalURL(ctx context.Context, token string) (string, bool, error) {
	if token == "" {
		return "", false, errors.New("token cannot be empty")
	}

	longURL, err := uu.DecodeURL(token, s.storage)
	if err != nil {
		if errors.Is(err, storage.ErrURLDeleted) {
			return "", true, nil
		}
		return "", false, err
	}

	return longURL, false, nil
}

// ShortenBatch shortens multiple URLs in a single operation
func (s *URLService) ShortenBatch(ctx context.Context, items []BatchItem, userID string) ([]BatchResult, error) {
	if len(items) == 0 {
		return nil, errors.New("batch cannot be empty")
	}

	results := make([]BatchResult, 0, len(items))

	for _, item := range items {
		token, err := uu.EncodeURL(item.OriginalURL, s.storage, userID)
		if err != nil && !errors.Is(err, storage.ErrURLExists) {
			s.logger.Error("Error encoding URL", zap.Error(err), zap.String("url", item.OriginalURL))
			return nil, err
		}
		results = append(results, BatchResult{
			CorrelationID: item.CorrelationID,
			ShortURL:      token,
		})
	}

	return results, nil
}

// GetUserURLs retrieves all URLs for a specific user
func (s *URLService) GetUserURLs(ctx context.Context, userID string) ([]UserURL, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	urls, err := s.storage.GetUserURLs(userID)
	if err != nil {
		return nil, err
	}

	userURLs := make([]UserURL, len(urls))
	for i, url := range urls {
		userURLs[i] = UserURL{
			ShortURL:    url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
	}

	return userURLs, nil
}

// DeleteUserURLs marks URLs as deleted for a specific user
func (s *URLService) DeleteUserURLs(ctx context.Context, userID string, tokens []string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	// This operation is async in the original implementation
	go func() {
		if err := s.storage.DeleteURLs(userID, tokens); err != nil {
			s.logger.Error("Error deleting URLs", zap.Error(err))
		}
	}()

	return nil
}

// Ping checks the health of the service
func (s *URLService) Ping(ctx context.Context) (bool, error) {
	// Check if we have database storage
	dbStorage, ok := s.storage.(*storage.DatabaseStorage)
	if !ok || dbStorage.GetDBPool() == nil {
		return false, nil
	}

	// Ping the database
	if err := dbStorage.GetDBPool().Ping(ctx); err != nil {
		return false, nil
	}

	return true, nil
}

// GetStats returns service statistics
func (s *URLService) GetStats(ctx context.Context, clientIP string) (*models.Stats, error) {
	// Check trusted subnet if configured
	if s.trustedSubnet != "" {
		_, trustedNet, err := net.ParseCIDR(s.trustedSubnet)
		if err != nil {
			return nil, errors.New("invalid trusted subnet configuration")
		}

		ip := net.ParseIP(clientIP)
		if ip == nil || !trustedNet.Contains(ip) {
			return nil, errors.New("unauthorized: IP not in trusted subnet")
		}
	}

	stats, err := s.storage.GetStats()
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// BatchItem represents a single item in a batch request
type BatchItem struct {
	CorrelationID string
	OriginalURL   string
}

// BatchResult represents a single result in a batch response
type BatchResult struct {
	CorrelationID string
	ShortURL      string
}
