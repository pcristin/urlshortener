package app

import (
	"io"
	"net/http"
)

// EncodeURLHandler handles requests to shorten a URL.
// It accepts the long URL in the request body as plain text and returns the shortened URL.
// This handler supports HTTP POST requests only.
//
// If the URL already exists in the system, it returns the existing shortened URL with a 409 Conflict status.
// If successful in creating a new shortened URL, it returns the shortened URL with a 201 Created status.
//
// The response is plain text containing the fully qualified shortened URL.
func (h *Handler) EncodeURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	longURL, err := io.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil || len(string(longURL)) == 0 {
		http.Error(res, "bad request: incorrect long URL", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := getUserIDFromContext(req.Context())

	token, alreadyExists, err := h.service.ShortenURL(req.Context(), string(longURL), userID)
	if err != nil {
		http.Error(res, "bad request: unable to shorten provided url", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")

	if alreadyExists {
		res.WriteHeader(http.StatusConflict)
	} else {
		res.WriteHeader(http.StatusCreated)
	}

	resBody := h.constructURL(token, req)
	res.Write([]byte(resBody))
}
