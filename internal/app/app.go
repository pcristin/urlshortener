package app

import (
	"errors"
	"io"
	randMath "math/rand/v2"
	"net/http"
)

var urlStorage = make(map[string]string)

func EncodeURLHandler(res http.ResponseWriter, req *http.Request) {
	longUrl, err := io.ReadAll(req.Body)

	if req.Method != http.MethodPost || req.Host != "localhost:8080" || err != nil || req.Header.Get("Content-Type") != "text/plain; charset=utf-8" || len(longUrl) == 0 {
		http.Error(res, "Bad request!", http.StatusBadRequest)
		return
	}

	token := encodeURL(string(longUrl))
	res.Header().Set("content-type", "text/plain")
	res.Header()["Date"] = nil
	res.WriteHeader(http.StatusCreated)
	resBody := "https://" + req.Host + "/" + string(token)
	res.Write([]byte(resBody))
}

func DecodeURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet || req.Host != "localhost:8080" || req.Header.Get("Content-Type") != "text/plain" {
		http.Error(res, "Bad request!", http.StatusBadRequest)
		return
	}
	token := req.PathValue("id")
	if token == "" {
		http.Error(res, "Bad request!", http.StatusBadRequest)
		return
	}
	longUrl, err := decodeURL(token)
	if err != nil {
		http.Error(res, "Bad request!", http.StatusBadRequest)
		return
	}
	res.Header().Add("Location", longUrl)
	res.Header()["Date"] = nil
	res.Header()["Content-Length"] = nil
	res.Header()["Transfer-Encoding"] = nil
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func decodeURL(token string) (string, error) {
	if url, found := urlStorage[token]; found {
		return url, nil
	} else {
		return "", errors.New("haven't found the URL")
	}
}

func randRange(a int, b int) int {
	return randMath.IntN(b-a) + a
}

func generateToken(length int) string {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	token := make([]byte, length)
	for i := range token {
		token[i] = letters[randMath.IntN(len(letters))]
	}
	return string(token)
}

func encodeURL(url string) string {
	// Creating the random length (from 6 to including 9) slice of bytes
	length := randRange(6, 10)
	token := generateToken(length)
	urlStorage[string(token)] = url
	return string(token)
}
