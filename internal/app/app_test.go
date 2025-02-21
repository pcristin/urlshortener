package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailru/easyjson"
	"github.com/pcristin/urlshortener/internal/config"
	"github.com/pcristin/urlshortener/internal/logger"
	mod "github.com/pcristin/urlshortener/internal/models"
	"github.com/pcristin/urlshortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	MemoryStorage   = storage.StorageType(0)
	FileStorage     = storage.StorageType(1)
	DatabaseStorage = storage.StorageType(2)
	testUserID      = "test-user-id"
	testSecret      = "test-secret-key"
)

func setupTestConfig() *config.Options {
	cfg := config.NewOptions()
	os.Setenv("SECRET_URL_SERVICE", testSecret)
	// Don't parse flags in tests to avoid flag redefinition
	cfg.LoadEnvVariables()
	return cfg
}

// MockStorage implements URLStorager interface
type MockStorage struct {
	urls        map[string]mod.URLStorageNode
	filepath    string
	storageType storage.StorageType
	dbPool      *pgxpool.Pool
}

func NewMockStorage(storageType storage.StorageType) *MockStorage {
	return &MockStorage{
		urls:        make(map[string]mod.URLStorageNode),
		storageType: storageType,
	}
}

func (m *MockStorage) AddURL(token, longURL string, userID string) error {
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}
	// Check for duplicate URLs
	for _, node := range m.urls {
		if node.OriginalURL == longURL {
			return storage.ErrURLExists
		}
	}
	m.urls[token] = mod.URLStorageNode{
		UUID:        uuid.New(),
		ShortURL:    token,
		OriginalURL: longURL,
		UserID:      userID,
	}
	return nil
}

func (m *MockStorage) GetURL(token string) (string, error) {
	if node, ok := m.urls[token]; ok {
		return node.OriginalURL, nil
	}
	return "", errors.New("URL not found")
}

func (m *MockStorage) GetTokenByURL(longURL string) (string, error) {
	for _, node := range m.urls {
		if node.OriginalURL == longURL {
			return node.ShortURL, nil
		}
	}
	return "", errors.New("url not found")
}

func (m *MockStorage) GetUserURLs(userID string) ([]mod.URLStorageNode, error) {
	var result []mod.URLStorageNode
	for _, node := range m.urls {
		if node.UserID == userID {
			result = append(result, node)
		}
	}
	return result, nil
}

func (m *MockStorage) SaveToFile() error {
	if m.filepath == "" {
		return nil
	}

	file, err := os.OpenFile(m.filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, node := range m.urls {
		data, err := easyjson.Marshal(&node)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return writer.Flush()
}

func (m *MockStorage) LoadFromFile(filepath string) error {
	m.filepath = filepath

	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var node mod.URLStorageNode
		if err := easyjson.Unmarshal(scanner.Bytes(), &node); err != nil {
			return err
		}
		m.urls[node.ShortURL] = node
	}

	return scanner.Err()
}

func (m *MockStorage) SetDBPool(pool *pgxpool.Pool) {
	m.dbPool = pool
}

func (m *MockStorage) GetStorageType() storage.StorageType {
	return m.storageType
}

func (m *MockStorage) GetDBPool() *pgxpool.Pool {
	if m.storageType == DatabaseStorage {
		if m.dbPool == nil {
			config, _ := pgxpool.ParseConfig("")
			pool, _ := pgxpool.NewWithConfig(context.Background(), config)
			m.dbPool = pool
		}
		return m.dbPool
	}
	return nil
}

func (m *MockStorage) AddURLBatch(urls map[string]string) error {
	if len(urls) == 0 {
		return errors.New("batch cannot be empty")
	}
	for token, longURL := range urls {
		if err := m.AddURL(token, longURL, testUserID); err != nil {
			return err
		}
	}
	return nil
}

func TestEncodeURLHandler(t *testing.T) {
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	cfg := setupTestConfig()

	tests := []struct {
		name        string
		method      string
		url         string
		body        string
		contentType string
		setupFunc   func(*MockStorage)
		wantStatus  int
	}{
		{
			name:        "valid url",
			method:      http.MethodPost,
			url:         "/",
			body:        "https://google.com",
			contentType: "text/plain; charset=utf-8",
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "duplicate url",
			method:      http.MethodPost,
			url:         "/",
			body:        "https://google.com",
			contentType: "text/plain; charset=utf-8",
			setupFunc: func(s *MockStorage) {
				_ = s.AddURL("abc123", "https://google.com", testUserID)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:        "empty url",
			method:      http.MethodPost,
			url:         "/",
			body:        "",
			contentType: "text/plain; charset=utf-8",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid url",
			method:      http.MethodPost,
			url:         "/",
			body:        "not-a-url",
			contentType: "text/plain; charset=utf-8",
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "wrong method",
			method:      http.MethodGet,
			url:         "/",
			body:        "https://google.com",
			contentType: "text/plain; charset=utf-8",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage(storage.MemoryStorageType)
			if tt.setupFunc != nil {
				tt.setupFunc(storage)
			}

			handler := NewHandler(storage, cfg)
			loggedHandler := logger.WithLogging(handler.EncodeURLHandler, log)

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			ctx := setUserIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			loggedHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
			} else if tt.wantStatus == http.StatusConflict {
				assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Contains(t, string(body), "abc123")
			}
		})
	}
}

