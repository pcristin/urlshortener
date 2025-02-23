package app

import (
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// DeleteUserURLsHandler handles DELETE /api/user/urls requests
func (h *Handler) DeleteUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var tokens []string
	if err := json.Unmarshal(body, &tokens); err != nil {
		http.Error(w, "Invalid request body format", http.StatusBadRequest)
		return
	}

	// Asynchronously delete URLs
	go func() {
		if err := h.storage.DeleteURLs(userID, tokens); err != nil {
			// Log error but don't return it to client as per requirements
			h.logger.Error("Error deleting URLs", zap.Error(err))
		}
	}()

	// Return 202 Accepted immediately
	w.WriteHeader(http.StatusAccepted)
}
