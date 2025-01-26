package app

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	mod "github.com/pcristin/urlshortener/internal/models"
	"github.com/pcristin/urlshortener/internal/storage"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
)

type HandlerInterface interface {
	EncodeURLHandler(http.ResponseWriter, *http.Request)
	DecodeURLHandler(http.ResponseWriter, *http.Request)
	APIEncodeHandler(http.ResponseWriter, *http.Request)
}

type Handler struct {
	storage storage.URLStorager
}

func NewHandler(storage storage.URLStorager) HandlerInterface {
	return &Handler{
		storage: storage,
	}
}

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

	token, err := uu.EncodeURL(string(longURL), h.storage)
	if err != nil {
		http.Error(res, "bad request: unable to shorten provided url", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	resBody := "http://" + req.Host + "/" + token
	res.Write([]byte(resBody))
}

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

	// Encode the long URL to a short URL
	shortURL, err := uu.EncodeURL(body.URL, h.storage)
	if err != nil {
		http.Error(res, "bad request: unable to shorten provided url", http.StatusBadRequest)
		return
	}

	// Prepare the response payload
	response := mod.Response{
		Result: "http://" + req.Host + "/" + shortURL,
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	responseBytes, err := easyjson.Marshal(response)

	if err != nil {
		http.Error(res, "internal server error: unable to marshal response", http.StatusInternalServerError)
	}
	res.Write(responseBytes)
}
