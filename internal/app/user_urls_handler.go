package app

import (
	"encoding/json"
	"net/http"
)

// UserURL represents a shortened URL with its original URL for API responses
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// GetUserURLsHandler handles GET /api/user/urls requests
func (h *Handler) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's URLs from storage
	urls, err := h.storage.GetUserURLs(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If no URLs found, return 204 No Content
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Convert storage nodes to response format
	response := make([]UserURL, len(urls))
	for i, url := range urls {
		response[i] = UserURL{
			ShortURL:    h.constructURL(url.ShortURL, r),
			OriginalURL: url.OriginalURL,
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
