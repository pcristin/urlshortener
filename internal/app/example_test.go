package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/pcristin/urlshortener/internal/config"
	"github.com/pcristin/urlshortener/internal/models"
	"github.com/pcristin/urlshortener/internal/storage"
)

func Example_setupTestServer() {
	// This example shows how to set up a test server for the URL shortener service

	// Create a new router
	r := chi.NewRouter()

	// Initialize storage
	urlStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)

	// Initialize config
	cfg := config.NewOptions()

	// Initialize handler
	handler := NewHandler(urlStorage, cfg)

	// Set up routes
	r.Post("/", handler.EncodeURLHandler)
	r.Get("/{id}", handler.DecodeURLHandler)
	r.Post("/api/shorten", handler.APIEncodeHandler)
	r.Post("/api/shorten/batch", handler.APIEncodeBatchHandler)
	r.Get("/ping", handler.PingHandler)
	r.Get("/api/user/urls", handler.GetUserURLsHandler)
	r.Delete("/api/user/urls", handler.DeleteUserURLsHandler)

	// Create a test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	fmt.Printf("Test server running at: %s\n", ts.URL)

	// Example output:
	// Test server running at: http://127.0.0.1:xxxxx
}

func ExampleHandlerInterface_EncodeURLHandler() {
	// This example demonstrates how to shorten a URL using the plain text endpoint

	// Set up a test server with a mock storage
	mockStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)
	cfg := config.NewOptions()
	handler := NewHandler(mockStorage, cfg)

	// Create a test request
	longURL := "https://github.com/pcristin/urlshortener"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(longURL))

	// Create a recorder to capture the response
	w := httptest.NewRecorder()

	// Set a user ID in the context
	ctx := context.WithValue(req.Context(), userIDContextKey, "test-user")
	req = req.WithContext(ctx)

	// Call the handler
	handler.EncodeURLHandler(w, req)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	// Read the response body
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Response contains token: %t\n", len(body) > 0)

	// Example output:
	// Status: 201
	// Content-Type: text/plain
	// Response contains token: true
}

func ExampleHandlerInterface_APIEncodeHandler() {
	// This example demonstrates how to shorten a URL using the JSON API

	// Set up a test server with a mock storage
	mockStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)
	cfg := config.NewOptions()
	handler := NewHandler(mockStorage, cfg)

	// Create the request body
	reqBody := models.Request{
		URL: "https://github.com/pcristin/urlshortener",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a recorder to capture the response
	w := httptest.NewRecorder()

	// Set a user ID in the context
	ctx := context.WithValue(req.Context(), userIDContextKey, "test-user")
	req = req.WithContext(ctx)

	// Call the handler
	handler.APIEncodeHandler(w, req)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	// Read and parse the response
	body, _ := io.ReadAll(resp.Body)
	var response models.Response
	json.Unmarshal(body, &response)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Response contains result URL: %t\n", len(response.Result) > 0)

	// Example output:
	// Status: 201
	// Content-Type: application/json
	// Response contains result URL: true
}

func ExampleHandlerInterface_APIEncodeBatchHandler() {
	// This example demonstrates how to shorten multiple URLs in a batch

	// Set up a test server with a mock storage
	mockStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)
	cfg := config.NewOptions()
	handler := NewHandler(mockStorage, cfg)

	// Create the batch request body
	batchReq := models.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://github.com/pcristin/urlshortener"},
		{CorrelationID: "2", OriginalURL: "https://golang.org"},
	}
	jsonBody, _ := json.Marshal(batchReq)

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a recorder to capture the response
	w := httptest.NewRecorder()

	// Set a user ID in the context
	ctx := context.WithValue(req.Context(), userIDContextKey, "test-user")
	req = req.WithContext(ctx)

	// Call the handler
	handler.APIEncodeBatchHandler(w, req)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	// Read and parse the response
	body, _ := io.ReadAll(resp.Body)
	var response models.BatchResponse
	json.Unmarshal(body, &response)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Response items: %d\n", len(response))

	// Example output:
	// Status: 201
	// Content-Type: application/json
	// Response items: 2
}

func ExampleHandlerInterface_DecodeURLHandler() {
	// This example demonstrates how to use the redirect endpoint

	// Set up a mock storage and add a URL
	mockStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)
	token := "abc123"
	longURL := "https://github.com/pcristin/urlshortener"
	_ = mockStorage.AddURL(token, longURL, "test-user")

	// Set up the handler
	cfg := config.NewOptions()
	handler := NewHandler(mockStorage, cfg)

	// Set up the router to handle URL parameters
	r := chi.NewRouter()
	r.Get("/{id}", handler.DecodeURLHandler)

	// Create a test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Make a request to the test server
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, _ := client.Get(ts.URL + "/" + token)
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Location header: %s\n", resp.Header.Get("Location"))

	// Example output:
	// Status: 307
	// Location header: https://github.com/pcristin/urlshortener
}

func ExampleHandlerInterface_GetUserURLsHandler() {
	// This example demonstrates how to get all URLs shortened by a user

	// Set up a mock storage and add some URLs
	mockStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)
	userID := "test-user"
	_ = mockStorage.AddURL("abc123", "https://github.com/pcristin/urlshortener", userID)
	_ = mockStorage.AddURL("def456", "https://golang.org", userID)

	// Set up the handler
	cfg := config.NewOptions()
	handler := NewHandler(mockStorage, cfg)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	// Create a recorder to capture the response
	w := httptest.NewRecorder()

	// Set a user ID in the context
	ctx := context.WithValue(req.Context(), userIDContextKey, userID)
	req = req.WithContext(ctx)

	// Call the handler
	handler.GetUserURLsHandler(w, req)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	// Read and parse the response
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Response body length: %d\n", len(body))

	// Example output:
	// Status: 200
	// Content-Type: application/json
	// Response body length: >0
}

func ExampleHandlerInterface_DeleteUserURLsHandler() {
	// This example demonstrates how to delete URLs

	// Set up a mock storage and add some URLs
	mockStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)
	userID := "test-user"
	_ = mockStorage.AddURL("abc123", "https://github.com/pcristin/urlshortener", userID)
	_ = mockStorage.AddURL("def456", "https://golang.org", userID)

	// Set up the handler
	cfg := config.NewOptions()
	handler := NewHandler(mockStorage, cfg)

	// Create the request body (list of tokens to delete)
	tokens := []string{"abc123", "def456"}
	jsonBody, _ := json.Marshal(tokens)

	// Create a test request
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a recorder to capture the response
	w := httptest.NewRecorder()

	// Set a user ID in the context
	ctx := context.WithValue(req.Context(), userIDContextKey, userID)
	req = req.WithContext(ctx)

	// Call the handler
	handler.DeleteUserURLsHandler(w, req)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Verify the URLs are marked as deleted
	_, err1 := mockStorage.GetURL("abc123")
	_, err2 := mockStorage.GetURL("def456")
	fmt.Printf("First URL deleted: %t\n", err1 != nil)
	fmt.Printf("Second URL deleted: %t\n", err2 != nil)

	// Example output:
	// Status: 202
	// First URL deleted: true
	// Second URL deleted: true
}

func ExampleHandlerInterface_PingHandler() {
	// This example demonstrates how to check database connectivity

	// Set up a mock storage
	mockStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)

	// Set up the handler
	cfg := config.NewOptions()
	handler := NewHandler(mockStorage, cfg)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	// Create a recorder to capture the response
	w := httptest.NewRecorder()

	// Call the handler
	handler.PingHandler(w, req)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Example output:
	// Status: 200
}
