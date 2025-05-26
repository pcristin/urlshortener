package models

import "github.com/google/uuid"

// URLStorageNode represents a URL entry stored in the system.
// It contains information about the original and shortened URLs,
// the user who created the shortened URL, and the deletion status.
type URLStorageNode struct {
	UUID        uuid.UUID `json:"uuid"`         // Unique identifier for the URL node
	ShortURL    string    `json:"short_url"`    // The shortened URL or token
	OriginalURL string    `json:"original_url"` // The original, full-length URL
	UserID      string    `json:"user_id"`      // ID of the user who created this URL
	IsDeleted   bool      `json:"is_deleted"`   // Whether this URL has been marked as deleted
}

// Stats represents the statistics of the URL shortener service.
// It contains the number of URLs and users.
type Stats struct {
	URLs  int `json:"urls"`  // Total number of URLs in the system
	Users int `json:"users"` // Total number of users in the systemcommand not found: easyjson
}
