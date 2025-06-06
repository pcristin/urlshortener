package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DecodeURLHandler handles requests to redirect from a shortened URL to the original URL.
// It extracts the token from the URL path parameter, looks up the original URL,
// and redirects the client with a 307 Temporary Redirect status.
//
// This handler only supports HTTP GET requests.
//
// If the URL is found but has been marked as deleted, it returns a 410 Gone status.
// If the token is not found or invalid, it returns a 400 Bad Request status.
func (h *Handler) DecodeURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := chi.URLParam(req, "id")
	if token == "" {
		http.Error(res, "bad request: incorrect token", http.StatusBadRequest)
		return
	}

	longURL, isDeleted, err := h.service.GetOriginalURL(req.Context(), token)
	if err != nil {
		http.Error(res, "bad request: unable to decode provided token", http.StatusBadRequest)
		return
	}

	if isDeleted {
		http.Error(res, "URL was deleted", http.StatusGone)
		return
	}

	res.Header().Set("Location", longURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
