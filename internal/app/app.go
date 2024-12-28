package app

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
)

func EncodeURLHandler(res http.ResponseWriter, req *http.Request) {
	longURL, err := io.ReadAll(req.Body)

	if req.Method != http.MethodPost || err != nil || !uu.URLCheck(string(longURL)) {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	log.Printf("Encoding: Provided long URL: %s\r\n", longURL)

	defer req.Body.Close()

	token := uu.EncodeURL(string(longURL))

	log.Printf("Encoding: Generated token %s for %s\r\n", string(token), longURL)

	res.Header().Set("content-type", "text/plain")
	res.Header()["Date"] = nil
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
	res.Header().Add("Location", longURL)
	res.Header()["Date"] = nil
	res.Header()["Content-Length"] = nil
	res.WriteHeader(http.StatusTemporaryRedirect)
}
