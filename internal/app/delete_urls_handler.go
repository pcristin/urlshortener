package app

import (
	"encoding/json"
	"io"
	"net/http"
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

	// Delete URLs using service (async operation)
	if err := h.service.DeleteUserURLs(r.Context(), userID, tokens); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return 202 Accepted immediately
	w.WriteHeader(http.StatusAccepted)
}
