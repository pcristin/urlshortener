package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage is test storage
type MockStorage struct {
	urls map[string]string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		urls: make(map[string]string),
	}
}

func (m *MockStorage) AddURL(token, longURL string) error {
	m.urls[token] = longURL
	return nil
}

func (m *MockStorage) GetURL(token string) (string, error) {
	if url, ok := m.urls[token]; ok {
		return url, nil
	}
	return "", nil
}

func TestEncodeURLHandler(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		body        string
		contentType string
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
			wantStatus:  http.StatusBadRequest,
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
			// Initialize mock storage and handler
			storage := NewMockStorage()
			handler := NewHandler(storage)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.EncodeURLHandler(w, req)

			// Check response
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestDecodeURLHandler(t *testing.T) {
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
			name:       "non-existing url",
			method:     http.MethodGet,
			token:      "nonexistent",
			storedURL:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			token:      "abc123",
			storedURL:  "https://google.com",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mock storage and handler
			storage := NewMockStorage()
			if tt.storedURL != "" {
				err := storage.AddURL(tt.token, tt.storedURL)
				require.NoError(t, err)
			}

			handler := NewHandler(storage)

			// Create chi router for URL parameter handling
			r := chi.NewRouter()
			r.Get("/{id}", handler.DecodeURLHandler)

			// Create request
			req := httptest.NewRequest(tt.method, "/"+tt.token, nil)
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Check response
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.storedURL, resp.Header.Get("Location"))
			}
		})
	}
}
