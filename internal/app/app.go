package app

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
)

func EncodeURLHandler(res http.ResponseWriter, req *http.Request) {
	longURL, err := io.ReadAll(req.Body)
	defer req.Body.Close()

	if req.Method != http.MethodPost || err != nil || req.Header.Get("Content-Type") != "text/plain; charset=utf-8" || !uu.URLCheck(string(longURL)) {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	fmt.Printf("Encoding: Provided long URL: %s\r\n", longURL)

	token := uu.EncodeURL(string(longURL))

	fmt.Printf("Encoding: Generated token %s for %s\r\n", string(token), longURL)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	resBody := "http://" + req.Host + "/" + string(token)
	res.Write([]byte(resBody))
}

func DecodeURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	token := chi.URLParam(req, "id")
	if token == "" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	log.Printf("Decoding: Extracted token %s from GET request\r\n", token)

	longURL, err := uu.DecodeURL(token)
	if err != nil {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", longURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
