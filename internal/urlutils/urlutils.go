package urlutils

import (
	"errors"
	randMath "math/rand/v2"
)

var urlStorage = make(map[string]string)

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
