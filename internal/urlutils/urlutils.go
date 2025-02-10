package urlutils

import (
	"errors"
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
func EncodeURL(url string, s storage.URLStorager) (string, error) {
	length := generateRandomNumber(6, 10)
	token := generateToken(length)
	err := s.AddURL(token, url)
	if err != nil {
		// If the error indicates that the URL already exists, return the existing token
		if errors.Is(err, storage.ErrURLExists) {
			return s.GetTokenByURL(url)
		}
		return "", err
	}
	return token, nil
}

// Check the validity of provided URL address
func URLCheck(url string) bool {
	var regExpURLPattern = regexp.MustCompile(`^((http|https):\/\/)?([a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})+)(\/[a-zA-Z0-9-._~:?#@!$&'()*+,;=]*)?$`)
	return regExpURLPattern.MatchString(url)
}

func GenerateToken() string {
	length := generateRandomNumber(6, 10)
	return generateToken(length)
}
