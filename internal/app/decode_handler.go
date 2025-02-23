package app

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pcristin/urlshortener/internal/storage"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
)

// Handler to decode URL with plain text and without compressing the data
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

	longURL, err := uu.DecodeURL(token, h.storage)
	if err != nil {
		if errors.Is(err, storage.ErrURLDeleted) {
			http.Error(res, "URL was deleted", http.StatusGone)
			return
		}
		http.Error(res, "bad request: unable to decode provided token", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", longURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