func TestDecodeURLHandler(t *testing.T) {
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	cfg := setupTestConfig()

	tests := []struct {
		name       string
		method     string
		token      string
		storedURL  string
		wantStatus int
	}{
		{
			name:       "existing url",
			method:     http.MethodGet,
			token:      "abc123",
			storedURL:  "https://google.com",
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:       "non existing url",
			method:     http.MethodGet,
			token:      "nonexistent",
			storedURL:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "testing protocol scheme",
			method:     http.MethodGet,
			token:      "A8gtZk8",
			storedURL:  "www.dzen.ru",
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			token:      "abc123",
			storedURL:  "https://google.com",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage(storage.MemoryStorageType)

			if tt.storedURL != "" {
				err := storage.AddURL(tt.token, tt.storedURL, testUserID)
				require.NoError(t, err, "Failed to populate storage")
			}

			handler := NewHandler(storage, cfg)
			loggedHandler := logger.WithLogging(handler.DecodeURLHandler, log)

			r := chi.NewRouter()
			r.Get("/{id}", loggedHandler)

			req := httptest.NewRequest(tt.method, "/"+tt.token, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusTemporaryRedirect {
				location := resp.Header.Get("Location")
				assert.Equal(t, tt.storedURL, location)
			}
		})
	}
}

func TestApiEncodeHandler(t *testing.T) {
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	cfg := setupTestConfig()

	tests := []struct {
		name       string
		method     string
		url        string
		body       mod.Request
		setupFunc  func(*MockStorage)
		wantStatus int
		wantInBody string
	}{
		{
			name:   "valid url",
			method: http.MethodPost,
			url:    "/api/shorten",
			body: mod.Request{
				URL: "https://google.com",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "duplicate url",
			method: http.MethodPost,
			url:    "/api/shorten",
			body: mod.Request{
				URL: "https://google.com",
			},
			setupFunc: func(s *MockStorage) {
				_ = s.AddURL("abc123", "https://google.com", testUserID)
			},
			wantStatus: http.StatusConflict,
			wantInBody: "abc123",
		},
		{
			name:       "empty url",
			method:     http.MethodPost,
			url:        "/api/shorten",
			body:       mod.Request{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			url:    "/api/shorten",
			body: mod.Request{
				URL: "https://google.com",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage(storage.MemoryStorageType)
			if tt.setupFunc != nil {
				tt.setupFunc(storage)
			}

			handler := NewHandler(storage, cfg)
			loggedHandler := logger.WithLogging(handler.APIEncodeHandler, log)

			bodyBytes, err := easyjson.Marshal(&tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx := setUserIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			loggedHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus == http.StatusCreated || tt.wantStatus == http.StatusConflict {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
				if tt.wantInBody != "" {
					body, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					assert.Contains(t, string(body), tt.wantInBody)
				}
			}
		})
	}
}

func TestAPIEncodeBatchHandler(t *testing.T) {
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	cfg := setupTestConfig()

	tests := []struct {
		name       string
		method     string
		url        string
		body       mod.BatchRequest
		setupFunc  func(*MockStorage)
		wantStatus int
		wantInBody string
	}{
		{
			name:   "valid batch",
			method: http.MethodPost,
			url:    "/api/shorten/batch",
			body: mod.BatchRequest{
				{CorrelationID: "1", OriginalURL: "https://google.com"},
				{CorrelationID: "2", OriginalURL: "https://yandex.ru"},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "batch with duplicates",
			method: http.MethodPost,
			url:    "/api/shorten/batch",
			body: mod.BatchRequest{
				{CorrelationID: "1", OriginalURL: "https://google.com"},
				{CorrelationID: "2", OriginalURL: "https://yandex.ru"},
			},
			setupFunc: func(s *MockStorage) {
				_ = s.AddURL("abc123", "https://google.com", testUserID)
			},
			wantStatus: http.StatusCreated,
			wantInBody: "abc123",
		},
		{
			name:       "empty batch",
			method:     http.MethodPost,
			url:        "/api/shorten/batch",
			body:       mod.BatchRequest{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			url:    "/api/shorten/batch",
			body: mod.BatchRequest{
				{CorrelationID: "1", OriginalURL: "https://google.com"},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage(storage.MemoryStorageType)
			if tt.setupFunc != nil {
				tt.setupFunc(storage)
			}

			handler := NewHandler(storage, cfg)
			loggedHandler := logger.WithLogging(handler.APIEncodeBatchHandler, log)

			bodyBytes, err := easyjson.Marshal(&tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx := setUserIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			loggedHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
				if tt.wantInBody != "" {
					body, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					assert.Contains(t, string(body), tt.wantInBody)
				}
			}
		})
	}
}

func TestGetUserURLsHandler(t *testing.T) {
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	cfg := setupTestConfig()

	tests := []struct {
		name       string
		method     string
		userID     string
		setupFunc  func(*MockStorage)
		wantStatus int
		wantURLs   int
	}{
		{
			name:       "no urls",
			method:     http.MethodGet,
			userID:     "user1",
			wantStatus: http.StatusNoContent,
			wantURLs:   0,
		},
		{
			name:   "has urls",
			method: http.MethodGet,
			userID: "user1",
			setupFunc: func(s *MockStorage) {
				_ = s.AddURL("abc123", "https://google.com", "user1")
				_ = s.AddURL("def456", "https://yandex.ru", "user1")
			},
			wantStatus: http.StatusOK,
			wantURLs:   2,
		},
		{
			name:   "other user urls",
			method: http.MethodGet,
			userID: "user1",
			setupFunc: func(s *MockStorage) {
				_ = s.AddURL("abc123", "https://google.com", "user2")
			},
			wantStatus: http.StatusNoContent,
			wantURLs:   0,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			userID:     "user1",
			wantStatus: http.StatusMethodNotAllowed,
			wantURLs:   0,
		},
		{
			name:       "no auth",
			method:     http.MethodGet,
			wantStatus: http.StatusUnauthorized,
			wantURLs:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage(storage.MemoryStorageType)
			if tt.setupFunc != nil {
				tt.setupFunc(storage)
			}

			handler := NewHandler(storage, cfg)
			loggedHandler := logger.WithLogging(handler.GetUserURLsHandler, log)

			req := httptest.NewRequest(tt.method, "/api/user/urls", nil)
			if tt.userID != "" {
				ctx := setUserIDToContext(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}
			w := httptest.NewRecorder()

			loggedHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusOK {
				var urls []UserURL
				err := json.NewDecoder(resp.Body).Decode(&urls)
				require.NoError(t, err)
				assert.Equal(t, tt.wantURLs, len(urls))
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	cfg := setupTestConfig()

	tests := []struct {
		name           string
		existingCookie bool
		validSignature bool
	}{
		{
			name:           "no cookies",
			existingCookie: false,
		},
		{
			name:           "valid cookies",
			existingCookie: true,
			validSignature: true,
		},
		{
			name:           "invalid signature",
			existingCookie: true,
			validSignature: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage(storage.MemoryStorageType)
			handler := NewHandler(storage, cfg)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := getUserIDFromContext(r.Context())
				assert.NotEmpty(t, userID)
				w.WriteHeader(http.StatusOK)
			})

			wrappedHandler := handler.AuthMiddleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.existingCookie {
				userID := "test-user"
				signature := generateSignature(userID, []byte(testSecret))
				if !tt.validSignature {
					signature = "invalid-signature"
				}
				req.AddCookie(&http.Cookie{Name: userIDCookieName, Value: userID})
				req.AddCookie(&http.Cookie{Name: signatureCookieName, Value: signature})
			}

			w := httptest.NewRecorder()
			wrappedHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			cookies := resp.Cookies()
			if !tt.existingCookie || !tt.validSignature {
				assert.Len(t, cookies, 2)
				var foundUserID, foundSignature bool
				for _, cookie := range cookies {
					if cookie.Name == userIDCookieName {
						foundUserID = true
					}
					if cookie.Name == signatureCookieName {
						foundSignature = true
					}
				}
				assert.True(t, foundUserID)
				assert.True(t, foundSignature)
			} else {
				assert.Empty(t, cookies)
			}
		})
	}
}
