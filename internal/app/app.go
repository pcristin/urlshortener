package app

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pcristin/urlshortener/internal/storage"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
)

type HandlerInterface interface {
	EncodeURLHandler(http.ResponseWriter, *http.Request)
	DecodeURLHandler(http.ResponseWriter, *http.Request)
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
	if req.Method != http.MethodPost || req.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	longURL, err := io.ReadAll(req.Body)
	log.Printf("Received Long URL: %s", longURL)
	req.Body = io.NopCloser(bytes.NewBuffer(longURL))
	defer req.Body.Close()

	if err != nil || len(longURL) == 0 || !uu.URLCheck(string(longURL)) {
		http.Error(res, "bad request: incorrect long URL", http.StatusBadRequest)
		return
	}

	token, err := uu.EncodeURL(string(longURL), h.storage)
	if err != nil {
		http.Error(res, "bad request", http.StatusBadRequest)
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

	res.Header().Add("Location", longURL)
	res.Header()["Date"] = nil
	res.Header()["Content-Length"] = nil
	res.WriteHeader(http.StatusTemporaryRedirect)
}
