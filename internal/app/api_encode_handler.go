package app

import (
	"errors"
	"net/http"

	"github.com/mailru/easyjson"
	mod "github.com/pcristin/urlshortener/internal/models"
	"github.com/pcristin/urlshortener/internal/storage"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
)

// Handler to encode the url with compressed data
func (h *Handler) APIEncodeHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost || req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	var body mod.Request
	err := easyjson.UnmarshalFromReader(req.Body, &body)
	defer req.Body.Close()

	if err != nil || len(body.URL) == 0 {
		http.Error(res, "bad request: incorrect url", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := getUserIDFromContext(req.Context())

	// Encode the long URL to a short URL
	shortURL, err := uu.EncodeURL(body.URL, h.storage, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			response := mod.Response{
				Result: h.constructURL(shortURL, req),
			}
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)
			responseBytes, err := easyjson.Marshal(response)
			if err != nil {
				http.Error(res, "internal server error: unable to marshal response", http.StatusInternalServerError)
				return
			}
			res.Write(responseBytes)
			return
		}
		http.Error(res, "bad request: unable to shorten provided url", http.StatusBadRequest)
		return
	}

	// Prepare the response payload
	response := mod.Response{
		Result: h.constructURL(shortURL, req),
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	responseBytes, err := easyjson.Marshal(response)
	if err != nil {
		http.Error(res, "internal server error: unable to marshal response", http.StatusInternalServerError)
		return
	}
	res.Write(responseBytes)
}
