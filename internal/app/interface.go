package app

import "net/http"

// HandlerInterface defines the contract for all HTTP handlers in the URL shortener service.
// It provides methods for encoding and decoding URLs, handling API requests, and middleware functionality.
type HandlerInterface interface {
	// EncodeURLHandler handles requests to shorten a URL, accepting the URL in plain text format
	EncodeURLHandler(http.ResponseWriter, *http.Request)

	// DecodeURLHandler handles requests to redirect to the original URL using a token
	DecodeURLHandler(http.ResponseWriter, *http.Request)

	// APIEncodeHandler handles requests to shorten a URL through the JSON API
	APIEncodeHandler(http.ResponseWriter, *http.Request)

	// APIEncodeBatchHandler handles requests to shorten multiple URLs in a single batch operation
	APIEncodeBatchHandler(http.ResponseWriter, *http.Request)

	// PingHandler checks the database connection health
	PingHandler(http.ResponseWriter, *http.Request)

	// GetUserURLsHandler returns all URLs shortened by a specific user
	GetUserURLsHandler(http.ResponseWriter, *http.Request)

	// DeleteUserURLsHandler marks user's URLs as deleted
	DeleteUserURLsHandler(http.ResponseWriter, *http.Request)

	// AuthMiddleware provides authentication and user identification functionality
	AuthMiddleware(http.HandlerFunc) http.HandlerFunc
}
