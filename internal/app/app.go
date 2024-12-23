package app

import (
	"io"
	"net/http"

	uu "github.com/pcristin/urlshortener/internal/urlutils"
)

func EncodeURLHandler(res http.ResponseWriter, req *http.Request) {
	longURL, err := io.ReadAll(req.Body)

	if req.Method != http.MethodPost || req.Host != "localhost:8080" || err != nil || req.Header.Get("Content-Type") != "text/plain; charset=utf-8" || len(longURL) == 0 {
		http.Error(res, "Bad request!", http.StatusBadRequest)
		return
	}

	token := uu.EncodeURL(string(longURL))
	res.Header().Set("content-type", "text/plain")
	res.Header()["Date"] = nil
	res.WriteHeader(http.StatusCreated)
	resBody := "http://" + req.Host + "/" + string(token)
	res.Write([]byte(resBody))
}

func DecodeURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet || req.Host != "localhost:8080" {
		http.Error(res, "Bad request!", http.StatusBadRequest)
		return
	}
	token := req.PathValue("id")
	if token == "" {
		http.Error(res, "Bad request!", http.StatusBadRequest)
		return
	}
	longURL, err := uu.DecodeURL(token)
	if err != nil {
		http.Error(res, "Bad request!", http.StatusBadRequest)
		return
	}
	res.Header().Add("Location", longURL)
	res.Header()["Date"] = nil
	res.Header()["Content-Length"] = nil
	res.Header()["Transfer-Encoding"] = nil
	res.WriteHeader(http.StatusTemporaryRedirect)
}
