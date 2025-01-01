package urlutils

import (
	randMath "math/rand/v2"
	"regexp"

	"github.com/pcristin/urlshortener/internal/storage"
)

func DecodeURL(token string, storage storage.URLStorager) (string, error) {
	return storage.GetURL(token)
}

func generateRandomNumber(a int, b int) int {
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

// Encode URL to a range number from 6 to 9 of random characters
func EncodeURL(url string, storage storage.URLStorager) (string, error) {
	length := generateRandomNumber(6, 10)
	token := generateToken(length)
	err := storage.AddURL(token, url)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Check the validity of provided URL address
func URLCheck(url string) bool {
	var regExpURLPattern = regexp.MustCompile(`^((http|https):\/\/)?([a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})+)(\/[a-zA-Z0-9-._~:?#@!$&'()*+,;=]*)?$`)
	return regExpURLPattern.MatchString(url)
}
