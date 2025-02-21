package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
)

// Handler to decode encoded long URL
func (h *Handler) DecodeURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	token := chi.URLParam(req, "id")
	if token == "" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	longURL, err := uu.DecodeURL(token, h.storage)
	if err != nil || longURL == "" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", longURL)
	res.Header().Del("Date")
	res.Header().Del("Content-Type")
	res.WriteHeader(http.StatusTemporaryRedirect)
}
